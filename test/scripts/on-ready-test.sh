#!/bin/bash
set -e

echo "[ToReady] ============================================="
echo "[ToReady] Node started at $(date)"
echo "[ToReady] Node ID: $NODE_ID"
echo "[ToReady] Event Type: $EVENT_TYPE"
echo "[ToReady] ============================================="

# 可选：读取VIP配置并打印
VIP_CONF="${VIP_CONF:-./test/configs/vip-test.conf}"
if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
    echo "[ToReady] VIP Configuration:"
    echo "[ToReady]   VIP Address: $VIP_ADDRESS"
    echo "[ToReady]   Interface: $INTERFACE"
    echo "[ToReady]   ARP Count: $ARP_COUNT"
    echo "[ToReady]   ARP Delay: ${ARP_DELAY_MS}ms"
else
    echo "[ToReady] WARNING: VIP config not found: $VIP_CONF"
fi

echo "[ToReady] ============================================="
echo "[ToReady] Process info: PID=$$"
echo "[ToReady] Working directory: $(pwd)"
echo "[ToReady] Environment variables:"
env | grep -E "NODE_ID|EVENT_TYPE|VIP_|INTERFACE" | sed 's/^/[ToReady]   /' || true
echo "[ToReady] ============================================="
echo "[ToReady] Initialization complete, waiting for Raft election..."
echo "[ToReady] ============================================="

exit 0
