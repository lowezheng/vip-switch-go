#!/bin/bash
set -e

VIP_CONF="${VIP_CONF:-/etc/vip-switch/vip.conf}"

if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
else
    echo "[ToSlave] ERROR: VIP config not found: $VIP_CONF" >&2
    exit 1
fi

echo "[ToSlave] Event received at $(date)"
echo "[ToSlave] Node ID: $NODE_ID"
echo "[ToSlave] Unbinding VIP: $VIP_ADDRESS from $INTERFACE"

ip addr del "$VIP_ADDRESS" dev "$INTERFACE" 2>/dev/null || true

echo "[ToSlave] VIP unbound successfully"
