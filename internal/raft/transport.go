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

package raft

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
)

func NewTCPTransport(addr string, logger *slog.Logger) (raft.Transport, error) {
	writer := &logWriter{logger: logger}
	transport, err := raft.NewTCPTransport(addr, nil, 3, 10*time.Second, writer)
	if err != nil {
		return nil, err
	}
	return transport, nil
}

type logWriter struct {
	logger *slog.Logger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.logger.Debug(string(p))
	return len(p), nil
}

// NewRaftLogger creates a logger adapter for Raft
func NewRaftLogger(logger *slog.Logger) hclog.Logger {
	return &raftLogger{
		logger: logger,
	}
}

// raftLogger adapts slog.Logger to Raft's hclog.Logger interface
type raftLogger struct {
	logger *slog.Logger
}

// GetLevel returns current log level
func (l *raftLogger) GetLevel() hclog.Level {
	return hclog.Info // Default to info level
}

// SetLevel sets minimum log level
func (l *raftLogger) SetLevel(level hclog.Level) {
	// slog handles level filtering via handler options
	// This is a no-op for now
}

// Trace logs a trace-level message
func (l *raftLogger) Trace(msg string, args ...interface{}) {
	l.logger.Debug(msg, toKV(args)...)
}

// Debug logs a debug-level message
func (l *raftLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, toKV(args)...)
}

// Info logs an info-level message
func (l *raftLogger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, toKV(args)...)
}

// Warn logs a warning-level message
func (l *raftLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, toKV(args)...)
}

// Error logs an error-level message
func (l *raftLogger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, toKV(args)...)
}

// IsTrace returns true if trace-level logging is enabled
func (l *raftLogger) IsTrace() bool {
	return l.logger.Enabled(context.Background(), slog.LevelDebug)
}

// IsDebug returns true if debug-level logging is enabled
func (l *raftLogger) IsDebug() bool {
	return l.logger.Enabled(context.Background(), slog.LevelDebug)
}

// IsInfo returns true if info-level logging is enabled
func (l *raftLogger) IsInfo() bool {
	return l.logger.Enabled(context.Background(), slog.LevelInfo)
}

// IsWarn returns true if warning-level logging is enabled
func (l *raftLogger) IsWarn() bool {
	return l.logger.Enabled(context.Background(), slog.LevelWarn)
}

// IsError returns true if error-level logging is enabled
func (l *raftLogger) IsError() bool {
	return l.logger.Enabled(context.Background(), slog.LevelError)
}

// ImpliedArgs returns any implied arguments for the logger
func (l *raftLogger) ImpliedArgs() []interface{} {
	return nil
}

// With returns a new logger with the given name and additional key-value pairs
func (l *raftLogger) With(args ...interface{}) hclog.Logger {
	return &raftLogger{
		logger: l.logger.With(toKV(args)...),
	}
}

// Named returns a new logger with the given name
func (l *raftLogger) Named(name string) hclog.Logger {
	return &raftLogger{
		logger: l.logger.With("name", name),
	}
}

// ResetNamed returns a new logger with the given name, without previous context
func (l *raftLogger) ResetNamed(name string) hclog.Logger {
	return &raftLogger{
		logger: slog.Default().With("name", name),
	}
}

// StandardLogger returns a standard library logger
func (l *raftLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(os.Stdout, "", log.LstdFlags)
}

// StandardWriter returns a standard library io.Writer
func (l *raftLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return os.Stdout
}

// Log logs a message with the specified level
func (l *raftLogger) Log(level hclog.Level, msg string, args ...interface{}) {
	var logLevel slog.Level
	switch level {
	case hclog.Trace:
		logLevel = slog.LevelDebug
	case hclog.Debug:
		logLevel = slog.LevelDebug
	case hclog.Info:
		logLevel = slog.LevelInfo
	case hclog.Warn:
		logLevel = slog.LevelWarn
	case hclog.Error:
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	l.logger.Log(context.Background(), logLevel, msg, toKV(args)...)
}

// Name returns name of the logger
func (l *raftLogger) Name() string {
	return "raft-logger"
}

// toKV converts args to slog key-value pairs
func toKV(args []interface{}) []interface{} {
	kv := make([]interface{}, 0, len(args))
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			kv = append(kv, args[i], args[i+1])
		}
	}
	return kv
}
