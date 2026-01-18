package hook

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

// Executor handles secure command execution
type Executor struct {
	logger *slog.Logger
}

// NewExecutor creates a new command executor
func NewExecutor(logger *slog.Logger) *Executor {
	return &Executor{
		logger: logger,
	}
}

// Execute executes a command securely with streaming output
func (e *Executor) Execute(ctx context.Context, command string, args []string, env []string, eventType string) error {
	if command == "" {
		return errors.New("command cannot be empty")
	}

	// Validate command path
	cmdPath, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("command not found: %w", err)
	}

	e.logger.Debug("Executing command", "command", cmdPath, "args", args, "event_type", eventType)

	// Create command
	cmd := exec.CommandContext(ctx, cmdPath, args...)
	cmd.Env = env

	// Setup stdout and stderr pipes
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Stream output in goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	go e.streamOutput(&wg, stdoutPipe, "stdout", eventType)
	go e.streamOutput(&wg, stderrPipe, "stderr", eventType)

	// Wait for command to complete
	err = cmd.Wait()
	wg.Wait()

	// Check exit status
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			e.logger.Error("Command exited with non-zero status",
				"event_type", eventType,
				"exit_code", status.ExitStatus(),
				"signal", status.Signal(),
			)
			return fmt.Errorf("command exited with status %d: %w", status.ExitStatus(), exitErr)
		}
	}

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("command timed out: %w", err)
		}
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("command canceled: %w", err)
		}
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// streamOutput streams command output to logs
func (e *Executor) streamOutput(wg *sync.WaitGroup, pipe io.Reader, streamName, eventType string) {
	defer wg.Done()

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		e.logger.Info("Hook output",
			"event_type", eventType,
			"stream", streamName,
			"line", line,
		)
	}

	if err := scanner.Err(); err != nil {
		e.logger.Error("Error reading output",
			"event_type", eventType,
			"stream", streamName,
			"error", err,
		)
	}
}

// SanitizeEnvironment sanitizes environment variables to prevent injection
func SanitizeEnvironment(env map[string]string) []string {
	sanitized := make([]string, 0, len(env))

	allowedPrefixes := []string{
		"EVENT_",
		"NODE_",
		"VIP_",
		"INTERFACE",
		"PATH",
		"HOME",
		"USER",
	}

	for key, value := range env {
		if isEnvKeyAllowed(key, allowedPrefixes) {
			sanitized = append(sanitized, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return sanitized
}

// isEnvKeyAllowed checks if environment variable key is allowed
func isEnvKeyAllowed(key string, allowedPrefixes []string) bool {
	upperKey := strings.ToUpper(key)

	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(upperKey, prefix) {
			return true
		}
	}

	return false
}

// ValidateCommandPath validates that a command path is safe
func ValidateCommandPath(command string) error {
	if strings.Contains(command, " ") || strings.Contains(command, "\t") {
		return errors.New("command path cannot contain spaces or tabs")
	}

	absPath, err := filepath.Abs(command)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check if path is in safe directories
	safeDirs := []string{
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
		"/usr/sbin",
		"/sbin",
	}

	for _, dir := range safeDirs {
		if strings.HasPrefix(absPath, dir) {
			return nil
		}
	}

	return fmt.Errorf("command path '%s' is not in a safe directory", absPath)
}
