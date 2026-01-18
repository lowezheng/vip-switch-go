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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		errContains string
		checkConfig func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid minimal config",
			yamlContent: `
node:
  id: node1
  raft_addr: 127.0.0.1:10001
  data_dir: ./data/node1
cluster:
  nodes:
    - id: node1
      addr: 127.0.0.1:10001
hooks:
  enabled: true
logging:
  level: info
  format: json
`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				if cfg.Node.ID != "node1" {
					t.Errorf("Node.ID = %v, want node1", cfg.Node.ID)
				}
				if cfg.Node.RaftAddr != "127.0.0.1:10001" {
					t.Errorf("Node.RaftAddr = %v, want 127.0.0.1:10001", cfg.Node.RaftAddr)
				}
				if cfg.Node.DataDir != "./data/node1" {
					t.Errorf("Node.DataDir = %v, want ./data/node1", cfg.Node.DataDir)
				}
			},
		},
		{
			name: "config with default timeout and on_failure",
			yamlContent: `
node:
  id: node1
  raft_addr: 127.0.0.1:10001
  data_dir: ./data/node1
cluster:
  nodes:
    - id: node1
      addr: 127.0.0.1:10001
hooks:
  enabled: true
logging:
  level: info
  format: json
`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				if cfg.Hooks.Timeout != 60*time.Second {
					t.Errorf("Hooks.Timeout = %v, want 60s", cfg.Hooks.Timeout)
				}
				if cfg.Hooks.OnFailure != "abort" {
					t.Errorf("Hooks.OnFailure = %v, want abort", cfg.Hooks.OnFailure)
				}
			},
		},
		{
			name: "config with custom timeout and on_failure",
			yamlContent: `
node:
  id: node1
  raft_addr: 127.0.0.1:10001
  data_dir: ./data/node1
cluster:
  nodes:
    - id: node1
      addr: 127.0.0.1:10001
hooks:
  enabled: true
  timeout: 30s
  on_failure: continue
logging:
  level: info
  format: json
`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				if cfg.Hooks.Timeout != 30*time.Second {
					t.Errorf("Hooks.Timeout = %v, want 30s", cfg.Hooks.Timeout)
				}
				if cfg.Hooks.OnFailure != "continue" {
					t.Errorf("Hooks.OnFailure = %v, want continue", cfg.Hooks.OnFailure)
				}
			},
		},
		{
			name:        "file not found",
			yamlContent: "",
			wantErr:     true,
			errContains: "failed to read config file",
		},
		{
			name: "invalid YAML",
			yamlContent: `
node:
  id: [invalid
`,
			wantErr:     true,
			errContains: "failed to parse config",
		},
		{
			name: "missing node.id",
			yamlContent: `
node:
  raft_addr: 127.0.0.1:10001
  data_dir: ./data/node1
cluster:
  nodes:
    - id: node1
      addr: 127.0.0.1:10001
hooks:
  enabled: true
logging:
  level: info
  format: json
`,
			wantErr:     true,
			errContains: "node.id is required",
		},
		{
			name: "missing node.raft_addr",
			yamlContent: `
node:
  id: node1
  data_dir: ./data/node1
cluster:
  nodes:
    - id: node1
      addr: 127.0.0.1:10001
hooks:
  enabled: true
logging:
  level: info
  format: json
`,
			wantErr:     true,
			errContains: "node.raft_addr is required",
		},
		{
			name: "missing node.data_dir",
			yamlContent: `
node:
  id: node1
  raft_addr: 127.0.0.1:10001
cluster:
  nodes:
    - id: node1
      addr: 127.0.0.1:10001
hooks:
  enabled: true
logging:
  level: info
  format: json
`,
			wantErr:     true,
			errContains: "node.data_dir is required",
		},
		{
			name: "empty cluster.nodes",
			yamlContent: `
node:
  id: node1
  raft_addr: 127.0.0.1:10001
  data_dir: ./data/node1
cluster:
  nodes: []
hooks:
  enabled: true
logging:
  level: info
  format: json
`,
			wantErr:     true,
			errContains: "cluster.nodes must have at least one entry",
		},
		{
			name: "invalid log level",
			yamlContent: `
node:
  id: node1
  raft_addr: 127.0.0.1:10001
  data_dir: ./data/node1
cluster:
  nodes:
    - id: node1
      addr: 127.0.0.1:10001
hooks:
  enabled: true
logging:
  level: invalid
  format: json
`,
			wantErr:     true,
			errContains: "invalid log level",
		},
		{
			name: "invalid log format",
			yamlContent: `
node:
  id: node1
  raft_addr: 127.0.0.1:10001
  data_dir: ./data/node1
cluster:
  nodes:
    - id: node1
      addr: 127.0.0.1:10001
hooks:
  enabled: true
logging:
  level: info
  format: invalid
`,
			wantErr:     true,
			errContains: "invalid log format",
		},
		{
			name: "case-insensitive log level",
			yamlContent: `
node:
  id: node1
  raft_addr: 127.0.0.1:10001
  data_dir: ./data/node1
cluster:
  nodes:
    - id: node1
      addr: 127.0.0.1:10001
hooks:
  enabled: true
logging:
  level: DEBUG
  format: TEXT
`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				if cfg.Logging.Level != "DEBUG" {
					t.Errorf("Logging.Level = %v, want DEBUG", cfg.Logging.Level)
				}
				if cfg.Logging.Format != "TEXT" {
					t.Errorf("Logging.Format = %v, want TEXT", cfg.Logging.Format)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.yamlContent != "" {
				filePath = filepath.Join(tmpDir, "config.yaml")
				if err := os.WriteFile(filePath, []byte(tt.yamlContent), 0644); err != nil {
					t.Fatalf("failed to write test config: %v", err)
				}
			} else {
				filePath = filepath.Join(tmpDir, "nonexistent.yaml")
			}

			cfg, err := Load(filePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error, got nil")
					return
				}
				if tt.errContains != "" {
					if !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("Load() error = %v, want to contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Load() unexpected error: %v", err)
				return
			}

			if tt.checkConfig != nil {
				tt.checkConfig(t, cfg)
			}
		})
	}
}

