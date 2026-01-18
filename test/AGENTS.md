## OVERVIEW
Simulation-based integration testing without actual network IP operations.

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Automated tests | test/run-test.sh | 8 test cases, log-based verification |
| Test hooks | test/scripts/on-*-test.sh | Simulation only, no real IP ops |
| Test configs | test/configs/*.yaml | node1/2/3 configurations |
| Log verification | grep -q "..." test/logs/*.log | Search for [ToMaster], [ToSlave], etc. |

## CONVENTIONS
- **SIMULATING markers**: "[ToMaster] SIMULATING VIP BINDING (NO REAL IP OPERATION)"
- **Test format**: text (not json) for readability
- **Log parsing**: `grep -q "[ToMaster] VIP bound successfully" "$log_file"`
- **Node lifecycle**: Start node, sleep 3-10s, verify state, kill node
- **Cleanup**: `rm -rf ./test/logs/* ./test/data/*` between tests
- **Performance targets**: Startup <5s, Election <3s, Hook <1s, Failover <5s

## ANTI-PATTERNS
- **Real network commands in test hooks**: Forbidden (ip, arping operations)
- **Persistent test data**: Never assume data survives between test cases
- **JSON logs in tests**: Hard to read, always use text format
- **Hardcoded sleep times**: Use configurable waits instead
- **Ignoring log failures**: Always check log file creation

## UNIQUE STYLES
- **Log-based verification**: Parse test/logs/*.log for event markers instead of return values
- **Simulation-only hooks**: test/scripts/on-*-test.sh echo commands without execution
- **No root required**: Safe for CI/CD without Linux capabilities
- **Environment var validation**: Hook scripts print env vars with grep -E prefix
- **Multi-node orchestration**: Spawn nodes as background processes with staggered starts
