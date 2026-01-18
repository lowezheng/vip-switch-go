# VIP-Switch

A Raft-based Virtual IP (VIP) failover solution with event-driven hooks for high availability.

## Overview

VIP-Switch is a distributed virtual IP failover system that uses Raft consensus to elect a master node. The master node binds a VIP (Virtual IP) through configurable hook scripts, while slave nodes ensure the VIP is unbound.

### Key Features

- Raft-based leader election for high availability
- Event-driven hook system for flexible VIP management
- Automatic failover and recovery
- Support for cluster membership changes
- Configurable failure strategies and timeouts
- Secure command execution (no shell injection)
- Real-time hook output streaming

## Architecture

| Component | Responsibility | Does NOT Handle |
|-----------|---------------|------------------|
| **Raft Layer** | Leader election | VIP configuration, network operations |
| **Hook System** | Event triggering, script execution | VIP logic, network operations |
| **Hook Scripts** | VIP binding/unbinding, ARP broadcast | Raft logic, cluster management |

## Quick Start

### Build

```bash
make build
```

### Install

```bash
make install
```

### Configuration

1. Create configuration directory:
```bash
mkdir -p /etc/vip-switch
```

2. Install main configuration:
```bash
cp config/config.yaml /etc/vip-switch/config.yaml
```
Edit and update node ID and Raft address for your node.

3. Install VIP configuration:
```bash
cp examples/vip.conf /etc/vip-switch/vip.conf
```
Edit and update VIP address and interface.

4. Install hook scripts:
```bash
chmod +x scripts/on-*.sh
cp scripts/on-*.sh /usr/local/bin/
```

### Run

```bash
vip-switch --config /etc/vip-switch/config.yaml
```

### Command Line Options

| Flag | Required | Description | Default |
|-------|-----------|-------------|----------|
| `--config` | ✅ | Configuration file path | - |
| `--node-id` | ❌ | Node ID (overrides config) | config.yaml value |
| `--raft-addr` | ❌ | Raft RPC address (overrides config) | config.yaml value |
| `--data-dir` | ❌ | Raft data directory (overrides config) | config.yaml value |
| `--log-level` | ❌ | Log level: debug, info, warn, error | info |
| `--log-format` | ❌ | Log format: json, text | json |

## Configuration

### Main Config (`config.yaml`)

Configures node, cluster, hooks, and logging settings.

### VIP Config (`vip.conf`)

VIP-specific configuration (read by hook scripts):

```bash
VIP_ADDRESS="192.168.1.100/32"
INTERFACE="eth0"
ARP_COUNT=3
ARP_DELAY_MS=200
```

## Hook Events

| Event | Trigger | Hook Script | Failure Strategy |
|--------|----------|--------------|-----------------|
| `ToReady` | Process starts | `on-ready.sh` | continue |
| `ToMaster` | Node becomes master | `on-master.sh` | abort |
| `ToSlave` | Node becomes slave | `on-slave.sh` | abort |
| `ToDestroy` | Process shuts down | `on-destroy.sh` | continue |

### Hook Execution Features

- Timeout control
- Failure strategies: abort, continue, retry (with exponential backoff)
- Environment variable expansion (`{{.NodeID}}`, `{{.Event}}`)
- Secure command execution (no shell injection)
- Real-time log streaming

## Security

### Command Execution

- **Never use** `exec.Command("sh", "-c", userInput)`
- **Always use** `exec.Command(commandName, validatedArgs...)`
- Environment variable sanitization (whitelist allowed prefixes)
- Command path validation (whitelist safe directories)

### Required Linux Capabilities

```bash
sudo setcap cap_net_admin+ep /usr/local/bin/vip-switch
```

## Development

### Build

```bash
make build
```

### Test

```bash
make test
```

### Lint

```bash
make lint
```

## Tech Stack

| Component | Library | License | Purpose |
|-----------|----------|----------|---------|
| **Raft Consensus** | `github.com/hashicorp/raft` | MPL-2.0 | Leader election |
| **Raft Storage** | `github.com/hashicorp/raft-boltdb` | MPL-2.0 | BoltDB log storage |
| **Config Parser** | `gopkg.in/yaml.v3` | Apache-2.0 | YAML configuration |
| **CLI Framework** | `github.com/spf13/cobra` | Apache-2.0 | Command line interface |
| **Logging** | `log/slog` (Go 1.21+) | BSD-3 | Structured logging |

## Project Structure

```
vip-switch-go/
├── cmd/vip-switch/          # Main CLI entry point
├── internal/
│   ├── raft/                # Raft consensus layer
│   ├── hook/                # Hook execution system
│   ├── state/               # State management
│   └── config/              # Configuration
├── config/                  # Configuration templates
├── scripts/                 # Example hook scripts
├── examples/                # Configuration examples
└── go.mod
```

## License

Apache License 2.0

## Contributing

Contributions are welcome! Please ensure:

- Code follows existing patterns
- All tests pass
- Linter passes
- Commit messages are clear
