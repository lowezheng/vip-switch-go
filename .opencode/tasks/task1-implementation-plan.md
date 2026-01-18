# VIP切换工具 实现计划

## 目标

- 实现一个网卡的VIP绑定和解绑工具
- 基于Raft的选择机制，选举出Master节点，该节点将绑定VIP
- 从节点需要解绑VIP

---

## 功能需求

### 1. Raft选举机制

- 实现3个及以上节点的Raft集群选举
- 选举出唯一的Master节点
- 支持节点故障自动重新选举
- 支持集群成员动态管理

### 2. 事件驱动的Hook系统

#### 事件类型

| 事件 | 触发时机 | Hook脚本 | 失败策略 |
|------|----------|----------|----------|
| `ToReady` | 进程启动时 | `on-ready.sh` | continue |
| `ToMaster` | 节点当选为Master | `on-master.sh` | abort |
| `ToSlave` | 节点变为Slave | `on-slave.sh` | abort |
| `ToDestroy` | 进程关闭时 | `on-destroy.sh` | continue |

#### Hook执行特性

- 支持超时控制
- 支持失败策略（abort/continue/retry）
- 支持环境变量传递
- 安全的命令执行（防止shell注入）
- 实时日志输出流

### 3. VIP管理

- **VIP配置由Hook脚本自行管理**，不在主程序配置文件中体现
- Hook脚本通过 `/etc/vip-switch/vip.conf` 读取VIP配置
- Master节点的 `on-master.sh` 负责绑定VIP并广播ARP
- Slave节点的 `on-slave.sh` 负责解绑VIP
- 使用 `ip addr replace` 实现幂等操作

---

## 技术约束

- 语言/框架：Go 1.21+
- 架构模式：无（标准Go项目布局）
- 依赖限制：无

---

## 配置设计

### 1. 主配置文件：`config.yaml`

**职责**：节点配置、集群配置、Hook配置、日志配置

```yaml
# Node Configuration
node:
  id: "node1"                      # Unique node ID
  raft_addr: "192.168.1.10:7946"   # Raft RPC address
  data_dir: "/var/lib/vip-switch"  # Raft data directory

# Cluster Configuration
cluster:
  nodes:
    - id: "node1"
      addr: "192.168.1.10:7946"
    - id: "node2"
      addr: "192.168.1.11:7946"
    - id: "node3"
      addr: "192.168.1.12:7946"

# Hook Configuration
hooks:
  enabled: true
  timeout: 60s                     # Default hook timeout
  on_failure: "abort"              # abort | continue | retry

  ToMaster:
    command: "/usr/local/bin/on-master.sh"
    args: []                        # 无参数，脚本自行获取配置
    timeout: 30s
    on_failure: "abort"
    environment:
      EVENT_TYPE: "ToMaster"
      NODE_ID: "{{.NodeID}}"        # 仅传递节点信息

  ToSlave:
    command: "/usr/local/bin/on-slave.sh"
    args: []
    timeout: 30s
    on_failure: "abort"
    environment:
      EVENT_TYPE: "ToSlave"
      NODE_ID: "{{.NodeID}}"

  ToReady:
    command: "/usr/local/bin/on-ready.sh"
    args: []
    timeout: 10s
    on_failure: "continue"
    environment:
      EVENT_TYPE: "ToReady"
      NODE_ID: "{{.NodeID}}"

  ToDestroy:
    command: "/usr/local/bin/on-destroy.sh"
    args: []
    timeout: 10s
    on_failure: "continue"
    environment:
      EVENT_TYPE: "ToDestroy"
      NODE_ID: "{{.NodeID}}"

# Logging Configuration
logging:
  level: "info"                    # debug | info | warn | error
  format: "json"                   # json | text
  output: "/var/log/vip-switch/daemon.log"
```

### 2. VIP配置文件：`vip.conf`

**职责**：VIP地址、网络接口、ARP配置

**位置**：`/etc/vip-switch/vip.conf`（由Hook脚本读取）

```bash
# VIP Configuration (由脚本读取)
VIP_ADDRESS="192.168.1.100/32"
INTERFACE="eth0"
ARP_COUNT=3
ARP_DELAY_MS=200
```

