# INTERNAL/STATE/ KNOWLEDGE BASE

**Generated:** 2026-01-18
**Score:** 11 (distinct domain)

## OVERVIEW
State machine for Master/Slave transitions with 2-second debounce and leadership monitoring

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| State constants | machine.go:16-21 | StateReady, StateSlave, StateMaster, StateDestroy (iota) |
| Leadership monitoring | machine.go:90-113 | monitorLeadership: LeaderCh() + ticker (1s) |
| State transitions | machine.go:154-181 | transition(): 2s debounce, hook execution |
| Hook mapping | machine.go:184-201 | executeHookForState: State→EventType mapping |
| Shutdown sequence | machine.go:204-223 | Shutdown(): ToDestroy hook, close channel |

## CONVENTIONS

### Mutex Usage
- `sync.RWMutex` for currentState, previousState, raftNode access
- Read lock: GetCurrentState()
- Write lock: SetRaftNode(), transition(), Shutdown()

### Leadership Events
- Primary: `case isLeader := <-m.raftNode.LeaderCh()` (event-driven)
- Fallback: `case <-ticker.C` (1s periodic check)
- Both paths call transition() on state change

### State Transition Logging
- `"State transition: from -> to"` with state.String()
- Debug log for debounced transitions includes elapsed time

### Debounce Logic
- 2 seconds hardcoded: `time.Since(m.lastStateChange) < m.debounceDelay`
- Exception: `m.currentState != StateReady` (Ready state bypasses debounce)
- Prevents rapid oscillation during Raft leader election

## ANTI-PATTERNS (THIS PACKAGE)
- **Bypassing mutex**: All state access MUST hold lock. transition() calls executeHookForState() while holding lock (acceptable, hookSystem is read-only).
- **Ignoring Raft nil**: monitorLeadership() returns early if raftNode == nil (guard clause).
- **Multiple shutdowns**: Shutdown() uses select with default to check if already closed.
- **Race in debounce**: lastStateChange must be updated BEFORE hook execution to prevent re-entry.

## UNIQUE STYLES

### Raft Integration Pattern
- SetRaftNode() called AFTER Machine creation (raftNode can be nil initially)
- monitorLeadership() starts goroutine that waits for both LeaderCh() events AND ticker
- checkRaftState() queries m.raftNode.Leader() and m.raftNode.IsLeader() separately

### Hook System Integration
- EventType mapping hardcoded: StateReady→"ToReady", StateMaster→"ToMaster", StateSlave→"ToSlave", StateDestroy→"ToDestroy"
- Hook execution happens INSIDE transition() while holding mutex (safe because hookSystem is immutable)
- Hook errors are logged but don't block state change (error propagation only via log)

### Shutdown Safety
- shutdown chan protects against double-close
- ToDestroy hook executes BEFORE closing channel
- Context honored in monitorLeadership loop
