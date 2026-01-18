#!/bin/bash
# VIP-Switch 手动测试脚本
# 用于快速手动测试各个功能

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 清理函数
cleanup() {
    echo -e "${YELLOW}[INFO] Cleaning up...${NC}"
    pkill -f "vip-switch" 2>/dev/null || true
    sleep 2
}

trap cleanup EXIT

echo -e "${BLUE}=== VIP-Switch Manual Test ===${NC}"
echo ""

# 检查构建
if [ ! -f "./build/vip-switch" ]; then
    echo -e "${YELLOW}[WARNING] Binary not found. Building...${NC}"
    make build
fi

# 清理测试环境
rm -rf ./test/logs/*
rm -rf ./test/data/*
mkdir -p ./test/logs ./test/data

echo -e "${GREEN}[INFO] Test environment prepared${NC}"
echo ""
echo -e "${BLUE}=== Quick Test Commands ===${NC}"
echo ""
echo "1. Start single node (node1):"
echo "   ./build/vip-switch --config ./test/configs/node1-config.yaml"
echo ""
echo "2. Start all 3 nodes (in separate terminals):"
echo "   Terminal 1: ./build/vip-switch --config ./test/configs/node1-config.yaml"
echo "   Terminal 2: ./build/vip-switch --config ./test/configs/node2-config.yaml"
echo "   Terminal 3: ./build/vip-switch --config ./test/configs/node3-config.yaml"
echo ""
echo "3. View logs in real-time:"
echo "   tail -f ./test/logs/node1.log"
echo "   tail -f ./test/logs/node2.log"
echo "   tail -f ./test/logs/node3.log"
echo ""
echo "4. Test failover (kill the leader):"
echo "   Find the leader: grep -l \"\[ToMaster\]\" ./test/logs/*.log"
echo "   Kill the leader: pkill -f \"<leader-config>\""
echo ""
echo "5. Search for specific events:"
echo "   All ToMaster events: grep \"\[ToMaster\]\" ./test/logs/*.log"
echo "   All ToSlave events: grep \"\[ToSlave\]\" ./test/logs/*.log"
echo "   State transitions: grep \"State transition\" ./test/logs/*.log"
echo ""
echo -e "${BLUE}=== Ready to test! ===${NC}"
echo ""
echo -e "${GREEN}[TIP] Use Ctrl+C to stop this script${NC}"

# 保持运行
tail -f /dev/null
