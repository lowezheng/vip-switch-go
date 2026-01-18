# PROJECT KNOWLEDGE BASE

**Generated:** 2026-01-18
**Commit:** 8aa628d
**Branch:** main

## OVERVIEW
Raft-based Virtual IP (VIP) failover solution with event-driven hooks. Distributed consensus elects master node for high availability.

## STRUCTURE
```
vip-switch-go/
├── cmd/vip-switch/          # CLI entry point (cobra)
├── internal/
│   ├── raft/                # HashiCorp Raft consensus layer
│   ├── hook/                # Event-driven hook system
│   ├── state/               # State machine (Ready/Slave/Master/Destroy)
│   └── config/              # YAML configuration
├── scripts/                 # Production hook scripts (real IP ops)
├── test/
│   ├── scripts/             # Test hooks (simulation only)
│   ├── configs/             # Test configurations
│   └── run-test.sh         # Automated test suite (8 cases)
└── build/                   # Compiled binary (gitignored)
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Main entry | `cmd/vip-switch/main.go` | CLI flags, graceful shutdown |
| Raft setup | `internal/raft/node.go` | Node bootstrap, cluster joining |
| Hook execution | `internal/hook/executor.go` | Secure command execution |
| State transitions | `internal/state/machine.go` | Leadership monitoring, debounce |
| Config loading | `internal/config/config.go` | YAML parsing, validation |
| Test logic | `test/run-test.sh` | Log-based verification |

## CODE MAP

### Key Symbols
| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `Node` | struct | internal/raft/node.go | Raft consensus node |
| `FSM` | struct | internal/raft/fsm.go | Raft finite state machine |
| `Transport` | struct | internal/raft/transport.go | TCP layer for Raft |
| `Executor` | struct | internal/hook/executor.go | Command execution |
| `System` | struct | internal/hook/system.go | Hook orchestration |
| `Machine` | struct | internal/state/machine.go | State management |
| `Config` | struct | internal/config/config.go | Configuration |

## CONVENTIONS

### Code Style
- Error handling: `fmt.Errorf("context: %w", err)` (always wrap)
- Logging: `log/slog` (Go 1.21+ structured logging)
- Imports: stdlib first, then third-party
- Struct tags: `yaml:"snake_case"` for YAML config

### Hook System
- Events: `ToReady`, `ToMaster`, `ToSlave`, `ToDestroy`
- Failure strategies: `abort`, `continue`, `retry` (exponential backoff)
- Environment variables: `{{.NodeID}}` template expansion
- Secure execution: Path validation + env sanitization

### State Machine
- States: `Ready` → `Slave` ↔ `Master` → `Destroy`
- Debounce: 2 seconds (prevents rapid oscillation)
- Transitions logged: `"State transition: from -> to"`

### Security
- **NEVER** use `exec.Command("sh", "-c", userInput)`
- **ALWAYS** use `exec.Command(commandName, validatedArgs...)`
- Path whitelist: `/usr/local/bin`, `/usr/bin`, `/bin`, `/usr/sbin`, `/sbin`
- Env whitelist: `EVENT_`, `NODE_`, `VIP_`, `INTERFACE`, `PATH`, `HOME`, `USER`
- Linux capability required: `cap_net_admin+ep`

## ANTI-PATTERNS (THIS PROJECT)
- **Shell injection**: Forbidden. Use `exec.Command()` with validated args.
- **Context cancellation**: Don't ignore context in long-running operations.
- **Empty catch blocks**: Never ignore errors in hook execution.

## UNIQUE STYLES

### Simulation-First Testing
- Test hooks in `test/scripts/*.sh` simulate only (no real IP ops)
- Log markers: `[ToMaster]`, `[ToSlave]`, `[ToReady]`, `[ToDestroy]`
- Verification: Parse log files instead of return values
- Safe CI/CD: No root privileges required

### Configuration Structure
- Node config: `id`, `raft_addr`, `data_dir` (all required)
- Cluster config: `nodes` array (at least one)
- Hook config: `command`, `args`, `timeout`, `on_failure`, `environment`
- Logging: `level` (debug/info/warn/error), `format` (json/text), `output` (file path)

### CLI Flags (cobra)
- Required: `--config` (or `-c`)
- Override: `--node-id`, `--raft-addr`, `--data-dir`, `--log-level`, `--log-format`
- Flag naming: kebab-case

## COMMANDS
```bash
# Build
make build
# Output: build/vip-switch

# Install
make install
# Copies to /usr/local/bin/vip-switch

# Run tests (automated suite)
./test/run-test.sh
# 8 test cases, simulation-based

# Manual testing
./test/manual-test.sh

# Unit tests (Go)
make test
# Note: No *_test.go files exist yet

# Lint
make lint
# Requires: golangci-lint installed
```

## NOTES

### Project Status
- **Production-ready**: Yes (for small clusters, 3-5 nodes)
- **Test coverage**: Integration tests only (no Go unit tests)
- **Documentation**: Comprehensive (README, TEST_PLAN.md, test/README.md)
- **Language**: Code comments English, test docs Chinese

### Known Issues
- Go version in go.mod: `1.25.4` (invalid, should be `1.21` or `1.22`)
- No CI/CD: No GitHub Actions configured
- Empty directory: `pkg/client/` (placeholder)
- Vendored code: `https:/github.com/docker/docker.git/` (26MB, investigate)

### Architecture Notes
- Raft timeout: Heartbeat 1s, Election 1s, Leader lease 500ms
- Hook timeout: Default 60s
- Snapshot interval: 30s, threshold 2 logs
- Performance: Startup <5s, Election <3s, Hook execution <1s, Failover <5s

### Runtime Requirements
- Linux: `cap_net_admin+ep` capability for IP operations
- VIP config: `/etc/vip-switch/vip.conf`
- Hook scripts: `/usr/local/bin/` or configured paths
- Data directory: Per-node (default: `./data/node1/`, etc.)

## internal/hook/

### OVERVIEW
Secure command execution system with streaming output and retry orchestration.

### WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Command execution | `executor.go` | `Execute()` with stdout/stderr streaming |
| Hook orchestration | `system.go` | `ExecuteHook()` with retry logic |
| Env sanitization | `executor.go` | Whitelist filtering in `SanitizeEnvironment()` |
| Retry logic | `system.go` | Exponential backoff in `retryHook()` |

### CONVENTIONS
- Execution: Always use `executor.Execute(ctx, command, args, env, eventType)`
- Retry backoff: `i²` seconds (i=0,1,2...) implemented in `retryHook()`
- Env format: `SanitizeEnvironment()` returns `[]string` as `"key=value"`
- Exit status: Parse via `syscall.WaitStatus(status.ExitStatus())`
- Streaming: Goroutines for stdout/stderr with `sync.WaitGroup`

### ANTI-PATTERNS (hook/ package)
- **Shell invocation**: Never use `exec.Command("sh", "-c", ...)` - always validate args
- **Missing timeout**: Always use `context.WithTimeout()` around `executor.Execute()`
- **Empty error checks**: Log all hook execution errors (line 72 in system.go)

### UNIQUE STYLES
- **Structured logging**: `logger.Info("Hook output", "stream", "stdout"/"stderr", "line", line)`
- **Exponential retry**: Backoff formula `time.Duration(i*i) * time.Second` (line 98)
- **Context-aware**: Check `context.DeadlineExceeded` vs `context.Canceled` separately
## internal/test/

### OVERVIEW
Simulation-based integration testing without actual network IP operations.

### WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Automated tests | `test/run-test.sh` | 8 test cases, log-based verification |
| Test hooks | `test/scripts/on-*-test.sh` | Simulation only, no real IP ops |
| Manual testing | `test/manual-test.sh` | Helper for interactive testing |
| Test configs | `test/configs/*.yaml` | node1/2/3 configurations |
| Log verification | `grep -q "..." test/logs/*.log` | Search for [ToMaster], [ToSlave], etc. |

### CONVENTIONS
- SIMULATING markers: "[ToMaster] SIMULATING VIP BINDING (NO REAL IP OPERATION)"
- Test format: text (not json) for readability
- Log parsing: grep -q "[ToMaster] VIP bound successfully" "$log_file"
- Node lifecycle: Start node, sleep 3-10s, verify state, kill node
- Cleanup: rm -rf ./test/logs/* ./test/data/* between tests
- Performance targets: Startup <5s, Election <3s, Hook <1s, Failover <5s

### ANTI-PATTERNS
- **Real network commands in test hooks**: Forbidden (ip, arping operations)
- **Persistent test data**: Never assume data survives between test cases
- **JSON logs in tests**: Hard to read, always use text format
- **Hardcoded sleep times**: Use configurable waits instead
- **Ignoring log failures**: Always check log file creation

### UNIQUE STYLES
- **Log-based verification**: Parse test/logs/*.log for event markers instead of return values
- **Simulation-only hooks**: test/scripts/on-*-test.sh echo commands without execution
- **No root required**: Safe for CI/CD without Linux capabilities
- **Environment var validation**: Hook scripts print env vars with grep -E prefix
- **Multi-node orchestration**: Spawn nodes as background processes with staggered starts
