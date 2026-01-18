#!/bin/bash
set -e

VIP_CONF="${VIP_CONF:-/etc/vip-switch/vip.conf}"

if [ -f "$VIP_CONF" ]; then
    source "$VIP_CONF"
else
    echo "[ToMaster] ERROR: VIP config not found: $VIP_CONF" >&2
    exit 1
fi

echo "[ToMaster] Event received at $(date)"
echo "[ToMaster] Node ID: $NODE_ID"
echo "[ToMaster] Binding VIP: $VIP_ADDRESS on $INTERFACE"

ip addr replace "$VIP_ADDRESS" dev "$INTERFACE"

for i in $(seq 1 $ARP_COUNT); do
    arping -U -c 1 -I "$INTERFACE" "${VIP_ADDRESS%/*}"
    [ $i -lt $ARP_COUNT ] && sleep 0.${ARP_DELAY_MS}
done

echo "[ToMaster] VIP bound successfully"
