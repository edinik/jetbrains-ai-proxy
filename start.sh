#!/bin/bash

# JetBrains AI Proxy 启动脚本
# 支持自动配置发现和多种配置方式

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 检查可执行文件
check_executable() {
    if [ ! -f "./jetbrains-ai-proxy" ]; then
        print_error "可执行文件 './jetbrains-ai-proxy' 不存在"
        print_info "请先编译项目: go build -o jetbrains-ai-proxy"
        exit 1
    fi
    
    if [ ! -x "./jetbrains-ai-proxy" ]; then
        print_info "设置可执行权限..."
        chmod +x ./jetbrains-ai-proxy
    fi
}

# 检查配置
check_configuration() {
    print_info "检查配置..."
    
    # 检查是否存在配置文件
    config_files=(
        "config.json"
        "config/config.json"
        "configs/config.json"
        ".config/jetbrains-ai-proxy.json"
    )
    
    config_found=false
    for config_file in "${config_files[@]}"; do
        if [ -f "$config_file" ]; then
            print_success "找到配置文件: $config_file"
            config_found=true
            break
        fi
    done
    
    # 检查环境变量
    env_configured=false
    if [ -n "$JWT_TOKENS" ] || [ -n "$JWT_TOKEN" ]; then
        if [ -n "$BEARER_TOKEN" ]; then
            print_success "检测到环境变量配置"
            env_configured=true
        else
            print_warning "检测到JWT tokens但缺少BEARER_TOKEN环境变量"
        fi
    fi
    
    # 检查.env文件
    if [ -f ".env" ]; then
        print_success "找到 .env 文件"
        env_configured=true
    fi
    
    # 如果没有找到任何配置，生成示例配置
    if [ "$config_found" = false ] && [ "$env_configured" = false ]; then
        print_warning "未找到配置文件或环境变量配置"
        print_info "生成示例配置文件..."
        
        if ./jetbrains-ai-proxy --generate-config; then
            print_success "示例配置文件已生成"
            print_info "请编辑 config/config.json 或 .env.example 文件"
            print_info "然后重新运行此脚本"
            exit 0
        else
            print_error "生成示例配置失败"
            exit 1
        fi
    fi
}

# 显示配置信息
show_config() {
    print_info "当前配置信息:"
    ./jetbrains-ai-proxy --print-config
}

# 启动服务
start_service() {
    print_info "启动 JetBrains AI Proxy..."
    
    # 如果有命令行参数，直接传递
    if [ $# -gt 0 ]; then
        print_info "使用命令行参数: $*"
        exec ./jetbrains-ai-proxy "$@"
    else
        # 使用配置文件启动
        exec ./jetbrains-ai-proxy
    fi
}

# 显示帮助信息
show_help() {
    echo "JetBrains AI Proxy 启动脚本"
    echo ""
    echo "用法:"
    echo "  $0                          # 使用配置文件启动"
    echo "  $0 [options]                # 使用命令行参数启动"
    echo "  $0 --help                   # 显示帮助信息"
    echo "  $0 --config                 # 显示当前配置"
    echo "  $0 --generate               # 生成示例配置文件"
    echo ""
    echo "配置方式 (优先级从高到低):"
    echo "  1. 命令行参数"
    echo "  2. 环境变量"
    echo "  3. 配置文件 (config.json, config/config.json 等)"
    echo "  4. 默认值"
    echo ""
    echo "示例:"
    echo "  # 生成配置文件"
    echo "  $0 --generate"
    echo ""
    echo "  # 使用配置文件启动"
    echo "  $0"
    echo ""
    echo "  # 使用命令行参数启动"
    echo "  $0 -c \"jwt1,jwt2,jwt3\" -k \"bearer_token\" -s random"
    echo ""
    echo "  # 使用环境变量启动"
    echo "  export JWT_TOKENS=\"jwt1,jwt2,jwt3\""
    echo "  export BEARER_TOKEN=\"your_token\""
    echo "  $0"
    echo ""
    echo "管理端点:"
    echo "  GET  /health    - 健康检查"
    echo "  GET  /config    - 配置信息"
    echo "  GET  /stats     - 统计信息"
    echo "  POST /reload    - 重载配置"
}

# 主函数
main() {
    echo "🚀 JetBrains AI Proxy 启动脚本"
    echo "================================"
    
    # 处理特殊参数
    case "${1:-}" in
        --help|-h)
            show_help
            exit 0
            ;;
        --config)
            check_executable
            show_config
            exit 0
            ;;
        --generate)
            check_executable
            ./jetbrains-ai-proxy --generate-config
            exit 0
            ;;
    esac
    
    # 检查可执行文件
    check_executable
    
    # 检查配置
    check_configuration
    
    # 启动服务
    start_service "$@"
}

# 捕获中断信号
trap 'print_info "正在停止服务..."; exit 0' INT TERM

# 运行主函数
main "$@"
