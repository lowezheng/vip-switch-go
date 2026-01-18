#!/bin/bash
# VIP-Switch 自动化测试脚本
# 功能：运行所有测试用例，验证系统功能

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 测试统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 清理函数
cleanup() {
    echo -e "${YELLOW}[INFO] Cleaning up processes...${NC}"
    pkill -9 -f "vip-switch" 2>/dev/null || true
    sleep 1
    echo -e "${GREEN}[INFO] Cleanup complete${NC}"
}

# 捕获退出信号
trap cleanup EXIT INT TERM

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 检查构建
check_build() {
    log_info "Checking if binary exists..."
    if [ ! -f "./build/vip-switch" ]; then
        log_error "Binary not found. Please run: make build"
        exit 1
    fi
    log_success "Binary found: ./build/vip-switch"
}

# 清理测试环境
clean_test_env() {
    log_info "Cleaning test environment..."
    rm -rf ./test/logs/*
    rm -rf ./test/data/*
    mkdir -p ./test/logs ./test/data
    log_success "Test environment cleaned"
}

# 测试函数
run_test() {
    local test_name=$1
    local test_cmd=$2
    local expected_log=$3
    local log_file=$4
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo ""
    echo -e "${CYAN}==================================================${NC}"
    echo -e "${CYAN}TEST $TOTAL_TESTS: $test_name${NC}"
    echo -e "${CYAN}==================================================${NC}"
    
    log_info "Running test: $test_name"
    log_info "Command: $test_cmd"
    log_info "Expected log: $expected_log"
    
    # 执行测试命令
    eval $test_cmd > /dev/null 2>&1 &
    local pid=$!
    
    # 等待指定时间
    sleep 10
    
    # 检查日志
    if [ -f "$log_file" ]; then
        if grep -q "$expected_log" "$log_file"; then
            PASSED_TESTS=$((PASSED_TESTS + 1))
            log_success "✓ TEST PASSED: $test_name"
            
            # 显示相关日志片段
            echo -e "${BLUE}[LOG SNIPPET]${NC}"
            grep -A 5 "$expected_log" "$log_file" | head -20 || true
        else
            FAILED_TESTS=$((FAILED_TESTS + 1))
            log_error "✗ TEST FAILED: $test_name"
            log_error "Expected log not found: $expected_log"
            
            # 显示日志内容以便调试
            echo -e "${BLUE}[FULL LOG]${NC}"
            cat "$log_file" || true
        fi
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log_error "✗ TEST FAILED: $test_name"
        log_error "Log file not found: $log_file"
    fi
    
    # 清理测试进程
    kill $pid 2>/dev/null || true
    sleep 2
}

# 测试用例1：单节点启动测试
test_1_single_node_startup() {
    run_test "Single Node Startup" \
        "./build/vip-switch --config ./test/configs/node1-config.yaml" \
        "\[ToReady\] Node started" \
        "./test/logs/node1.log"
}

# 测试用例2：单节点选举测试
test_2_single_node_election() {
    run_test "Single Node Election" \
        "./build/vip-switch --config ./test/configs/node1-config.yaml" \
        "\[ToMaster\] VIP bound successfully" \
        "./test/logs/node1.log"
}

# 测试用例3：双节点集群测试
test_3_two_node_cluster() {
    log_info "Starting node1..."
    ./build/vip-switch --config ./test/configs/node1-config.yaml > /dev/null 2>&1 &
    local node1_pid=$!
    
    sleep 5
    
    log_info "Starting node2..."
    ./build/vip-switch --config ./test/configs/node2-config.yaml > /dev/null 2>&1 &
    local node2_pid=$!
    
    sleep 10
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo -e "${CYAN}==================================================${NC}"
    echo -e "${CYAN}TEST $TOTAL_TESTS: Two Node Cluster${NC}"
    echo -e "${CYAN}==================================================${NC}"
    
    # 检查是否有节点成为Master
    if grep -q "\[ToMaster\]" ./test/logs/node1.log || grep -q "\[ToMaster\]" ./test/logs/node2.log; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        log_success "✓ TEST PASSED: Two Node Cluster"
        echo -e "${BLUE}[node1.log]${NC}"
        grep -E "\[ToMaster\]|\[ToSlave\]" ./test/logs/node1.log || true
        echo -e "${BLUE}[node2.log]${NC}"
        grep -E "\[ToMaster\]|\[ToSlave\]" ./test/logs/node2.log || true
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log_error "✗ TEST FAILED: Two Node Cluster"
    fi
    
    kill $node1_pid $node2_pid 2>/dev/null || true
    sleep 2
}

# 测试用例4：三节点集群测试
test_4_three_node_cluster() {
    log_info "Starting node1..."
    ./build/vip-switch --config ./test/configs/node1-config.yaml > /dev/null 2>&1 &
    local node1_pid=$!
    
    sleep 3
    
    log_info "Starting node2..."
    ./build/vip-switch --config ./test/configs/node2-config.yaml > /dev/null 2>&1 &
    local node2_pid=$!
    
    sleep 3
    
    log_info "Starting node3..."
    ./build/vip-switch --config ./test/configs/node3-config.yaml > /dev/null 2>&1 &
    local node3_pid=$!
    
    sleep 10
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo -e "${CYAN}==================================================${NC}"
    echo -e "${CYAN}TEST $TOTAL_TESTS: Three Node Cluster${NC}"
    echo -e "${CYAN}==================================================${NC}"
    
    # 统计Master数量
    local master_count=$(grep -c "\[ToMaster\] VIP bound successfully" ./test/logs/*.log || true)
    
    if [ "$master_count" -eq 1 ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        log_success "✓ TEST PASSED: Three Node Cluster (Exactly 1 Master)"
        
        echo -e "${BLUE}[Cluster State]${NC}"
        echo "  Master: $(grep -l "\[ToMaster\] VIP bound successfully" ./test/logs/*.log | xargs basename)"
        echo "  Slaves: $(grep -l "\[ToSlave\] VIP unbound successfully" ./test/logs/*.log | xargs basename)"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log_error "✗ TEST FAILED: Three Node Cluster (Expected 1 Master, got $master_count)"
    fi
    
    kill $node1_pid $node2_pid $node3_pid 2>/dev/null || true
    sleep 2
}

# 测试用例5：故障转移测试
test_5_failover() {
    log_info "Starting 3 nodes..."
    ./build/vip-switch --config ./test/configs/node1-config.yaml > /dev/null 2>&1 &
    local node1_pid=$!
    sleep 2
    
    ./build/vip-switch --config ./test/configs/node2-config.yaml > /dev/null 2>&1 &
    local node2_pid=$!
    sleep 2
    
    ./build/vip-switch --config ./test/configs/node3-config.yaml > /dev/null 2>&1 &
    local node3_pid=$!
    
    sleep 10
    
    # 找到Leader节点
    local leader_log=$(grep -l "\[ToMaster\] VIP bound successfully" ./test/logs/*.log | head -1)
    local leader_name=$(basename "$leader_log" .log)
    local leader_pid_var="${leader_name}_pid"
    
    log_info "Current Leader: $leader_name"
    log_info "Killing leader to trigger failover..."
    
    # 杀死Leader
    kill ${!leader_pid_var} 2>/dev/null || true
    
    # 清空日志
    > ./test/logs/node2.log
    > ./test/logs/node3.log
    
    sleep 10
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo -e "${CYAN}==================================================${NC}"
    echo -e "${CYAN}TEST $TOTAL_TESTS: Failover Test${NC}"
    echo -e "${CYAN}==================================================${NC}"
    
    # 检查是否产生了新的Leader
    local new_master_count=$(grep -c "\[ToMaster\] VIP bound successfully" ./test/logs/node2.log ./test/logs/node3.log || true)
    
    if [ "$new_master_count" -eq 1 ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        log_success "✓ TEST PASSED: Failover (New Leader elected)"
        
        echo -e "${BLUE}[New Leader]${NC}"
        grep -l "\[ToMaster\] VIP bound successfully" ./test/logs/*.log | xargs basename
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log_error "✗ TEST FAILED: Failover (Expected 1 new Master, got $new_master_count)"
    fi
    
    kill $node2_pid $node3_pid 2>/dev/null || true
    sleep 2
}

# 测试用例6：环境变量测试
test_6_environment_variables() {
    run_test "Environment Variables" \
        "./build/vip-switch --config ./test/configs/node1-config.yaml" \
        "NODE_ID=node1" \
        "./test/logs/node1.log"
}

# 测试用例7：VIP配置读取测试
test_7_vip_config() {
    run_test "VIP Configuration" \
        "./build/vip-switch --config ./test/configs/node1-config.yaml" \
        "VIP Address: 192.168.1.100/32" \
        "./test/logs/node1.log"
}

# 测试用例8：优雅关闭测试
test_8_graceful_shutdown() {
    log_info "Starting node..."
    ./build/vip-switch --config ./test/configs/node1-config.yaml > /dev/null 2>&1 &
    local node_pid=$!
    
    sleep 5
    
    log_info "Sending SIGTERM for graceful shutdown..."
    kill -TERM $node_pid
    
    sleep 5
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo -e "${CYAN}==================================================${NC}"
    echo -e "${CYAN}TEST $TOTAL_TESTS: Graceful Shutdown${NC}"
    echo -e "${CYAN}==================================================${NC}"
    
    if grep -q "\[ToDestroy\] Shutdown complete" ./test/logs/node1.log; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        log_success "✓ TEST PASSED: Graceful Shutdown"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log_error "✗ TEST FAILED: Graceful Shutdown"
        cat ./test/logs/node1.log || true
    fi
}

# 打印测试报告
print_report() {
    echo ""
    echo -e "${CYAN}==================================================${NC}"
    echo -e "${CYAN}TEST REPORT${NC}"
    echo -e "${CYAN}==================================================${NC}"
    echo -e "Total Tests:  $TOTAL_TESTS"
    echo -e "${GREEN}Passed:       $PASSED_TESTS${NC}"
    echo -e "${RED}Failed:       $FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        local pass_rate=100
    else
        local pass_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    fi
    
    echo -e "Pass Rate:    ${pass_rate}%"
    echo -e "${CYAN}==================================================${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        log_success "All tests passed! 🎉"
        return 0
    else
        log_error "Some tests failed!"
        return 1
    fi
}

# 主函数
main() {
    echo ""
    echo -e "${CYAN}╔═════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║        VIP-Switch Automated Test Suite            ║${NC}"
    echo -e "${CYAN}╚═════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    check_build
    clean_test_env
    
    log_info "Starting test suite..."
    echo ""
    
    # 运行测试用例
    test_1_single_node_startup
    clean_test_env
    
    test_2_single_node_election
    clean_test_env
    
    test_3_two_node_cluster
    clean_test_env
    
    test_4_three_node_cluster
    clean_test_env
    
    test_5_failover
    clean_test_env
    
    test_6_environment_variables
    clean_test_env
    
    test_7_vip_config
    clean_test_env
    
    test_8_graceful_shutdown
    clean_test_env
    
    # 打印报告
    print_report
}

# 执行主函数
main "$@"
