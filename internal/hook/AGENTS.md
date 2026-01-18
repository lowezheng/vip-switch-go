## OVERVIEW
Secure command execution system with streaming output and retry orchestration.

## WHERE TO LOOK
| Task | File | Notes |
|------|------|-------|
| Command execution | executor.go | Execute() with stdout/stderr streaming |
| Hook orchestration | system.go | ExecuteHook() with retry logic |
| Env sanitization | executor.go | SanitizeEnvironment() returns []string |
| Retry logic | system.go | Exponential backoff in retryHook() |

## CONVENTIONS
- **Execution**: Always use `executor.Execute(ctx, command, args, env, eventType)`
- **Retry backoff**: `i²` seconds (i=0,1,2...) implemented in `retryHook()`
- **Env format**: `SanitizeEnvironment()` returns `[]string` as `"key=value"`
- **Exit status**: Parse via `syscall.WaitStatus(status.ExitStatus())`
- **Streaming**: Goroutines for stdout/stderr with `sync.WaitGroup`

## ANTI-PATTERNS (HOOK/ PACKAGE)
- **Shell invocation**: Never use `exec.Command("sh", "-c", ...)` - always validate args
- **Missing timeout**: Always use `context.WithTimeout()` around `executor.Execute()`
- **Empty error checks**: Log all hook execution errors (line 72 in system.go)

## UNIQUE STYLES
- **Structured logging**: `logger.Info("Hook output", "stream", "stdout"/"stderr", "line", line)`
- **Exponential retry**: Backoff formula `time.Duration(i*i) * time.Second` (line 98)
- **Context-aware**: Check `context.DeadlineExceeded` vs `context.Canceled` separately
