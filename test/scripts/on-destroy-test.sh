#!/bin/bash
set -e

echo "[ToDestroy] ============================================="
echo "[ToDestroy] Node shutting down at $(date)"
echo "[ToDestroy] Node ID: $NODE_ID"
echo "[ToDestroy] Event Type: $EVENT_TYPE"
echo "[ToDestroy] ============================================="

# 可选：解绑VIP（确保资源清理）
VIP_CONF="${VIP_CONF:-./test/configs/vip-test.conf}"
if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
    echo "[ToDestroy] Cleaning up VIP configuration:"
    echo "[ToDestroy]   VIP Address: $VIP_ADDRESS"
    echo "[ToDestroy]   Interface: $INTERFACE"
    
    # 模拟解绑VIP的操作（不执行实际命令）
    echo "[ToDestroy] Executing: ip addr del $VIP_ADDRESS dev $INTERFACE" >&2
    echo "[ToDestroy] ✓ VIP cleanup executed successfully (simulated)"
    echo "[ToDestroy] Cleaned up VIP"
else
    echo "[ToDestroy] WARNING: VIP config not found: $VIP_CONF"
fi

echo "[ToDestroy] ============================================="
echo "[ToDestroy] Process info: PID=$$"
echo "[ToDestroy] Environment variables:"
env | grep -E "NODE_ID|EVENT_TYPE|VIP_|INTERFACE" | sed 's/^/[ToDestroy]   /' || true
echo "[ToDestroy] ============================================="
echo "[ToDestroy] Shutdown complete"
echo "[ToDestroy] ============================================="

exit 0
