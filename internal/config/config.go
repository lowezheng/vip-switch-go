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

package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration
type Config struct {
	Node     NodeConfig    `yaml:"node"`
	Cluster  ClusterConfig `yaml:"cluster"`
	Hooks    HooksConfig   `yaml:"hooks"`
	Logging  LoggingConfig `yaml:"logging"`
	filePath string
}

// NodeConfig represents node-specific configuration
type NodeConfig struct {
	ID       string `yaml:"id"`
	RaftAddr string `yaml:"raft_addr"`
	DataDir  string `yaml:"data_dir"`
}

// ClusterConfig represents cluster configuration
type ClusterConfig struct {
	Nodes []ClusterNode `yaml:"nodes"`
}

// ClusterNode represents a single cluster node
type ClusterNode struct {
	ID   string `yaml:"id"`
	Addr string `yaml:"addr"`
}

// HooksConfig represents hooks configuration
type HooksConfig struct {
	Enabled   bool           `yaml:"enabled"`
	Timeout   time.Duration  `yaml:"timeout"`
	OnFailure string         `yaml:"on_failure"` // abort | continue | retry
	ToMaster  HookDefinition `yaml:"ToMaster"`
	ToSlave   HookDefinition `yaml:"ToSlave"`
	ToReady   HookDefinition `yaml:"ToReady"`
	ToDestroy HookDefinition `yaml:"ToDestroy"`
}

// HookDefinition defines a single hook
type HookDefinition struct {
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args"`
	Timeout     time.Duration     `yaml:"timeout"`
	OnFailure   string            `yaml:"on_failure"` // abort | continue | retry
	Environment map[string]string `yaml:"environment"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug | info | warn | error
	Format string `yaml:"format"` // json | text
	Output string `yaml:"output"`
}

// Load loads configuration from a YAML file
func Load(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.filePath = filePath

	// Set defaults
	if cfg.Hooks.Timeout == 0 {
		cfg.Hooks.Timeout = 60 * time.Second
	}
	if cfg.Hooks.OnFailure == "" {
		cfg.Hooks.OnFailure = "abort"
	}

	// Validate
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validate validates configuration
func (c *Config) validate() error {
	if c.Node.ID == "" {
		return fmt.Errorf("node.id is required")
	}
	if c.Node.RaftAddr == "" {
		return fmt.Errorf("node.raft_addr is required")
	}
	if c.Node.DataDir == "" {
		return fmt.Errorf("node.data_dir is required")
	}

	if len(c.Cluster.Nodes) == 0 {
		return fmt.Errorf("cluster.nodes must have at least one entry")
	}

	// Validate log level
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[strings.ToLower(c.Logging.Level)] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.Logging.Level)
	}

	// Validate log format
	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[strings.ToLower(c.Logging.Format)] {
		return fmt.Errorf("invalid log format: %s (must be json or text)", c.Logging.Format)
	}

	return nil
}

// GetClusterPeers returns all peer addresses excluding the current node
func (c *Config) GetClusterPeers() []string {
	var peers []string
	for _, node := range c.Cluster.Nodes {
		if node.ID != c.Node.ID {
			peers = append(peers, node.Addr)
		}
	}
	return peers
}

// GetHookByEventType returns the hook definition for a given event type
func (c *Config) GetHookByEventType(eventType string) (*HookDefinition, error) {
	var hookDef *HookDefinition
	switch eventType {
	case "ToMaster":
		hookDef = &c.Hooks.ToMaster
	case "ToSlave":
		hookDef = &c.Hooks.ToSlave
	case "ToReady":
		hookDef = &c.Hooks.ToReady
	case "ToDestroy":
		hookDef = &c.Hooks.ToDestroy
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}

	// Set defaults from global hooks config if not specified
	if hookDef.Timeout == 0 {
		hookDef.Timeout = c.Hooks.Timeout
	}
	if hookDef.OnFailure == "" {
		hookDef.OnFailure = c.Hooks.OnFailure
	}

	return hookDef, nil
}
