#!/bin/bash
set -e

echo "[ToDestroy] Node shutting down at $(date)"
echo "[ToDestroy] Node ID: $NODE_ID"
echo "[ToDestroy] Event Type: $EVENT_TYPE"

VIP_CONF="${VIP_CONF:-/etc/vip-switch/vip.conf}"
if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
    ip addr del "$VIP_ADDRESS" dev "$INTERFACE" 2>/dev/null || true
    echo "[ToDestroy] Cleaned up VIP"
fi

echo "[ToDestroy] Shutdown complete"