---

## 命令行参数

```bash
# 基本用法
vip-switch --config /etc/vip-switch/config.yaml

# 完整参数（覆盖配置文件）
vip-switch --config /etc/vip-switch/config.yaml \
           --node-id node1 \
           --raft-addr 192.168.1.10:7946 \
           --data-dir /var/lib/vip-switch \
           --log-level info
```

### 参数说明

| 参数 | 必需 | 说明 | 默认值 |
|------|------|------|--------|
| `--config` | ✅ | 配置文件路径 | - |
| `--node-id` | ❌ | 节点ID | config.yaml中的值 |
| `--raft-addr` | ❌ | Raft RPC地址 | config.yaml中的值 |
| `--data-dir` | ❌ | Raft数据目录 | config.yaml中的值 |
| `--log-level` | ❌ | 日志级别 | info |
| `--log-format` | ❌ | 日志格式 | json |

**注意**：不包含任何VIP相关参数（`--vip-address`, `--interface`等）

---

## 项目结构

```
vip-switch-go/
├── cmd/
│   └── vip-switch/          # Main CLI entry point
│       └── main.go
├── internal/
│   ├── raft/                # Raft consensus layer
│   │   ├── node.go          # Node initialization & lifecycle
│   │   ├── fsm.go           # Finite State Machine implementation
│   │   └── transport.go     # RPC transport (TCP)
│   ├── hook/                # Hook execution system
│   │   ├── system.go        # Hook registry & execution
│   │   └── executor.go      # Shell command runner
│   ├── state/               # State management
│   │   └── machine.go       # State: Ready ↔ Slave ↔ Master → Destroy
│   └── config/              # Global configuration
│       ├── config.go        # Config struct & loading
│       └── template.go      # Template variable expansion
├── pkg/                     # (Optional) Public API for external clients
│   └── client/
├── config/                  # Configuration templates
│   └── config.yaml          # 主配置文件（无VIP配置）
├── scripts/                 # Example hook scripts
│   ├── on-master.sh         # VIP绑定脚本（读取vip.conf）
│   ├── on-slave.sh          # VIP解绑脚本（读取vip.conf）
│   ├── on-ready.sh          # 启动日志脚本
│   └── on-destroy.sh        # 关闭日志脚本
├── examples/
│   ├── config.yaml          # 主配置示例
│   └── vip.conf             # VIP配置示例
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Hook脚本示例

### on-master.sh

```bash
#!/bin/bash
set -e

# VIP配置文件路径
VIP_CONF="${VIP_CONF:-/etc/vip-switch/vip.conf}"

# 读取VIP配置
if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
else
    echo "[ToMaster] ERROR: VIP config not found: $VIP_CONF" >&2
    exit 1
fi

echo "[ToMaster] Event received at $(date)"
echo "[ToMaster] Node ID: $NODE_ID"
echo "[ToMaster] Binding VIP: $VIP_ADDRESS on $INTERFACE"

# 绑定VIP（幂等操作）
ip addr replace "$VIP_ADDRESS" dev "$INTERFACE"

# 发送ARP广播
for i in $(seq 1 $ARP_COUNT); do
    arping -U -c 1 -I "$INTERFACE" "${VIP_ADDRESS%/*}"
    [ $i -lt $ARP_COUNT ] && sleep 0.${ARP_DELAY_MS}
done

echo "[ToMaster] VIP bound successfully"
```

### on-slave.sh

```bash
#!/bin/bash
set -e

# VIP配置文件路径
VIP_CONF="${VIP_CONF:-/etc/vip-switch/vip.conf}"

# 读取VIP配置
if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
else
    echo "[ToSlave] ERROR: VIP config not found: $VIP_CONF" >&2
    exit 1
fi

echo "[ToSlave] Event received at $(date)"
echo "[ToSlave] Node ID: $NODE_ID"
echo "[ToSlave] Unbinding VIP: $VIP_ADDRESS from $INTERFACE"

# 解绑VIP（忽略错误，因为可能已经不存在）
ip addr del "$VIP_ADDRESS" dev "$INTERFACE" 2>/dev/null || true

