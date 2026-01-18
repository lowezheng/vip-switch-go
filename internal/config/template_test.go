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
	"testing"
)

func TestExpandTemplate(t *testing.T) {
	tests := []struct {
		name        string
		templateStr string
		data        TemplateData
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:        "simple variable expansion",
			templateStr: "Node {{.NodeID}}",
			data:        TemplateData{NodeID: "node1"},
			want:        "Node node1",
			wantErr:     false,
		},
		{
			name:        "multiple variables",
			templateStr: "{{.NodeID}}-{{.Event}}-{{.RaftAddr}}",
			data:        TemplateData{NodeID: "node1", Event: "ToMaster", RaftAddr: "127.0.0.1:10001"},
			want:        "node1-ToMaster-127.0.0.1:10001",
			wantErr:     false,
		},
		{
			name:        "no variables",
			templateStr: "static text",
			data:        TemplateData{NodeID: "node1"},
			want:        "static text",
			wantErr:     false,
		},
		{
			name:        "variable in middle",
			templateStr: "prefix {{.NodeID}} suffix",
			data:        TemplateData{NodeID: "node1"},
			want:        "prefix node1 suffix",
			wantErr:     false,
		},
		{
			name:        "undefined variable",
			templateStr: "{{.UndefinedVar}}",
			data:        TemplateData{NodeID: "node1"},
			wantErr:     true,
			errContains: "",
		},
		{
			name:        "invalid template syntax",
			templateStr: "{{.NodeID",
			data:        TemplateData{NodeID: "node1"},
			wantErr:     true,
			errContains: "",
		},
		{
			name:        "empty template",
			templateStr: "",
			data:        TemplateData{NodeID: "node1"},
			want:        "",
			wantErr:     false,
		},
		{
			name:        "variable with special chars",
			templateStr: "{{.Event}}_hook_{{.NodeID}}",
			data:        TemplateData{Event: "ToMaster", NodeID: "node1"},
			want:        "ToMaster_hook_node1",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandTemplate(tt.templateStr, tt.data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExpandTemplate() expected error, got nil")
					return
				}
				return
			}

			if err != nil {
				t.Errorf("ExpandTemplate() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("ExpandTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		data        TemplateData
		want        map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "expand multiple environment variables",
			env: map[string]string{
				"NODE_ID":    "{{.NodeID}}",
				"EVENT_TYPE": "{{.Event}}",
				"RAFT_ADDR":  "{{.RaftAddr}}",
			},
			data: TemplateData{
				NodeID:   "node1",
				Event:    "ToMaster",
				RaftAddr: "127.0.0.1:10001",
			},
			want: map[string]string{
				"NODE_ID":    "node1",
				"EVENT_TYPE": "ToMaster",
				"RAFT_ADDR":  "127.0.0.1:10001",
			},
			wantErr: false,
		},
		{
			name:    "empty environment map",
			env:     map[string]string{},
			data:    TemplateData{NodeID: "node1"},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "mix of templates and static values",
			env: map[string]string{
				"NODE_ID":  "{{.NodeID}}",
				"STATIC":   "static_value",
				"COMBINED": "{{.Event}}_static",
			},
			data: TemplateData{
				NodeID: "node1",
				Event:  "ToMaster",
			},
			want: map[string]string{
				"NODE_ID":  "node1",
				"STATIC":   "static_value",
				"COMBINED": "ToMaster_static",
			},
			wantErr: false,
		},
		{
			name: "template error in environment",
			env: map[string]string{
				"NODE_ID": "{{.UndefinedVar}}",
			},
			data:    TemplateData{NodeID: "node1"},
			wantErr: true,
		},
		{
			name: "multiple variables in single value",
			env: map[string]string{
				"COMBO": "{{.NodeID}}_{{.Event}}",
			},
			data: TemplateData{
				NodeID: "node1",
				Event:  "ToMaster",
			},
			want: map[string]string{
				"COMBO": "node1_ToMaster",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandEnvironment(tt.env, tt.data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExpandEnvironment() expected error, got nil")
					return
				}
				return
			}

			if err != nil {
				t.Errorf("ExpandEnvironment() unexpected error: %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ExpandEnvironment() returned %d items, want %d", len(got), len(tt.want))
				return
			}

			for key, expectedValue := range tt.want {
				if got[key] != expectedValue {
					t.Errorf("ExpandEnvironment()[%s] = %v, want %v", key, got[key], expectedValue)
				}
			}
		})
	}
}
