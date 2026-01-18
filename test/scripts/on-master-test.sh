#!/bin/bash
set -e

# VIP配置文件路径
VIP_CONF="${VIP_CONF:-./test/configs/vip-test.conf}"

# 读取VIP配置
if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
else
    echo "[ToMaster] ERROR: VIP config not found: $VIP_CONF" >&2
    exit 1
fi

echo "[ToMaster] Event received at $(date)"
echo "[ToMaster] Node ID: $NODE_ID"
echo "[ToMaster] Event Type: $EVENT_TYPE"
echo "[ToMaster] ============================================="
echo "[ToMaster] SIMULATING VIP BINDING (NO REAL IP OPERATION)"
echo "[ToMaster] Binding VIP: $VIP_ADDRESS on $INTERFACE"
echo "[ToMaster] ============================================="

# 模拟绑定VIP的操作（不执行实际命令）
echo "[ToMaster] Executing: ip addr replace $VIP_ADDRESS dev $INTERFACE" >&2
echo "[ToMaster] ✓ VIP bind command executed successfully (simulated)"

# 模拟发送ARP广播
for i in $(seq 1 $ARP_COUNT); do
    echo "[ToMaster] Sending ARP broadcast ($i/$ARP_COUNT): arping -U -c 1 -I $INTERFACE ${VIP_ADDRESS%/*}"
    [ $i -lt $ARP_COUNT ] && sleep 0.${ARP_DELAY_MS}
done

echo "[ToMaster] ============================================="
echo "[ToMaster] VIP bound successfully (simulated)"
echo "[ToMaster] Current state: MASTER with VIP $VIP_ADDRESS"
echo "[ToMaster] ============================================="
echo "[ToMaster] Process info: PID=$$"
echo "[ToMaster] Environment variables:"
env | grep -E "NODE_ID|EVENT_TYPE|VIP_|INTERFACE" | sed 's/^/[ToMaster]   /' || true

exit 0
