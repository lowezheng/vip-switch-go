#!/bin/bash
set -e

echo "[ToReady] Node started at $(date)"
echo "[ToReady] Node ID: $NODE_ID"
echo "[ToReady] Event Type: $EVENT_TYPE"

VIP_CONF="${VIP_CONF:-/etc/vip-switch/vip.conf}"
if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
    echo "[ToReady] VIP Address: $VIP_ADDRESS"
    echo "[ToReady] Interface: $INTERFACE"
fi
