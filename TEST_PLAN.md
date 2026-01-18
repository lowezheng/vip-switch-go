# VIP-Switch 测试方案

## 📌 测试目标

验证 VIP-Switch 系统的核心功能，重点测试：
- ✅ Hook 事件触发机制（ToReady, ToMaster, ToSlave, ToDestroy）
- ✅ 状态机转换（Ready ↔ Slave ↔ Master → Destroy）
- ✅ Raft 领导者选举和故障转移
- ✅ 环境变量传递（NODE_ID, EVENT_TYPE等）
- ✅ 实时日志输出
- ❌ **不实际操作网络接口** - 所有测试仅通过日志验证

---

## 🏗️ 测试环境

### 目录结构
```
test/
├── scripts/              # 测试用Hook脚本（只做日志输出）
│   ├── on-master-test.sh
│   ├── on-slave-test.sh
│   ├── on-ready-test.sh
│   └── on-destroy-test.sh
├── configs/              # 测试配置文件
│   ├── node1-config.yaml
│   ├── node2-config.yaml
│   ├── node3-config.yaml
│   └── vip-test.conf
├── logs/                 # 测试日志输出
│   ├── node1.log
│   ├── node2.log
│   └── node3.log
├── data/                 # Raft数据存储
│   ├── node1/
│   ├── node2/
│   └── node3/
└── run-test.sh          # 自动化测试脚本
```

### 测试配置

#### VIP配置 (vip-test.conf)
```bash
VIP_ADDRESS="192.168.1.100/32"
INTERFACE="eth0"
ARP_COUNT=3
ARP_DELAY_MS=200
```

#### 节点配置 (node1-config.yaml)
```yaml
node:
  id: "node1"
  raft_addr: "127.0.0.1:7946"
  data_dir: "./test/data/node1"

cluster:
  nodes:
    - id: "node1"
      addr: "127.0.0.1:7946"
    - id: "node2"
      addr: "127.0.0.1:7947"
    - id: "node3"
      addr: "127.0.0.1:7948"

hooks:
  enabled: true
  timeout: 60s
  on_failure: "abort"

  ToMaster:
    command: "./test/scripts/on-master-test.sh"
    timeout: 30s
    on_failure: "abort"
    environment:
      EVENT_TYPE: "ToMaster"
      NODE_ID: "{{.NodeID}}"

  ToSlave:
    command: "./test/scripts/on-slave-test.sh"
    timeout: 30s
    on_failure: "abort"
    environment:
      EVENT_TYPE: "ToSlave"
      NODE_ID: "{{.NodeID}}"

  ToReady:
    command: "./test/scripts/on-ready-test.sh"
    timeout: 10s
    on_failure: "continue"
    environment:
      EVENT_TYPE: "ToReady"
      NODE_ID: "{{.NodeID}}"

  ToDestroy:
    command: "./test/scripts/on-destroy-test.sh"
    timeout: 10s
    on_failure: "continue"
    environment:
      EVENT_TYPE: "ToDestroy"
      NODE_ID: "{{.NodeID}}"

logging:
  level: "info"
  format: "text"  # 使用text格式方便查看
  output: "./test/logs/node1.log"
```

---

## 🧪 测试用例

### 测试用例 1: 单节点启动测试

**目的**: 验证节点启动时 ToReady 事件触发

**步骤**:
1. 启动 node1
2. 检查日志中是否包含:
   ```
   [ToReady] Node started at <timestamp>
   [ToReady] Node ID: node1
   [ToReady] Event Type: ToReady
   [ToReady] VIP Address: 192.168.1.100/32
   [ToReady] Interface: eth0
   ```

**预期结果**: ✅ ToReady Hook 成功执行，日志输出正确

---

### 测试用例 2: 单节点选举测试

**目的**: 验证单节点自动成为 Master

**步骤**:
1. 启动 node1（等待5秒）
2. 检查日志中是否包含:
   ```
   State transition: Ready -> Master
   [ToMaster] Event received at <timestamp>
   [ToMaster] Node ID: node1
   [ToMaster] Binding VIP: 192.168.1.100/32 on eth0
   [ToMaster] Sending ARP broadcast (1/3)
   [ToMaster] Sending ARP broadcast (2/3)
   [ToMaster] Sending ARP broadcast (3/3)
   [ToMaster] VIP bound successfully
   ```

**预期结果**: ✅ ToMaster Hook 成功执行，状态转换正确

---

### 测试用例 3: 双节点启动测试

**目的**: 验证两个节点形成集群，产生一个 Leader 和一个 Follower

**步骤**:
1. 启动 node1
2. 等待5秒后启动 node2
3. 检查两个节点的日志

**预期结果**:
- node1: 可能保持 Master 或转变为 Slave
- node2: 成为 Slave（因为node1先启动）
- 某个节点执行 ToMaster Hook
- 另一个节点执行 ToSlave Hook

