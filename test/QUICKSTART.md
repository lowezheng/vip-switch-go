# 🚀 VIP-Switch 快速测试指南

## 5分钟快速开始

### 步骤1：构建项目

```bash
cd /Users/lowezheng/lowe/workspace/vip-switch-go
make build
```

### 步骤2：运行自动化测试

```bash
./test/run-test.sh
```

这将自动运行8个测试用例，包括：
- ✅ 单节点启动和选举
- ✅ 双节点集群
- ✅ 三节点集群
- ✅ 故障转移
- ✅ 环境变量传递
- ✅ VIP配置读取
- ✅ 优雅关闭

### 步骤3：查看测试结果

测试完成后，你会看到：
```
Total Tests:  8
Passed:       8
Failed:       0
Pass Rate:    100%
[SUCCESS] All tests passed! 🎉
```

---

## 🎯 关键特点

### ✅ 不执行实际网络操作

所有测试Hook脚本只做日志输出：
```
[ToMaster] SIMULATING VIP BINDING (NO REAL IP OPERATION)
[ToMaster] Binding VIP: 192.168.1.100/32 on eth0
[ToMaster] ✓ VIP bind command executed successfully (simulated)
```

**不会真正执行**：
- ❌ `ip addr replace`
- ❌ `ip addr del`
- ❌ `arping -U`

### ✅ 完整的日志验证

通过日志验证所有功能：
- Hook事件触发
- 状态转换
- Raft选举
- 环境变量传递

---

## 📊 测试文件清单

| 文件 | 用途 |
|------|------|
| `test/run-test.sh` | 自动化测试脚本（8个测试用例）|
| `test/manual-test.sh` | 手动测试辅助脚本 |
| `test/README.md` | 完整测试文档 |
| `test/scripts/on-master-test.sh` | Master事件Hook（只做日志）|
| `test/scripts/on-slave-test.sh` | Slave事件Hook（只做日志）|
| `test/scripts/on-ready-test.sh` | Ready事件Hook（只做日志）|
| `test/scripts/on-destroy-test.sh` | Destroy事件Hook（只做日志）|
| `test/configs/vip-test.conf` | VIP配置（测试用）|
| `test/configs/node1-config.yaml` | 节点1配置 |
| `test/configs/node2-config.yaml` | 节点2配置 |
| `test/configs/node3-config.yaml` | 节点3配置 |

---

## 🔧 手动测试示例

### 启动单节点测试

```bash
# 启动node1
./build/vip-switch --config ./test/configs/node1-config.yaml &

# 等待5秒
sleep 5

# 查看日志
cat ./test/logs/node1.log

# 应该看到：
# [ToReady] Node started at ...
# [ToMaster] VIP bound successfully (simulated)
```

### 启动三节点集群

```bash
# 终端1
./build/vip-switch --config ./test/configs/node1-config.yaml

# 终端2
./build/vip-switch --config ./test/configs/node2-config.yaml

# 终端3
./build/vip-switch --config ./test/configs/node3-config.yaml

# 终端4 - 查看日志
tail -f ./test/logs/node1.log
tail -f ./test/logs/node2.log
tail -f ./test/logs/node3.log
```

### 测试故障转移

```bash
# 1. 启动3个节点（如上）

# 2. 找到Leader
grep -l "\[ToMaster\]" ./test/logs/*.log

# 3. 杀死Leader（假设是node1）
pkill -f "node1-config.yaml"

# 4. 等待10秒，查看新Leader
grep "\[ToMaster\]" ./test/logs/node2.log ./test/logs/node3.log
```

---

## 📈 预期日志输出

### ToMaster 事件

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

### ToSlave 事件

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
[ToSlave] =============================================
[ToSlave] VIP unbound successfully (simulated)
[ToSlave] Current state: SLAVE (VIP unbound)
[ToSlave] =============================================
```

---

## 🔍 日志分析命令

### 查看所有Hook事件

```bash
# 所有ToMaster事件
grep "\[ToMaster\]" ./test/logs/*.log

# 所有ToSlave事件
grep "\[ToSlave\]" ./test/logs/*.log

# 所有ToReady事件
grep "\[ToReady\]" ./test/logs/*.log

# 所有ToDestroy事件
grep "\[ToDestroy\]" ./test/logs/*.log
```

### 统计Master/Slave数量

```bash
# 统计Master数量
echo "Master count: $(grep -c '\[ToMaster\]' ./test/logs/*.log || echo 0)"

# 统计Slave数量
echo "Slave count: $(grep -c '\[ToSlave\]' ./test/logs/*.log || echo 0)"
```

### 查找当前Leader

```bash
# 找到执行了ToMaster的节点
grep -l "\[ToMaster\] VIP bound successfully" ./test/logs/*.log
```

---

## 🛠️ 常见问题

### Q: 测试失败怎么办？

A: 检查以下几点：
1. 是否已编译项目：`make build`
2. 端口是否被占用：清理进程 `pkill -f vip-switch`
3. 清理测试数据：`rm -rf ./test/logs/* ./test/data/*`

### Q: 如何只运行特定测试？

A: 手动运行单个测试场景：
```bash
# 只测试单节点
./build/vip-switch --config ./test/configs/node1-config.yaml &
sleep 10
cat ./test/logs/node1.log
```

### Q: 测试Hook脚本真的不会执行网络操作吗？

A: 是的！所有测试Hook脚本只是输出日志，注释中明确标注"SIMULATING"（模拟）。你可以检查 `test/scripts/*.sh` 文件，没有真实的 `ip` 命令执行。

---

## 📚 完整文档

- **详细测试计划**: `TEST_PLAN.md`
- **测试指南**: `test/README.md`
- **项目README**: `README.md`
- **实现计划**: `.opencode/tasks/task1-implementation-plan.md`

---

## ✨ 开始测试

```bash
# 1. 构建
make build

# 2. 运行自动化测试
./test/run-test.sh

# 3. 查看结果
# 如果所有测试通过，你会看到：
# [SUCCESS] All tests passed! 🎉
```

**祝测试成功！ 🎉**
