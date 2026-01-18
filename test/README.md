# VIP-Switch 测试指南

## 📋 概述

本测试方案用于验证 VIP-Switch 系统的核心功能，重点通过日志输出验证，**不执行实际的网络IP操作**。

---

## 🏗️ 测试环境结构

```
test/
├── README.md              # 本文档
├── run-test.sh           # 自动化测试脚本
├── manual-test.sh        # 手动测试辅助脚本
├── scripts/              # 测试用Hook脚本（只做日志输出）
│   ├── on-master-test.sh    # Master事件：模拟绑定VIP
│   ├── on-slave-test.sh     # Slave事件：模拟解绑VIP
│   ├── on-ready-test.sh     # Ready事件：初始化日志
│   └── on-destroy-test.sh   # Destroy事件：清理日志
├── configs/              # 测试配置文件
│   ├── vip-test.conf        # VIP配置（测试用）
│   ├── node1-config.yaml    # 节点1配置
│   ├── node2-config.yaml    # 节点2配置
│   └── node3-config.yaml    # 节点3配置
├── logs/                 # 测试日志输出（运行时生成）
│   ├── node1.log
│   ├── node2.log
│   └── node3.log
└── data/                 # Raft数据存储（运行时生成）
    ├── node1/
    ├── node2/
    └── node3/
```

---

## 🚀 快速开始

### 前置条件

1. 确保已编译项目：
```bash
make build
```

2. 确保二进制文件存在：
```bash
ls -l ./build/vip-switch
```

---

## 📝 测试方式

### 方式1：自动化测试（推荐）

运行完整的自动化测试套件：

```bash
# 运行所有测试
./test/run-test.sh
```

**测试用例包括：**
1. ✅ 单节点启动测试
2. ✅ 单节点选举测试
3. ✅ 双节点集群测试
4. ✅ 三节点集群测试
5. ✅ 故障转移测试
6. ✅ 环境变量测试
7. ✅ VIP配置读取测试
8. ✅ 优雅关闭测试

**预期输出：**
```
╔═════════════════════════════════════════════════════╗
║        VIP-Switch Automated Test Suite            ║
╚═════════════════════════════════════════════════════╝

[INFO] Checking if binary exists...
[SUCCESS] Binary found: ./build/vip-switch
[INFO] Cleaning test environment...
[SUCCESS] Test environment cleaned

==================================================
TEST 1: Single Node Startup
==================================================
[INFO] Running test: Single Node Startup
...
[SUCCESS] ✓ TEST PASSED: Single Node Startup

==================================================
TEST REPORT
==================================================
Total Tests:  8
Passed:       8
Failed:       0
Pass Rate:    100%
==================================================
[SUCCESS] All tests passed! 🎉
```

---

### 方式2：手动测试

使用手动测试辅助脚本：

```bash
# 启动手动测试环境
./test/manual-test.sh
```

该脚本会：
- 自动构建项目（如需要）
- 清理测试环境
- 显示快速测试命令参考

然后在其他终端窗口中执行：

```bash
# 终端1：启动node1
./build/vip-switch --config ./test/configs/node1-config.yaml

# 终端2：启动node2
./build/vip-switch --config ./test/configs/node2-config.yaml

# 终端3：启动node3
./build/vip-switch --config ./test/configs/node3-config.yaml

# 终端4：查看日志
tail -f ./test/logs/node1.log
tail -f ./test/logs/node2.log
tail -f ./test/logs/node3.log
```

---

## 🧪 手动测试场景

### 场景1：单节点启动

```bash
# 1. 启动单节点
./build/vip-switch --config ./test/configs/node1-config.yaml

# 2. 等待5-10秒

# 3. 查看日志
cat ./test/logs/node1.log

# 预期看到：
# [ToReady] Node started at ...
# [ToReady] Node ID: node1
# [ToMaster] VIP bound successfully (simulated)
```

---

### 场景2：双节点集群

```bash
# 1. 启动node1
./build/vip-switch --config ./test/configs/node1-config.yaml &

# 2. 等待5秒

# 3. 启动node2
./build/vip-switch --config ./test/configs/node2-config.yaml &

# 4. 等待10秒

# 5. 查看日志
grep -E "\[ToMaster\]|\[ToSlave\]" ./test/logs/*.log

# 预期看到：
# 一个节点执行了 [ToMaster]
# 另一个节点执行了 [ToSlave]
```

---

### 场景3：三节点集群

```bash
# 1. 依次启动3个节点
./build/vip-switch --config ./test/configs/node1-config.yaml &
sleep 3

./build/vip-switch --config ./test/configs/node2-config.yaml &
sleep 3

./build/vip-switch --config ./test/configs/node3-config.yaml &

# 2. 等待10秒

# 3. 查看集群状态
grep "\[ToMaster\]" ./test/logs/*.log
grep "\[ToSlave\]" ./test/logs/*.log

# 预期看到：
# 恰好1个节点执行了 [ToMaster]
# 2个节点执行了 [ToSlave]
```

---

### 场景4：故障转移