**日志验证**:
```
node1.log:
[ToMaster] ... 或 [ToSlave] ...

node2.log:
[ToSlave] Node ID: node2
[ToSlave] Unbinding VIP: 192.168.1.100/32 from eth0
[ToSlave] VIP unbound successfully
```

---

### 测试用例 4: 三节点集群测试

**目的**: 验证完整的3节点Raft集群

**步骤**:
1. 按顺序启动 node1, node2, node3（间隔5秒）
2. 观察集群形成过程
3. 确认只有一个 Leader

**预期结果**:
- 一个节点为 Master
- 两个节点为 Slave
- Master 节点执行 ToMaster Hook
- Slave 节点执行 ToSlave Hook

---

### 测试用例 5: 故障转移测试

**目的**: 验证 Leader 故障后的自动选举

**步骤**:
1. 启动3个节点（node1, node2, node3）
2. 等待10秒确认 Leader
3. 杀死 Leader 节点（假设是 node1）
4. 观察剩余2个节点的选举
5. 新 Leader 执行 ToMaster Hook
6. 原 Leader 的其他节点保持 Slave 状态

**预期结果**:
- 新的 Leader 被选举出来（node2 或 node3）
- 新 Leader 执行 ToMaster Hook
- 状态转换 Slave -> Master
- VIP 绑定"切换"（通过日志验证）

---

### 测试用例 6: 优雅关闭测试

**目的**: 验证节点关闭时的 ToDestroy Hook 执行

**步骤**:
1. 启动一个节点
2. 等待5秒
3. 发送 SIGTERM 信号关闭节点

**预期结果**: 日志中包含:
```
[ToDestroy] Node shutting down at <timestamp>
[ToDestroy] Node ID: node1
[ToDestroy] Event Type: ToDestroy
[ToDestroy] Cleaned up VIP
[ToDestroy] Shutdown complete
```

---

### 测试用例 7: 快速切换测试（防抖动）

**目的**: 验证防抖动机制，避免频繁状态切换

**步骤**:
1. 快速杀死并重启 Leader 节点（间隔小于2秒）
2. 观察状态变化

**预期结果**: 
- 状态切换被抑制（防抖动）
- 避免频繁触发 ToMaster/ToSlave Hook
- 日志中包含 debounce 相关信息

---

### 测试用例 8: 环境变量传递测试

**目的**: 验证环境变量正确传递到 Hook 脚本

**步骤**:
1. 启动节点
2. 检查 Hook 脚本接收到的环境变量

**预期结果**: Hook 脚本日志中包含:
```
NODE_ID=node1
EVENT_TYPE=ToMaster/ToSlave/ToReady/ToDestroy
```

---

### 测试用例 9: 超时控制测试

**目的**: 验证 Hook 超时机制

**步骤**:
1. 修改 Hook 脚本，添加 `sleep 35`（超过30秒超时）
2. 触发 Hook 事件
3. 观察超时错误

**预期结果**: 日志中包含超时错误:
```
Hook timed out: context deadline exceeded
```

---

### 测试用例 10: 失败策略测试

**目的**: 验证不同的失败策略（abort/continue/retry）

**步骤**:
1. 修改 Hook 脚本返回非零退出码
2. 分别测试 abort, continue, retry 策略
3. 观察行为差异

**预期结果**:
- **abort**: Hook 失败后中止流程
- **continue**: Hook 失败后继续执行
- **retry**: Hook 失败后重试3次（指数退避）

---

## 📊 测试检查清单

### 功能验证
- [ ] ToReady 事件正确触发
- [ ] ToMaster 事件正确触发
- [ ] ToSlave 事件正确触发
- [ ] ToDestroy 事件正确触发
- [ ] 状态机转换正确（Ready → Master/Slave）
- [ ] Raft 领导者选举成功
- [ ] 故障自动转移
- [ ] 防抖动机制生效
- [ ] 优雅关闭执行 ToDestroy Hook

### 环境变量验证
- [ ] NODE_ID 正确传递
- [ ] EVENT_TYPE 正确传递
- [ ] VIP_ADDRESS 从 vip.conf 正确读取
- [ ] INTERFACE 从 vip.conf 正确读取

### 日志验证
- [ ] Hook stdout 实时输出到日志
- [ ] Hook stderr 实时输出到日志
- [ ] 日志格式正确（text/json）
- [ ] 日志级别正确（debug/info/warn/error）

### 异常处理验证
- [ ] Hook 超时正确处理
- [ ] Hook 失败策略生效
- [ ] 节点故障自动恢复
- [ ] 配置文件缺失报错
- [ ] Hook 脚本缺失报错

---

## 🚀 测试执行

### 自动化测试脚本

