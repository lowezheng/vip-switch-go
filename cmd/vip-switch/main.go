// Copyright 2026 lowezheng
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"vip-switch-go/internal/config"
	"vip-switch-go/internal/hook"
	"vip-switch-go/internal/raft"
	"vip-switch-go/internal/state"
)

var (
	configFile string
	nodeID     string
	raftAddr   string
	dataDir    string
	logLevel   string
	logFormat  string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "vip-switch",
		Short: "A Raft-based VIP failover solution with event-driven hooks",
		Long: `VIP-Switch is a distributed virtual IP failover system that uses Raft consensus
to elect a master node. The master node binds a VIP (Virtual IP) through configurable
hook scripts, while slave nodes ensure the VIP is unbound.

Key features:
- Raft-based leader election for high availability
- Event-driven hook system for flexible VIP management
- Automatic failover and recovery
- Support for cluster membership changes
- Configurable failure strategies and timeouts`,
		Run: run,
	}

	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file (required)")
	rootCmd.Flags().StringVar(&nodeID, "node-id", "", "Node ID (overrides config file)")
	rootCmd.Flags().StringVar(&raftAddr, "raft-addr", "", "Raft RPC address (overrides config file)")
	rootCmd.Flags().StringVar(&dataDir, "data-dir", "", "Raft data directory (overrides config file)")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "", "Log level: debug, info, warn, error (overrides config file)")
	rootCmd.Flags().StringVar(&logFormat, "log-format", "", "Log format: json, text (overrides config file)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	// Validate required flags
	if configFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --config is required")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Override config with CLI flags
	if nodeID != "" {
		cfg.Node.ID = nodeID
	}
	if raftAddr != "" {
		cfg.Node.RaftAddr = raftAddr
	}
	if dataDir != "" {
		cfg.Node.DataDir = dataDir
	}
	if logLevel != "" {
		cfg.Logging.Level = logLevel
	}
	if logFormat != "" {
		cfg.Logging.Format = logFormat
	}

	// Initialize logger
	logger := initLogger(cfg.Logging)

	logger.Info("Starting VIP-Switch",
		"node_id", cfg.Node.ID,
		"raft_addr", cfg.Node.RaftAddr,
		"data_dir", cfg.Node.DataDir,
		"config_file", configFile,
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize hook system
	hookSystem := hook.NewSystem(cfg, logger)

	stateMachine := state.NewMachine(hookSystem, cfg.Node.ID, logger)

	fsm := raft.NewFSM(logger)
	raftNode, err := raft.NewNode(cfg, fsm, logger)
	if err != nil {
		logger.Error("Failed to initialize Raft node", "error", err)
		os.Exit(1)
	}

	stateMachine.SetRaftNode(raftNode)

	if err := raftNode.Start(); err != nil {
		logger.Error("Failed to start Raft node", "error", err)
		os.Exit(1)
	}
	defer raftNode.Shutdown()

	if err := stateMachine.Start(ctx); err != nil {
		logger.Error("Failed to start state machine", "error", err)
		os.Exit(1)
	}

	logger.Info("Executing ToReady hook")
	if err := hookSystem.ExecuteHook(ctx, "ToReady"); err != nil {
		logger.Error("ToReady hook failed", "error", err)
	} else {
		logger.Info("ToReady hook completed successfully")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received signal, shutting down", "signal", sig.String())
		cancel()
	}()

	<-ctx.Done()

	logger.Info("Executing ToDestroy hook")
	if err := hookSystem.ExecuteHook(ctx, "ToDestroy"); err != nil {
		logger.Error("ToDestroy hook failed", "error", err)
	} else {
		logger.Info("ToDestroy hook completed successfully")
	}

	logger.Info("VIP-Switch shutdown complete")
}

func initLogger(cfg config.LoggingConfig) *slog.Logger {
	var logLevel slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var output io.Writer = os.Stdout
	if cfg.Output != "" {
		logDir := filepath.Dir(cfg.Output)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
			os.Exit(1)
		}

		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
			os.Exit(1)
		}
		output = file
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	if strings.ToLower(cfg.Format) == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler)
}
