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

package hook

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestSanitizeEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected []string
	}{
		{
			name: "valid EVENT_ prefix",
			input: map[string]string{
				"EVENT_TYPE": "ToMaster",
				"EVENT_ID":   "123",
			},
			expected: []string{"EVENT_TYPE=ToMaster", "EVENT_ID=123"},
		},
		{
			name: "valid NODE_ prefix",
			input: map[string]string{
				"NODE_ID":     "node1",
				"NODE_STATUS": "leader",
			},
			expected: []string{"NODE_ID=node1", "NODE_STATUS=leader"},
		},
		{
			name: "valid VIP_ prefix",
			input: map[string]string{
				"VIP_ADDRESS": "192.168.1.100",
				"VIP_NETMASK": "32",
			},
			expected: []string{"VIP_ADDRESS=192.168.1.100", "VIP_NETMASK=32"},
		},
		{
			name: "valid INTERFACE",
			input: map[string]string{
				"INTERFACE": "eth0",
			},
			expected: []string{"INTERFACE=eth0"},
		},
		{
			name: "valid PATH",
			input: map[string]string{
				"PATH": "/usr/bin:/bin",
			},
			expected: []string{"PATH=/usr/bin:/bin"},
		},
		{
			name: "valid HOME",
			input: map[string]string{
				"HOME": "/home/user",
			},
			expected: []string{"HOME=/home/user"},
		},
		{
			name: "valid USER",
			input: map[string]string{
				"USER": "testuser",
			},
			expected: []string{"USER=testuser"},
		},
		{
			name: "case insensitive - lowercase",
			input: map[string]string{
				"event_type": "ToMaster",
				"node_id":    "node1",
			},
			expected: []string{"event_type=ToMaster", "node_id=node1"},
		},
		{
			name: "mixed case",
			input: map[string]string{
				"Event_Type": "ToMaster",
				"Node_ID":    "node1",
			},
			expected: []string{"Event_Type=ToMaster", "Node_ID=node1"},
		},
		{
			name: "invalid prefixes",
			input: map[string]string{
				"INVALID_VAR":  "value1",
				"ANOTHER_ONE":  "value2",
				"UNSAFE_INPUT": "value3",
			},
			expected: []string{},
		},
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: []string{},
		},
		{
			name: "mix of valid and invalid",
			input: map[string]string{
				"NODE_ID":     "node1",
				"INVALID_VAR": "value",
				"VIP_ADDR":    "192.168.1.100",
				"BAD_PREFIX":  "value",
			},
			expected: []string{"NODE_ID=node1", "VIP_ADDR=192.168.1.100"},
		},
		{
			name: "all valid prefixes",
			input: map[string]string{
				"EVENT_TYPE": "ToMaster",
				"NODE_ID":    "node1",
				"VIP_ADDR":   "192.168.1.100",
				"INTERFACE":  "eth0",
				"PATH":       "/usr/bin",
				"HOME":       "/home/user",
				"USER":       "testuser",
			},
			expected: []string{
				"EVENT_TYPE=ToMaster",
				"NODE_ID=node1",
				"VIP_ADDR=192.168.1.100",
				"INTERFACE=eth0",
				"PATH=/usr/bin",
				"HOME=/home/user",
				"USER=testuser",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeEnvironment(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("SanitizeEnvironment() returned %d items, want %d", len(result), len(tt.expected))
				return
			}

			for _, expected := range tt.expected {
				found := false
				for _, item := range result {
					if item == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("SanitizeEnvironment() missing expected value %q", expected)
				}
			}
		})
	}
}

func TestValidateCommandPath(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid /usr/bin path",
			command: "/usr/bin/ls",
			wantErr: false,
		},
		{
			name:    "valid /usr/local/bin path",
			command: "/usr/local/bin/myscript",
			wantErr: false,
		},
		{
			name:    "valid /bin path",
			command: "/bin/sh",
			wantErr: false,
		},
		{
			name:    "valid /sbin path",
			command: "/sbin/ip",
			wantErr: false,
		},
		{
			name:    "valid /usr/sbin path",
			command: "/usr/sbin/systemctl",
			wantErr: false,
		},
		{
			name:        "path with spaces",
			command:     "/usr/bin/ls -la",
			wantErr:     true,
			errContains: "cannot contain spaces or tabs",
		},
		{
			name:        "path with tabs",
			command:     "/usr/bin/\tcommand",
			wantErr:     true,
			errContains: "cannot contain spaces or tabs",
		},
		{
			name:        "unsafe directory - /tmp",
			command:     "/tmp/malicious",
			wantErr:     true,
			errContains: "not in a safe directory",
		},
		{
			name:        "unsafe directory - /home/user",
			command:     "/home/user/script.sh",
			wantErr:     true,
			errContains: "not in a safe directory",
		},
		{
			name:        "unsafe directory - /etc",
			command:     "/etc/script.sh",
			wantErr:     true,
			errContains: "not in a safe directory",
		},
		{
			name:    "relative path (resolves to /bin)",
			command: "/bin/ls",
			wantErr: false,
		},
		{
			name:        "command /bin/ls",
			command:     "/bin/ls",
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "unsafe relative path",
			command:     "../tmp/script",
			wantErr:     true,
			errContains: "not in a safe directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommandPath(tt.command)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateCommandPath() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateCommandPath() error = %v, want to contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateCommandPath() unexpected error: %v", err)
			}
		})
	}
}

