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
	"fmt"
	"log/slog"
	"time"

	"vip-switch-go/internal/config"
)

// System manages hook execution
type System struct {
	config       *config.Config
	logger       *slog.Logger
	executor     *Executor
	templateData config.TemplateData
}

// NewSystem creates a new hook system
func NewSystem(cfg *config.Config, logger *slog.Logger) *System {
	return &System{
		config:   cfg,
		logger:   logger,
		executor: NewExecutor(logger),
		templateData: config.TemplateData{
			NodeID:   cfg.Node.ID,
			RaftAddr: cfg.Node.RaftAddr,
		},
	}
}

// ExecuteHook executes a hook by event type
func (s *System) ExecuteHook(ctx context.Context, eventType string) error {
	if !s.config.Hooks.Enabled {
		s.logger.Debug("Hooks disabled, skipping", "event_type", eventType)
		return nil
	}

	hookDef, err := s.config.GetHookByEventType(eventType)
	if err != nil {
		return fmt.Errorf("failed to get hook definition: %w", err)
	}

	if hookDef.Command == "" {
		s.logger.Debug("No hook command configured", "event_type", eventType)
		return nil
	}

	s.logger.Info("Executing hook", "event_type", eventType, "command", hookDef.Command)

	// Set event type in template data
	s.templateData.Event = eventType

	// Expand environment variables
	env, err := config.ExpandEnvironment(hookDef.Environment, s.templateData)
	if err != nil {
		return fmt.Errorf("failed to expand environment variables: %w", err)
	}

	// Create hook context
	hookCtx, cancel := context.WithTimeout(ctx, hookDef.Timeout)
	defer cancel()

	// Build environment for hook
	osEnv := buildOSEnv(env, s.templateData)

	// Execute hook
	err = s.executor.Execute(hookCtx, hookDef.Command, hookDef.Args, osEnv, eventType)

	if err != nil {
		s.logger.Error("Hook execution failed", "event_type", eventType, "error", err)

		// Handle based on failure strategy
		switch hookDef.OnFailure {
		case "abort":
			return fmt.Errorf("hook failed with abort strategy: %w", err)
		case "continue":
			s.logger.Warn("Hook failed but continuing due to continue strategy", "event_type", eventType)
			return nil
		case "retry":
			return s.retryHook(hookCtx, hookDef, osEnv, eventType, 3)
		default:
			return fmt.Errorf("hook failed with unknown strategy '%s': %w", hookDef.OnFailure, err)
		}
	}

	s.logger.Info("Hook executed successfully", "event_type", eventType)
	return nil
}

// retryHook retries hook execution with exponential backoff
func (s *System) retryHook(ctx context.Context, hookDef *config.HookDefinition, env []string, eventType string, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(i*i) * time.Second
			s.logger.Info("Retrying hook", "event_type", eventType, "attempt", i+1, "backoff", backoff)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := s.executor.Execute(ctx, hookDef.Command, hookDef.Args, env, eventType)
		if err == nil {
			return nil
		}

		lastErr = err
		s.logger.Warn("Hook retry failed", "event_type", eventType, "attempt", i+1, "error", err)
	}

	return fmt.Errorf("hook failed after %d retries: %w", maxRetries, lastErr)
}

// buildOSEnv builds OS environment variables for hook
func buildOSEnv(hookEnv map[string]string, data config.TemplateData) []string {
	env := make([]string, 0, len(hookEnv)+2)

	// Add standard environment variables
	env = append(env, fmt.Sprintf("EVENT_TYPE=%s", data.Event))
	env = append(env, fmt.Sprintf("NODE_ID=%s", data.NodeID))

	// Add hook-specific environment variables
	for key, value := range hookEnv {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}
