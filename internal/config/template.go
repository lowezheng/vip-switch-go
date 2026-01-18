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