echo "[ToSlave] VIP unbound successfully"
```

### on-ready.sh

```bash
#!/bin/bash
set -e

echo "[ToReady] Node started at $(date)"
echo "[ToReady] Node ID: $NODE_ID"
echo "[ToReady] Event Type: $EVENT_TYPE"

# 可选：读取VIP配置并打印
VIP_CONF="${VIP_CONF:-/etc/vip-switch/vip.conf}"
if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
    echo "[ToReady] VIP Address: $VIP_ADDRESS"
    echo "[ToReady] Interface: $INTERFACE"
fi
```

### on-destroy.sh

```bash
#!/bin/bash
set -e

echo "[ToDestroy] Node shutting down at $(date)"
echo "[ToDestroy] Node ID: $NODE_ID"
echo "[ToDestroy] Event Type: $EVENT_TYPE"

# 可选：解绑VIP（确保资源清理）
VIP_CONF="${VIP_CONF:-/etc/vip-switch/vip.conf}"
if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
    ip addr del "$VIP_ADDRESS" dev "$INTERFACE" 2>/dev/null || true
    echo "[ToDestroy] Cleaned up VIP"
fi

echo "[ToDestroy] Shutdown complete"
```

---

## 技术栈

| 组件 | 库 | 许可证 | 用途 |
|------|-----|--------|------|
| **Raft共识** | `github.com/hashicorp/raft` | MPL-2.0 | 领导者选举 |
| **Raft存储** | `github.com/hashicorp/raft-boltdb` | MPL-2.0 | BoltDB日志存储 |
| **配置解析** | `gopkg.in/yaml.v3` | Apache-2.0 | YAML配置 |
| **CLI框架** | `github.com/spf13/cobra` | Apache-2.0 | 命令行接口 |
| **日志** | `log/slog` (Go 1.21+) | BSD-3 | 结构化日志 |

---

## 实现阶段

### 阶段1：基础框架（1天）

- [ ] 初始化Go模块及依赖
- [ ] 创建项目目录结构
- [ ] 实现配置加载器（`internal/config/config.go`）
- [ ] 实现CLI框架（Cobra）
- [ ] 添加结构化日志（slog）

### 阶段2：Raft集成（2天）

- [ ] 实现Raft节点初始化（`internal/raft/node.go`）
- [ ] 实现有限状态机（`internal/raft/fsm.go`）
- [ ] 实现TCP传输层（`internal/raft/transport.go`）
- [ ] 实现LeaderCh()领导权观察
- [ ] 实现集群引导逻辑

### 阶段3：Hook系统（2天）

- [ ] 实现Hook注册表（`internal/hook/system.go`）
- [ ] 实现安全命令执行器（`internal/hook/executor.go`）
- [ ] 定义事件类型（ToMaster, ToSlave, ToReady, ToDestroy）
- [ ] 实现YAML Hook配置加载
- [ ] 实现错误处理和失败策略

### 阶段4：状态机（1天）

- [ ] 实现状态转换（`internal/state/machine.go`）
- [ ] 将Hook执行与Raft领导权变化关联
- [ ] 实现防抖逻辑（防止快速切换）
- [ ] 实现优雅关闭

### 阶段5：集成与测试（2天）

- [ ] 端到端集成
- [ ] 多节点集群测试
- [ ] Hook脚本示例（`scripts/`）
- [ ] 集成测试
- [ ] 文档（README.md）

---

## 核心实现细节

### 1. Raft领导权观察

```go
// 使用LeaderCh() + 周期性State()轮询确保健壮性
go func() {
    for {
        select {
        case leader := <-r.LeaderCh():
            // 领导权变化
            if leader == "" {
                // 无领导者（集群失去法定人数）
                handleNoLeader()
            } else if leader == nodeID {
                // 我是领导者
                handleToMaster()
            } else {
                // 其他人是领导者
                handleToSlave()
            }
        }
    }
}()
```

### 2. 安全Hook执行

```go
// ✅ 永远不要使用: exec.Command("sh", "-c", userInput)
// ✅ 始终使用: exec.Command(commandName, validatedArgs...)