func TestGetClusterPeers(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{
			name: "single node - no peers",
			config: &Config{
				Node: NodeConfig{ID: "node1"},
				Cluster: ClusterConfig{
					Nodes: []ClusterNode{
						{ID: "node1", Addr: "127.0.0.1:10001"},
					},
				},
			},
			expected: []string{},
		},
		{
			name: "multi node - returns peers only",
			config: &Config{
				Node: NodeConfig{ID: "node1"},
				Cluster: ClusterConfig{
					Nodes: []ClusterNode{
						{ID: "node1", Addr: "127.0.0.1:10001"},
						{ID: "node2", Addr: "127.0.0.1:10002"},
						{ID: "node3", Addr: "127.0.0.1:10003"},
					},
				},
			},
			expected: []string{"127.0.0.1:10002", "127.0.0.1:10003"},
		},
		{
			name: "current node at end",
			config: &Config{
				Node: NodeConfig{ID: "node3"},
				Cluster: ClusterConfig{
					Nodes: []ClusterNode{
						{ID: "node1", Addr: "127.0.0.1:10001"},
						{ID: "node2", Addr: "127.0.0.1:10002"},
						{ID: "node3", Addr: "127.0.0.1:10003"},
					},
				},
			},
			expected: []string{"127.0.0.1:10001", "127.0.0.1:10002"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetClusterPeers()

			if len(result) != len(tt.expected) {
				t.Errorf("GetClusterPeers() returned %d items, want %d", len(result), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("GetClusterPeers()[%d] = %v, want %v", i, result[i], expected)
				}
			}
		})
	}
}

