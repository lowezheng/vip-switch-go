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
	"log/slog"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
)

func TestNewRaftLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	if raftLogger == nil {
		t.Fatal("NewRaftLogger() returned nil")
	}

	if raftLogger.Name() != "raft-logger" {
		t.Errorf("NewRaftLogger().Name() = %v, want raft-logger", raftLogger.Name())
	}
}

func TestRaftLogger_LogLevels(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	tests := []struct {
		name    string
		logFunc func(msg string, args ...interface{})
	}{
		{"Trace", raftLogger.Trace},
		{"Debug", raftLogger.Debug},
		{"Info", raftLogger.Info},
		{"Warn", raftLogger.Warn},
		{"Error", raftLogger.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc("test message", "key", "value")
		})
	}
}

func TestRaftLogger_IsTrace(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	raftLogger := NewRaftLogger(logger)

	isEnabled := raftLogger.IsTrace()
	if !isEnabled {
		t.Error("IsTrace() returned false, expected true with debug level")
	}
}

func TestRaftLogger_IsDebug(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	raftLogger := NewRaftLogger(logger)

	isEnabled := raftLogger.IsDebug()
	if !isEnabled {
		t.Error("IsDebug() returned false, expected true with debug level")
	}
}

func TestRaftLogger_IsInfo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	isEnabled := raftLogger.IsInfo()
	if !isEnabled {
		t.Error("IsInfo() returned false, expected true")
	}
}

func TestRaftLogger_IsWarn(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	isEnabled := raftLogger.IsWarn()
	if !isEnabled {
		t.Error("IsWarn() returned false, expected true")
	}
}

func TestRaftLogger_IsError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	isEnabled := raftLogger.IsError()
	if !isEnabled {
		t.Error("IsError() returned false, expected true")
	}
}

func TestRaftLogger_GetLevel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	level := raftLogger.GetLevel()
	if level != hclog.Info {
		t.Errorf("GetLevel() = %v, want hclog.Info", level)
	}
}

func TestRaftLogger_SetLevel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	raftLogger.SetLevel(hclog.Debug)
}

func TestRaftLogger_ImpliedArgs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	args := raftLogger.ImpliedArgs()
	if args != nil {
		t.Errorf("ImpliedArgs() = %v, want nil", args)
	}
}

func TestRaftLogger_With(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	newLogger := raftLogger.With("key", "value")

	if newLogger == nil {
		t.Fatal("With() returned nil")
	}

	if newLogger.Name() != "raft-logger" {
		t.Errorf("With().Name() = %v, want raft-logger", newLogger.Name())
	}
}

func TestRaftLogger_Named(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	newLogger := raftLogger.Named("sub-logger")

	if newLogger == nil {
		t.Fatal("Named() returned nil")
	}

	if newLogger.Name() != "raft-logger" {
		t.Errorf("Named().Name() = %v, want raft-logger", newLogger.Name())
	}
}

func TestRaftLogger_ResetNamed(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	newLogger := raftLogger.ResetNamed("reset-logger")

	if newLogger == nil {
		t.Fatal("ResetNamed() returned nil")
	}

	if newLogger.Name() != "raft-logger" {
		t.Errorf("ResetNamed().Name() = %v, want raft-logger", newLogger.Name())
	}
}

func TestRaftLogger_Log(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	tests := []struct {
		name  string
		level hclog.Level
	}{
		{"Trace", hclog.Trace},
		{"Debug", hclog.Debug},
		{"Info", hclog.Info},
		{"Warn", hclog.Warn},
		{"Error", hclog.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raftLogger.Log(tt.level, "test message", "key", "value")
		})
	}
}

func TestRaftLogger_StandardLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	stdLogger := raftLogger.StandardLogger(nil)

	if stdLogger == nil {
		t.Fatal("StandardLogger() returned nil")
	}
}

func TestRaftLogger_StandardWriter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	raftLogger := NewRaftLogger(logger)

	writer := raftLogger.StandardWriter(nil)

	if writer == nil {
		t.Fatal("StandardWriter() returned nil")
	}

	if writer != os.Stdout {
		t.Errorf("StandardWriter() = %v, want os.Stdout", writer)
	}
}

func TestLogWriter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	writer := &logWriter{logger: logger}

	message := "test log message"
	n, err := writer.Write([]byte(message))

	if err != nil {
		t.Errorf("Write() error = %v, want nil", err)
	}

	if n != len(message) {
		t.Errorf("Write() returned %d bytes, want %d", n, len(message))
	}
}

func TestToKV(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected []interface{}
	}{
		{
			name:     "even number of args",
			args:     []interface{}{"key1", "value1", "key2", "value2"},
			expected: []interface{}{"key1", "value1", "key2", "value2"},
		},
		{
			name:     "empty args",
			args:     []interface{}{},
			expected: []interface{}{},
		},
		{
			name:     "single pair",
			args:     []interface{}{"key", "value"},
			expected: []interface{}{"key", "value"},
		},
		{
			name:     "odd number of args (drop last)",
			args:     []interface{}{"key1", "value1", "key2"},
			expected: []interface{}{"key1", "value1"},
		},
		{
			name:     "multiple pairs",
			args:     []interface{}{"a", "1", "b", "2", "c", "3"},
			expected: []interface{}{"a", "1", "b", "2", "c", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toKV(tt.args)

			if len(result) != len(tt.expected) {
				t.Errorf("toKV() returned %d items, want %d", len(result), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("toKV()[%d] = %v, want %v", i, result[i], expected)
				}
			}
		})
	}
}