```bash
#!/bin/bash
# run-test.sh

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 清理函数
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    pkill -f "vip-switch" || true
    sleep 2
}

# 测试函数
run_test() {
    local test_name=$1
    local test_cmd=$2
    local expected_log=$3
    
    echo -e "${GREEN}Running: $test_name${NC}"
    
    eval $test_cmd > /dev/null 2>&1 &
    local pid=$!
    
    sleep 5
    
    if grep -q "$expected_log" ./test/logs/*.log; then
        echo -e "${GREEN}✓ PASSED: $test_name${NC}"
        return 0
    else
        echo -e "${RED}✗ FAILED: $test_name${NC}"
        return 1
    fi
    
    kill $pid 2>/dev/null || true
}

# 主测试流程
main() {
    trap cleanup EXIT
    
    echo "=== VIP-Switch Test Suite ==="
    echo ""
    
    # 清理旧数据
    rm -rf ./test/logs/*
    rm -rf ./test/data/*
    
    # 运行测试用例
    run_test "Test 1: Single node startup" \
        "build/vip-switch --config ./test/configs/node1-config.yaml" \
        "[ToReady] Node started"
    
    # ... 更多测试用例
    
    echo ""
    echo "=== Test Suite Complete ==="
}

main "$@"
```

### 手动测试步骤

```bash
# 1. 构建项目
make build

# 2. 准备测试环境
mkdir -p test/logs test/data

# 3. 安装测试Hook脚本
chmod +x test/scripts/*.sh

# 4. 测试1：单节点启动
./build/vip-switch --config ./test/configs/node1-config.yaml &
sleep 5
cat test/logs/node1.log | grep -E "\[ToReady\]|\[ToMaster\]"

# 5. 测试2：启动第二个节点
./build/vip-switch --config ./test/configs/node2-config.yaml &
sleep 5
cat test/logs/node2.log | grep -E "\[ToSlave\]"

# 6. 测试3：启动第三个节点
./build/vip-switch --config ./test/configs/node3-config.yaml &
sleep 5

# 7. 测试4：故障转移
pkill -f "node1-config.yaml"
sleep 10
grep -E "\[ToMaster\]|\[ToSlave\]" test/logs/node2.log test/logs/node3.log

# 8. 清理
pkill -f "vip-switch"
```

---

## 📈 性能指标

### 关键指标
| 指标 | 目标值 | 测量方法 |
|------|--------|----------|
| 节点启动时间 | < 5秒 | 从启动到ToReady完成 |
| 选举收敛时间 | < 3秒 | 从集群形成到Leader确定 |
| Hook 执行时间 | < 1秒 | Hook脚本执行完成 |
| 故障转移时间 | < 5秒 | Leader故障到新Leader产生 |
| 日志延迟 | < 100ms | Hook输出到日志的时间 |

---

## 🎯 成功标准

### 基本功能
- ✅ 所有4个Hook事件正确触发
- ✅ 状态机转换符合预期
- ✅ Raft选举和故障转移正常
- ✅ 环境变量正确传递
- ✅ 日志输出完整且实时

### 稳定性
- ✅ 连续运行24小时无崩溃
- ✅ 故障转移成功率 100%
- ✅ 无内存泄漏
- ✅ 无文件描述符泄漏

### 安全性
- ✅ 不会执行实际的IP操作（测试环境）
- ✅ Hook 脚本路径验证
- ✅ 环境变量净化
- ✅ 命令注入防护

---

## 📝 测试报告模板

```markdown
# VIP-Switch 测试报告

**测试日期**: 2026-01-18
**测试人员**: [姓名]
**测试版本**: [版本号]

## 测试结果概览
- 总测试用例: 10
- 通过: X
- 失败: Y
- 通过率: Z%

## 详细测试结果
| 测试用例 | 状态 | 说明 |
|----------|------|------|
| 单节点启动测试 | ✅/❌ | 说明 |
| ... | ... | ... |

## 发现的问题
1. [问题描述]
   - 严重程度: [高/中/低]
   - 复现步骤: ...
   - 预期行为: ...
   - 实际行为: ...

## 建议和改进
- [建议1]
- [建议2]

## 结论
[总体评价]
```

---

## 🔍 调试技巧

### 查看实时日志
```bash
# 查看某个节点的实时日志
tail -f test/logs/node1.log

# 查看所有节点的日志
tail -f test/logs/*.log
```

### 搜索特定事件
```bash
# 搜索所有 ToMaster 事件
grep "\[ToMaster\]" test/logs/*.log

# 搜索状态转换
grep "State transition" test/logs/*.log
```

### 检查Raft状态
```bash
# 查看Raft配置
grep "raft.*configuration" test/logs/*.log

# 查看领导权变化
grep -E "became leader|became follower" test/logs/*.log
```

---

## 📚 参考资料

- [Raft共识算法论文](https://raft.github.io/)
- [HashiCorp Raft文档](https://developer.hashicorp.com/raft)
- [Go testing最佳实践](https://go.dev/doc/tutorial/add-a-test)