func TestGetHookByEventType(t *testing.T) {
	cfg := &Config{
		Hooks: HooksConfig{
			Timeout:   60 * time.Second,
			OnFailure: "abort",
			ToMaster: HookDefinition{
				Command:   "/usr/local/bin/on-master.sh",
				Timeout:   30 * time.Second,
				OnFailure: "retry",
			},
			ToSlave: HookDefinition{
				Command: "/usr/local/bin/on-slave.sh",
			},
			ToReady: HookDefinition{
				Command: "/usr/local/bin/on-ready.sh",
			},
			ToDestroy: HookDefinition{
				Command: "/usr/local/bin/on-destroy.sh",
			},
		},
	}

	tests := []struct {
		name          string
		eventType     string
		wantErr       bool
		errContains   string
		expectedCmd   string
		expectedTout  time.Duration
		expectedOnErr string
	}{
		{
			name:          "ToMaster",
			eventType:     "ToMaster",
			wantErr:       false,
			expectedCmd:   "/usr/local/bin/on-master.sh",
			expectedTout:  30 * time.Second,
			expectedOnErr: "retry",
		},
		{
			name:          "ToSlave",
			eventType:     "ToSlave",
			wantErr:       false,
			expectedCmd:   "/usr/local/bin/on-slave.sh",
			expectedTout:  60 * time.Second,
			expectedOnErr: "abort",
		},
		{
			name:          "ToReady",
			eventType:     "ToReady",
			wantErr:       false,
			expectedCmd:   "/usr/local/bin/on-ready.sh",
			expectedTout:  60 * time.Second,
			expectedOnErr: "abort",
		},
		{
			name:          "ToDestroy",
			eventType:     "ToDestroy",
			wantErr:       false,
			expectedCmd:   "/usr/local/bin/on-destroy.sh",
			expectedTout:  60 * time.Second,
			expectedOnErr: "abort",
		},
		{
			name:        "unknown event type",
			eventType:   "UnknownEvent",
			wantErr:     true,
			errContains: "unknown event type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hookDef, err := cfg.GetHookByEventType(tt.eventType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetHookByEventType() expected error, got nil")
					return
				}
				if tt.errContains != "" {
					if !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("GetHookByEventType() error = %v, want to contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("GetHookByEventType() unexpected error: %v", err)
				return
			}

			if hookDef.Command != tt.expectedCmd {
				t.Errorf("Command = %v, want %v", hookDef.Command, tt.expectedCmd)
			}
			if hookDef.Timeout != tt.expectedTout {
				t.Errorf("Timeout = %v, want %v", hookDef.Timeout, tt.expectedTout)
			}
			if hookDef.OnFailure != tt.expectedOnErr {
				t.Errorf("OnFailure = %v, want %v", hookDef.OnFailure, tt.expectedOnErr)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			config: &Config{
				Node: NodeConfig{
					ID:       "node1",
					RaftAddr: "127.0.0.1:10001",
					DataDir:  "./data/node1",
				},
				Cluster: ClusterConfig{
					Nodes: []ClusterNode{
						{ID: "node1", Addr: "127.0.0.1:10001"},
					},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr: false,
		},
		{
			name: "missing node.id",
			config: &Config{
				Node: NodeConfig{
					RaftAddr: "127.0.0.1:10001",
					DataDir:  "./data/node1",
				},
				Cluster: ClusterConfig{
					Nodes: []ClusterNode{
						{ID: "node1", Addr: "127.0.0.1:10001"},
					},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr:     true,
			errContains: "node.id is required",
		},
		{
			name: "missing node.raft_addr",
			config: &Config{
				Node: NodeConfig{
					ID:      "node1",
					DataDir: "./data/node1",
				},
				Cluster: ClusterConfig{
					Nodes: []ClusterNode{
						{ID: "node1", Addr: "127.0.0.1:10001"},
					},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr:     true,
			errContains: "node.raft_addr is required",
		},
		{
			name: "missing node.data_dir",
			config: &Config{
				Node: NodeConfig{
					ID:       "node1",
					RaftAddr: "127.0.0.1:10001",
				},
				Cluster: ClusterConfig{
					Nodes: []ClusterNode{
						{ID: "node1", Addr: "127.0.0.1:10001"},
					},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr:     true,
			errContains: "node.data_dir is required",
		},
		{
			name: "empty cluster.nodes",
			config: &Config{
				Node: NodeConfig{
					ID:       "node1",
					RaftAddr: "127.0.0.1:10001",
					DataDir:  "./data/node1",
				},
				Cluster: ClusterConfig{
					Nodes: []ClusterNode{},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr:     true,
			errContains: "cluster.nodes must have at least one entry",
		},
		{
			name: "invalid log level",
			config: &Config{
				Node: NodeConfig{
					ID:       "node1",
					RaftAddr: "127.0.0.1:10001",
					DataDir:  "./data/node1",
				},
				Cluster: ClusterConfig{
					Nodes: []ClusterNode{
						{ID: "node1", Addr: "127.0.0.1:10001"},
					},
				},
				Logging: LoggingConfig{
					Level:  "invalid",
					Format: "json",
				},
			},
			wantErr:     true,
			errContains: "invalid log level",
		},
		{
			name: "invalid log format",
			config: &Config{
				Node: NodeConfig{
					ID:       "node1",
					RaftAddr: "127.0.0.1:10001",
					DataDir:  "./data/node1",
				},
				Cluster: ClusterConfig{
					Nodes: []ClusterNode{
						{ID: "node1", Addr: "127.0.0.1:10001"},
					},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "invalid",
				},
			},
			wantErr:     true,
			errContains: "invalid log format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("validate() expected error, got nil")
					return
				}
				if tt.errContains != "" {
					if !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("validate() error = %v, want to contain %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("validate() unexpected error: %v", err)
			}
		})
	}
}
