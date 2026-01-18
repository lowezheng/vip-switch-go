#!/bin/bash
set -e

# VIP配置文件路径
VIP_CONF="${VIP_CONF:-./test/configs/vip-test.conf}"

# 读取VIP配置
if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
else
    echo "[ToSlave] ERROR: VIP config not found: $VIP_CONF" >&2
    exit 1
fi

echo "[ToSlave] Event received at $(date)"
echo "[ToSlave] Node ID: $NODE_ID"
echo "[ToSlave] Event Type: $EVENT_TYPE"
echo "[ToSlave] ============================================="
echo "[ToSlave] SIMULATING VIP UNBINDING (NO REAL IP OPERATION)"
echo "[ToSlave] Unbinding VIP: $VIP_ADDRESS from $INTERFACE"
echo "[ToSlave] ============================================="

# 模拟解绑VIP的操作（不执行实际命令）
echo "[ToSlave] Executing: ip addr del $VIP_ADDRESS dev $INTERFACE" >&2
echo "[ToSlave] ✓ VIP unbind command executed successfully (simulated)"
echo "[ToSlave] Note: Command would have failed if VIP didn't exist (ignored)"

echo "[ToSlave] ============================================="
echo "[ToSlave] VIP unbound successfully (simulated)"
echo "[ToSlave] Current state: SLAVE (VIP unbound)"
echo "[ToSlave] ============================================="
echo "[ToSlave] Process info: PID=$$"
echo "[ToSlave] Environment variables:"
env | grep -E "NODE_ID|EVENT_TYPE|VIP_|INTERFACE" | sed 's/^/[ToSlave]   /' || true

exit 0