func executeHook(ctx context.Context, hook Hook) error {
    cmdCtx, cancel := context.WithTimeout(ctx, hook.Timeout)
    defer cancel()

    cmd := exec.CommandContext(cmdCtx, hook.Command, hook.Args...)
    cmd.Env = sanitizeEnvironment(hook.Environment)

    // 实时输出流
    stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()
    go io.Copy(logger.InfoWriter(hook.ID), stdout)
    go io.Copy(logger.ErrorWriter(hook.ID), stderr)

    if err := cmd.Run(); err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return fmt.Errorf("hook timed out: %w", err)
        }
        return err
    }

    return nil
}
```

### 3. 防抖逻辑

```go
var lastRoleChange time.Time
const debounce = 2 * time.Second

func onRoleChange(newRole string) {
    if time.Since(lastRoleChange) > debounce {
        executeHook(newRole)
        lastRoleChange = time.Now()
    }
}
```

---

## 安全考虑

| 风险 | 缓解措施 |
|------|----------|
| **Shell注入** | 从不使用`sh -c`；验证命令路径；使用白名单 |
| **权限提升** | 降权；以专用用户运行；使用`setcap`代替root |
| **命令注入** | 参数化执行；转义参数；清理环境变量 |
| **Hook劫持** | HMAC签名验证；IP白名单；文件权限 |

### 所需Linux能力

```bash
# IP地址操作
sudo setcap cap_net_admin+ep /usr/local/bin/vip-switch
```

---

## 快速启动

```bash
# 1. 构建
go build -o /usr/local/bin/vip-switch ./cmd/vip-switch

# 2. 配置目录
mkdir -p /etc/vip-switch

# 3. 主配置文件
cp config/config.yaml /etc/vip-switch/config.yaml
# 编辑：修改节点ID和Raft地址

# 4. VIP配置文件
cp examples/vip.conf /etc/vip-switch/vip.conf
# 编辑：修改VIP地址和接口

# 5. 安装Hook脚本
chmod +x scripts/on-*.sh
cp scripts/on-*.sh /usr/local/bin/

# 6. 启动节点
vip-switch --config /etc/vip-switch/config.yaml \
           --node-id node1 \
           --raft-addr 192.168.1.10:7946
```

---

## 设计原则

### ✅ 职责分离清晰

| 组件 | 职责 | 不涉及 |
|------|------|--------|
| **Raft层** | 领导者选举 | VIP配置、网络操作 |
| **Hook系统** | 事件触发、脚本调用 | VIP逻辑、网络操作 |
| **Hook脚本** | VIP绑定/解绑、ARP广播 | Raft逻辑、集群管理 |

### ✅ 灵活性高

- 脚本可以从多种来源读取VIP配置：
  - `/etc/vip-switch/vip.conf`
  - 环境变量（`$VIP_CONF`, `$VIP_ADDRESS`）
  - 数据库
  - 云服务元数据
- 脚本可以实现任意逻辑：
  - 复杂的网络配置
  - 负载均衡器更新
  - 服务发现注册

### ✅ 配置简洁

```yaml
# config.yaml - 只关注节点和集群配置
node:
  id: "node1"
  raft_addr: "192.168.1.10:7946"
cluster:
  nodes: [...]

# vip.conf - 只关注VIP配置
VIP_ADDRESS="192.168.1.100/32"
INTERFACE="eth0"
```

---

## 估算工作量

| 阶段 | 天数 | 复杂度 | 风险 |
|------|------|--------|------|
| 基础框架 | 1 | 低 | 低 |
| Raft集成 | 2 | 高 | 中 |
| Hook系统 | 2 | 中 | 低 |
| 状态机 | 1 | 中 | 中 |
| 集成与测试 | 2 | 高 | 中 |
| **总计** | **8** | **高** | **中** |

---

## 待确认问题

1. **集群大小**：是否支持灵活的集群大小（2-10节点）还是硬编码3节点？
2. **IPv6支持**：是否需要支持IPv6 VIP？
3. **Hook执行模式**：关键Hook（ToMaster, ToSlave）应该是同步还是异步？
4. **配置热加载**：是否需要通过SIGHUP支持运行时重载配置？
5. **监控**：是否需要Prometheus指标用于Hook执行和Raft健康状态？
