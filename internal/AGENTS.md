# INTERNAL PACKAGES

**Generated:** 2026-01-18
**Commit:** 8aa628d
**Branch:** main

## OVERVIEW
Core packages implementing Raft consensus, hook orchestration, state management, and configuration.

## WHERE TO LOOK
| Task | Package | File | Notes |
|------|---------|------|-------|
| Config loading | config | config.go | YAML parsing + template expansion |
| Raft bootstrap | raft | node.go | Node creation, cluster join |
| Log storage | raft | fsm.go | BoltDB persistence, snapshots |
| Transport layer | raft | transport.go | TCP connection pooling |
| Hook execution | hook | executor.go | Path validation, command execution |
| Event handling | hook | system.go | Retry logic, failure strategies |
| State transitions | state | machine.go | Leadership monitoring, debounce |

## CONVENTIONS
- **Logging**: All packages use `log/slog` with structured fields (slog.String, slog.Int)
- **Error wrapping**: Always wrap with context: `fmt.Errorf("operation failed: %w", err)`
- **Context propagation**: First parameter in all methods, checked for cancellation
- **Thread safety**: Shared state protected with `sync.RWMutex` (no global state)
- **Package boundaries**: Clean interfaces, no circular dependencies

## ANTI-PATTERNS
- **Global mutable state**: Each package manages its own internal state, no globals
- **Silent context cancellation**: Always check `ctx.Done()` before long operations
- **Mixed responsibilities**: Config only loads, Raft only consensus, Hook only execution

## UNIQUE STYLES

### Config: Template-Based Expansion
- Go templates: `{{.NodeID}}`, `{{.RaftAddr}}` placeholders
- Two-pass loading: Parse YAML → Expand templates → Validate
- Whitelisted env vars: `VIP_`, `INTERFACE`, `NODE_` prefixes only

### Raft: BoltDB Snapshots
- Snapshot interval: 30s, threshold: 2 logs
- FSM methods: `Apply()` (logs), `Snapshot()` (BoltDB), `Restore()` (recovery)
- Non-blocking snapshots: Separate goroutine to avoid consensus delays

### Hook: Secure Execution
- Path whitelist: `/usr/local/bin`, `/usr/bin`, `/bin`, `/usr/sbin`, `/sbin`
- Args validation: No shell metacharacters, direct `exec.Command()`
- Failure strategies: `abort` (stop), `continue` (next hook), `retry` (exponential backoff)

### State: Debounced Transitions
- Leadership changes buffered for 2s before state change
- Channel-based updates: `stateCh`, `leadershipCh`
- Finalizer: Always call `Destroy` hook on shutdown

## CODE METRICS
- **Total**: 1,222 lines Go
- **Largest package**: raft (~500 lines)
- **Smallest package**: config (~150 lines)
