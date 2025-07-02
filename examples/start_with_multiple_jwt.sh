#!/bin/bash

# 多JWT负载均衡启动示例脚本

echo "=== JetBrains AI Proxy 多JWT负载均衡启动示例 ==="

# 检查是否提供了JWT tokens
if [ -z "$1" ]; then
    echo "用法: $0 \"jwt_token1,jwt_token2,jwt_token3\" [bearer_token] [strategy] [port]"
    echo ""
    echo "参数说明:"
    echo "  jwt_tokens    - 多个JWT tokens，用逗号分隔（必需）"
    echo "  bearer_token  - Bearer token（可选，默认从环境变量读取）"
    echo "  strategy      - 负载均衡策略：round_robin 或 random（可选，默认round_robin）"
    echo "  port          - 监听端口（可选，默认8080）"
    echo ""
    echo "示例:"
    echo "  $0 \"jwt1,jwt2,jwt3\""
    echo "  $0 \"jwt1,jwt2,jwt3\" \"bearer123\" \"random\" 9090"
    echo ""
    echo "环境变量配置示例:"
    echo "  export JWT_TOKENS=\"jwt1,jwt2,jwt3\""
    echo "  export BEARER_TOKEN=\"your_bearer_token\""
    echo "  export LOAD_BALANCE_STRATEGY=\"random\""
    echo "  ./jetbrains-ai-proxy"
    exit 1
fi

# 参数设置
JWT_TOKENS="$1"
BEARER_TOKEN="${2:-$BEARER_TOKEN}"
STRATEGY="${3:-round_robin}"
PORT="${4:-8080}"

# 检查Bearer token
if [ -z "$BEARER_TOKEN" ]; then
    echo "错误: 需要提供Bearer token"
    echo "请通过参数提供或设置环境变量 BEARER_TOKEN"
    exit 1
fi

# 检查策略有效性
if [ "$STRATEGY" != "round_robin" ] && [ "$STRATEGY" != "random" ]; then
    echo "警告: 无效的负载均衡策略 '$STRATEGY'，使用默认策略 'round_robin'"
    STRATEGY="round_robin"
fi

# 计算JWT tokens数量
TOKEN_COUNT=$(echo "$JWT_TOKENS" | tr ',' '\n' | wc -l | tr -d ' ')

echo "配置信息:"
echo "  JWT Tokens数量: $TOKEN_COUNT"
echo "  负载均衡策略: $STRATEGY"
echo "  监听端口: $PORT"
echo "  Bearer Token: ${BEARER_TOKEN:0:10}..."
echo ""

# 启动服务
echo "启动 JetBrains AI Proxy..."
echo "命令: ./jetbrains-ai-proxy -p $PORT -c \"$JWT_TOKENS\" -k \"$BEARER_TOKEN\" -s \"$STRATEGY\""
echo ""

# 检查可执行文件是否存在
if [ ! -f "./jetbrains-ai-proxy" ]; then
    echo "错误: 找不到可执行文件 './jetbrains-ai-proxy'"
    echo "请先编译项目: go build -o jetbrains-ai-proxy"
    exit 1
fi

# 启动服务
exec ./jetbrains-ai-proxy -p "$PORT" -c "$JWT_TOKENS" -k "$BEARER_TOKEN" -s "$STRATEGY"