```bash
# 1. 启动3个节点（如场景3）

# 2. 找到当前的Leader
grep -l "\[ToMaster\]" ./test/logs/*.log

# 假设是 node1，杀掉它
pkill -f "node1-config.yaml"

# 3. 等待10秒

# 4. 检查新的Leader
grep "\[ToMaster\]" ./test/logs/node2.log ./test/logs/node3.log

# 预期看到：
# node2 或 node3 执行了 [ToMaster]
# 新 Leader 产生了
```

---

### 场景5：优雅关闭

```bash
# 1. 启动一个节点
./build/vip-switch --config ./test/configs/node1-config.yaml &
NODE_PID=$!

# 2. 等待5秒

# 3. 发送SIGTERM信号
kill -TERM $NODE_PID

# 4. 查看日志
cat ./test/logs/node1.log | grep -A 10 "\[ToDestroy\]"

# 预期看到：
# [ToDestroy] Node shutting down at ...
# [ToDestroy] Shutdown complete
```

---

## 🔍 日志分析

### 查看实时日志

```bash
# 查看某个节点的实时日志
tail -f ./test/logs/node1.log

# 查看所有节点的日志
tail -f ./test/logs/*.log

# 查看最后50行
tail -50 ./test/logs/node1.log
```

### 搜索特定事件

```bash
# 搜索所有 ToMaster 事件
grep "\[ToMaster\]" ./test/logs/*.log

# 搜索所有 ToSlave 事件
grep "\[ToSlave\]" ./test/logs/*.log

# 搜索所有 ToReady 事件
grep "\[ToReady\]" ./test/logs/*.log

# 搜索所有 ToDestroy 事件
grep "\[ToDestroy\]" ./test/logs/*.log

# 搜索状态转换
grep "State transition" ./test/logs/*.log
```

### 统计事件

```bash
# 统计每个节点的Master事件数
for log in ./test/logs/*.log; do
    echo "$log: $(grep -c '\[ToMaster\]' $log)"
done

# 统计每个节点的Slave事件数
for log in ./test/logs/*.log; do
    echo "$log: $(grep -c '\[ToSlave\]' $log)"
done
```

---

## 📊 测试Hook脚本说明

### 特点

所有测试Hook脚本都遵循以下原则：

1. ✅ **只做日志输出** - 不执行任何实际的网络IP操作
2. ✅ **模拟实际流程** - 输出模拟操作的日志信息
3. ✅ **输出详细信息** - 包含时间戳、节点ID、环境变量等
4. ✅ **环境变量验证** - 打印接收到的环境变量
5. ✅ **清晰的分隔线** - 使用 `=====` 分隔不同部分

### on-master-test.sh

**功能**：模拟Master节点绑定VIP

**输出示例**：
```
[ToMaster] Event received at Sat Jan 18 10:30:00 CST 2026
[ToMaster] Node ID: node1
[ToMaster] Event Type: ToMaster
[ToMaster] =============================================
[ToMaster] SIMULATING VIP BINDING (NO REAL IP OPERATION)
[ToMaster] Binding VIP: 192.168.1.100/32 on eth0
[ToMaster] =============================================
[ToMaster] Executing: ip addr replace 192.168.1.100/32 dev eth0
[ToMaster] ✓ VIP bind command executed successfully (simulated)
[ToMaster] Sending ARP broadcast (1/3): arping -U -c 1 -I eth0 192.168.1.100
[ToMaster] Sending ARP broadcast (2/3): arping -U -c 1 -I eth0 192.168.1.100
[ToMaster] Sending ARP broadcast (3/3): arping -U -c 1 -I eth0 192.168.1.100
[ToMaster] =============================================
[ToMaster] VIP bound successfully (simulated)
[ToMaster] Current state: MASTER with VIP 192.168.1.100/32
[ToMaster] =============================================
```

### on-slave-test.sh

**功能**：模拟Slave节点解绑VIP

**输出示例**：
```
[ToSlave] Event received at Sat Jan 18 10:30:00 CST 2026
[ToSlave] Node ID: node2
[ToSlave] Event Type: ToSlave
[ToSlave] =============================================
[ToSlave] SIMULATING VIP UNBINDING (NO REAL IP OPERATION)
[ToSlave] Unbinding VIP: 192.168.1.100/32 from eth0
[ToSlave] =============================================
[ToSlave] Executing: ip addr del 192.168.1.100/32 dev eth0
[ToSlave] ✓ VIP unbind command executed successfully (simulated)
[ToSlave] Note: Command would have failed if VIP didn't exist (ignored)
[ToSlave] =============================================
[ToSlave] VIP unbound successfully (simulated)
[ToSlave] Current state: SLAVE (VIP unbound)
[ToSlave] =============================================
```

### on-ready-test.sh

**功能**：节点启动初始化

**输出示例**：
```
[ToReady] =============================================
[ToReady] Node started at Sat Jan 18 10:30:00 CST 2026
[ToReady] Node ID: node1
[ToReady] Event Type: ToReady
[ToReady] =============================================
[ToReady] VIP Configuration:
[ToReady]   VIP Address: 192.168.1.100/32
[ToReady]   Interface: eth0
[ToReady]   ARP Count: 3
[ToReady]   ARP Delay: 200ms
[ToReady] =============================================
[ToReady] Process info: PID=12345
[ToReady] Working directory: /path/to/vip-switch-go
[ToReady] Environment variables:
[ToReady]   NODE_ID=node1
[ToReady]   EVENT_TYPE=ToReady
[ToReady] =============================================
[ToReady] Initialization complete, waiting for Raft election...
[ToReady] =============================================
```

