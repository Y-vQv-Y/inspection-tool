#!/bin/bash

# 巡检工具示例脚本
# 用法: ./scripts/example.sh [server|k8s|all]

set -e

# 配置
TOOL="./inspection-tool"
OUTPUT_DIR="./reports"
SSH_USER="root"
SSH_PASSWORD="changeme"
SSH_PORT=22

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查工具是否存在
if [ ! -f "$TOOL" ]; then
    print_error "巡检工具不存在,请先编译: make build"
    exit 1
fi

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 获取巡检类型
INSPECTION_TYPE=${1:-all}

case $INSPECTION_TYPE in
    server)
        print_info "执行服务器巡检"
        
        # 服务器列表
        SERVERS=(
            "192.168.1.100"
            "192.168.1.101"
            "192.168.1.102"
        )
        
        for server in "${SERVERS[@]}"; do
            print_info "巡检服务器: $server"
            
            $TOOL server \
                --host "$server" \
                --user "$SSH_USER" \
                --password "$SSH_PASSWORD" \
                --port "$SSH_PORT" \
                --output "$OUTPUT_DIR" \
                --format json \
                --detailed || print_warn "服务器 $server 巡检失败"
        done
        ;;
        
    k8s)
        print_info "执行Kubernetes巡检"
        
        $TOOL k8s \
            --kubeconfig ~/.kube/config \
            --output "$OUTPUT_DIR" \
            --format json \
            --detailed || print_error "K8s巡检失败"
        ;;
        
    all)
        print_info "执行综合巡检"
        
        $TOOL all \
            --kubeconfig ~/.kube/config \
            --ssh-user "$SSH_USER" \
            --ssh-password "$SSH_PASSWORD" \
            --ssh-port "$SSH_PORT" \
            --output "$OUTPUT_DIR" \
            --format json \
            --detailed || print_error "综合巡检失败"
        ;;
        
    *)
        print_error "未知的巡检类型: $INSPECTION_TYPE"
        echo "用法: $0 [server|k8s|all]"
        exit 1
        ;;
esac

print_info "巡检完成,报告已保存到: $OUTPUT_DIR"

# 列出最新的报告
print_info "最新报告:"
ls -lht "$OUTPUT_DIR" | head -n 10

# 统计问题数量(如果安装了jq)
if command -v jq &> /dev/null; then
    print_info "问题统计:"
    
    for report in "$OUTPUT_DIR"/*.json; do
        if [ -f "$report" ]; then
            critical=$(jq '.summary.critical_issues // 0' "$report" 2>/dev/null || echo "0")
            warning=$(jq '.summary.warning_issues // 0' "$report" 2>/dev/null || echo "0")
            
            if [ "$critical" -gt 0 ] || [ "$warning" -gt 0 ]; then
                echo "  $(basename "$report"): 严重=$critical, 警告=$warning"
            fi
        fi
    done
fi

print_info "完成!"