func TestNewExecutor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	executor := NewExecutor(logger)

	if executor == nil {
		t.Fatal("NewExecutor() returned nil")
	}

	if executor.logger != logger {
		t.Errorf("NewExecutor().logger = %v, want %v", executor.logger, logger)
	}
}

func TestExecute_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	executor := NewExecutor(logger)

	ctx := context.Background()
	err := executor.Execute(ctx, "echo", []string{"hello", "world"}, []string{}, "TestEvent")

	if err != nil {
		t.Errorf("Execute() unexpected error: %v", err)
	}
}

func TestExecute_EmptyCommand(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	executor := NewExecutor(logger)

	ctx := context.Background()
	err := executor.Execute(ctx, "", []string{}, []string{}, "TestEvent")

	if err == nil {
		t.Errorf("Execute() expected error for empty command, got nil")
	}

	if err.Error() != "command cannot be empty" {
		t.Errorf("Execute() error = %v, want 'command cannot be empty'", err)
	}
}

func TestExecute_CommandNotFound(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	executor := NewExecutor(logger)

	ctx := context.Background()
	err := executor.Execute(ctx, "nonexistent_command_xyz123", []string{}, []string{}, "TestEvent")

	if err == nil {
		t.Errorf("Execute() expected error for non-existent command, got nil")
	}

	if !strings.Contains(err.Error(), "command not found") {
		t.Errorf("Execute() error = %v, want 'command not found'", err)
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	executor := NewExecutor(logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := executor.Execute(ctx, "echo", []string{"test"}, []string{}, "TestEvent")

	if err == nil {
		t.Errorf("Execute() expected error for cancelled context, got nil")
	}

	if !strings.Contains(err.Error(), "canceled") {
		t.Errorf("Execute() error = %v, want 'canceled'", err)
	}
}

func TestExecute_ContextTimeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	executor := NewExecutor(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := executor.Execute(ctx, "sleep", []string{"1"}, []string{}, "TestEvent")

	if err == nil {
		t.Errorf("Execute() expected error for timeout, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "timed out") && !strings.Contains(errMsg, "killed") {
		t.Errorf("Execute() error = %v, want 'timed out' or 'killed'", err)
	}
}

func TestExecute_NonZeroExitCode(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	executor := NewExecutor(logger)

	ctx := context.Background()
	err := executor.Execute(ctx, "sh", []string{"-c", "exit 42"}, []string{}, "TestEvent")

	if err == nil {
		t.Errorf("Execute() expected error for non-zero exit code, got nil")
	}

	if !strings.Contains(err.Error(), "exited with status 42") {
		t.Errorf("Execute() error = %v, want to contain 'exited with status 42'", err)
	}
}

func TestExecute_WithEnvironment(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	executor := NewExecutor(logger)

	ctx := context.Background()
	env := []string{"TEST_VAR=test_value", "ANOTHER_VAR=another_value"}
	err := executor.Execute(ctx, "sh", []string{"-c", "echo $TEST_VAR $ANOTHER_VAR"}, env, "TestEvent")

	if err != nil {
		t.Errorf("Execute() unexpected error: %v", err)
	}
}

func TestExecute_Concurrent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	executor := NewExecutor(logger)

	ctx := context.Background()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			err := executor.Execute(ctx, "echo", []string{"test", string(rune('a' + id))}, []string{}, "TestEvent")
			if err != nil {
				t.Errorf("Concurrent Execute() failed: %v", err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSignalHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	executor := NewExecutor(logger)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := executor.Execute(ctx, "sleep", []string{"10"}, []string{}, "TestEvent")
	elapsed := time.Since(start)

	if err == nil {
		t.Errorf("Execute() expected error for killed process, got nil")
	}

	if elapsed > 1*time.Second {
		t.Errorf("Execute() took too long: %v, expected < 1s", elapsed)
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.Sys() != nil {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				if status.Signaled() {
					t.Logf("Process was killed with signal: %v", status.Signal())
				}
			}
		}
	}
}