### on-destroy-test.sh

**功能**：节点关闭清理

**输出示例**：
```
[ToDestroy] =============================================
[ToDestroy] Node shutting down at Sat Jan 18 10:30:00 CST 2026
[ToDestroy] Node ID: node1
[ToDestroy] Event Type: ToDestroy
[ToDestroy] =============================================
[ToDestroy] Cleaning up VIP configuration:
[ToDestroy]   VIP Address: 192.168.1.100/32
[ToDestroy]   Interface: eth0
[ToDestroy] Executing: ip addr del 192.168.1.100/32 dev eth0
[ToDestroy] ✓ VIP cleanup executed successfully (simulated)
[ToDestroy] Cleaned up VIP
[ToDestroy] =============================================
[ToDestroy] Shutdown complete
[ToDestroy] =============================================
```

---

## 🛠️ 故障排查

### 问题1：测试失败 - Binary not found

**症状**：
```
[ERROR] Binary not found. Please run: make build
```

**解决方案**：
```bash
make build
```

---

### 问题2：端口被占用

**症状**：
```
Error: bind: address already in use
```

**解决方案**：
```bash
# 清理所有测试进程
pkill -f "vip-switch"

# 或者修改配置文件中的端口
# 编辑 test/configs/nodeX-config.yaml
# 修改 raft_addr 端口号
```

---

### 问题3：Hook脚本权限错误

**症状**：
```
Permission denied: ./test/scripts/on-master-test.sh
```

**解决方案**：
```bash
chmod +x ./test/scripts/*.sh
```

---

### 问题4：日志文件未生成

**症状**：
```
[ERROR] Log file not found: ./test/logs/node1.log
```

**解决方案**：
```bash
# 确保日志目录存在
mkdir -p ./test/logs

# 检查配置文件中的日志路径
# 应该是: output: "./test/logs/node1.log"
```

---

### 问题5：Raft选举超时

**症状**：
```
Waiting for Raft election... (长时间无响应)
```

**解决方案**：
```bash
# 检查防火墙设置
# 确保端口 7946, 7947, 7948 可访问

# 清理Raft数据
rm -rf ./test/data/*

# 重启测试
```

---

## 📈 性能基准

### 预期性能指标

| 指标 | 目标值 | 测量方法 |
|------|--------|----------|
| 节点启动时间 | < 5秒 | 从启动到ToReady完成 |
| 选举收敛时间 | < 3秒 | 从集群形成到Leader确定 |
| Hook执行时间 | < 1秒 | Hook脚本执行完成 |
| 故障转移时间 | < 5秒 | Leader故障到新Leader产生 |
| 日志延迟 | < 100ms | Hook输出到日志的时间 |

### 性能测试

```bash
# 测量启动时间
time ./build/vip-switch --config ./test/configs/node1-config.yaml &
# 查看日志中的 [ToReady] 时间戳

# 测量选举时间
# 启动3个节点，记录启动时间和第一个 [ToMaster] 时间

# 测量故障转移时间
# 杀死Leader，记录时间和新的 [ToMaster] 时间
```

---

## 📝 测试报告模板

```markdown
# VIP-Switch 测试报告

**测试日期**: 2026-01-18
**测试人员**: [你的名字]
**测试版本**: [版本号]

## 测试结果概览
- 总测试用例: 8
- 通过: 8
- 失败: 0
- 通过率: 100%

## 详细测试结果

| 测试用例 | 状态 | 说明 |
|----------|------|------|
| 单节点启动测试 | ✅ | ToReady和ToMaster正确触发 |
| 单节点选举测试 | ✅ | 自动选举为Master |
| 双节点集群测试 | ✅ | 产生1个Master和1个Slave |
| 三节点集群测试 | ✅ | 产生1个Master和2个Slave |
| 故障转移测试 | ✅ | Leader故障后自动选举新Leader |
| 环境变量测试 | ✅ | NODE_ID和EVENT_TYPE正确传递 |
| VIP配置读取测试 | ✅ | 正确读取vip-test.conf |
| 优雅关闭测试 | ✅ | ToDestroy正确执行 |

## 发现的问题
无

## 建议和改进
1. 可以添加更多边界测试用例
2. 可以添加性能监控和指标收集

## 结论
✅ 所有测试通过，系统功能正常
```

---

## 📚 相关文档

- [VIP-Switch README](../README.md) - 项目主文档
- [TEST_PLAN.md](../TEST_PLAN.md) - 详细测试计划
- [实现计划](../.opencode/tasks/task1-implementation-plan.md) - 完整实现计划

---

## 🤝 贡献

如果发现问题或有改进建议，请：
1. 提交 Issue
2. 创建 Pull Request
3. 联系维护者

---

**祝测试顺利！ 🎉**
