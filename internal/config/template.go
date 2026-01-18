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
	"bytes"
	"text/template"
)

// TemplateData provides data for template expansion
type TemplateData struct {
	NodeID   string
	Event    string
	RaftAddr string
}

// ExpandTemplate expands template variables in a string
func ExpandTemplate(templateStr string, data TemplateData) (string, error) {
	tmpl, err := template.New("hook").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExpandEnvironment expands template variables in environment variables map
func ExpandEnvironment(env map[string]string, data TemplateData) (map[string]string, error) {
	result := make(map[string]string, len(env))
	for key, value := range env {
		expanded, err := ExpandTemplate(value, data)
		if err != nil {
			return nil, err
		}
		result[key] = expanded
	}
	return result, nil
}
