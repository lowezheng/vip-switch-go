## OVERVIEW
HashiCorp Raft consensus implementation with BoltDB storage, custom logging adapter, and leadership monitoring.

## WHERE TO LOOK
| Task | File | Notes |
|------|------|-------|
| Raft initialization | node.go | Bootstrap, cluster joining |
| State machine | fsm.go | Minimal FSM, Apply returns index |
| Logging integration | transport.go | hclog→slog adapter |
| Leadership changes | node.go:LeaderCh() | Returns <-chan bool |

## CONVENTIONS
- **BoltDB storage**: Always at `filepath.Join(cfg.Node.DataDir, "raft.db")`
- **Bootstrapping**: Check `raft.ErrCantBootstrap` (not an error if already bootstrapped)
- **Logger adapter**: Use `NewRaftLogger(logger)` to bridge hclog and slog
- **FSM Apply**: Return `log.Index` (minimal implementation, actual state via hooks)
- **Cluster joining**: Use `AddVoter(ServerID, ServerAddress, 0, 0)` for peer registration
- **Shutdown**: Use `shutdownLock` to prevent duplicate shutdowns

## ANTI-PATTERNS
- **Bootstrap errors**: Don't treat `ErrCantBootstrap` as fatal (cluster may already exist)
- **FSM complexity**: Avoid complex state machine logic (leadership hooks handle state)
- **Logger bypass**: Don't use direct `log` package; always use structured slog logger
- **Transport reuse**: Never reuse transport instances; create new per node

## UNIQUE STYLES
- **Logger implementation**: Full hclog.Logger interface implementation for slog compatibility
- **FSM minimalism**: Simple `map[string]string` state with mutex-protected access
- **TCP transport**: Custom `logWriter` wraps slog for transport-level logging
- **Error handling**: Graceful handling of bootstrap and AddVoter failures
