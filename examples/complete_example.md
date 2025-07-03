# 完整使用示例

本文档提供了JetBrains AI Proxy多JWT负载均衡系统的完整使用示例。

## 📋 准备工作

### 1. 编译项目

```bash
# 进入项目目录
cd jetbrains-ai-proxy

# 编译项目
go build -o jetbrains-ai-proxy

# 或者使用交叉编译
GOOS=linux GOARCH=amd64 go build -o jetbrains-ai-proxy-linux
```

### 2. 准备JWT Tokens

确保你有有效的JetBrains AI JWT tokens。你可以从以下途径获取：
- JetBrains AI服务控制台
- 现有的JetBrains IDE配置
- API密钥管理界面

## 🎯 使用场景示例

### 场景1: 开发环境快速启动

```bash
# 1. 生成示例配置
./jetbrains-ai-proxy --generate-config

# 2. 编辑配置文件
vim config/config.json

# 3. 启动服务
./jetbrains-ai-proxy
```

配置文件内容：
```json
{
  "jetbrains_tokens": [
    {
      "token": "your_jwt_token_here",
      "name": "Dev_Token",
      "description": "Development JWT token"
    }
  ],
  "bearer_token": "your_bearer_token_here",
  "load_balance_strategy": "round_robin",
  "server_port": 8080,
  "server_host": "127.0.0.1"
}
```

### 场景2: 生产环境多JWT负载均衡

```bash
# 1. 创建生产配置目录
mkdir -p /etc/jetbrains-ai-proxy

# 2. 创建生产配置文件
cat > /etc/jetbrains-ai-proxy/config.json << 'EOF'
{
  "jetbrains_tokens": [
    {
      "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
      "name": "Primary_Production",
      "description": "Primary production JWT token",
      "priority": 1,
      "metadata": {
        "environment": "production",
        "region": "us-east-1",
        "tier": "primary"
      }
    },
    {
      "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
      "name": "Secondary_Production",
      "description": "Secondary production JWT token",
      "priority": 2,
      "metadata": {
        "environment": "production",
        "region": "us-west-2",
        "tier": "secondary"
      }
    },
    {
      "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
      "name": "Backup_Production",
      "description": "Backup production JWT token",
      "priority": 3,
      "metadata": {
        "environment": "production",
        "region": "eu-west-1",
        "tier": "backup"
      }
    }
  ],
  "bearer_token": "prod_bearer_token_here",
  "load_balance_strategy": "random",
  "health_check_interval": "30s",
  "server_port": 8080,
  "server_host": "0.0.0.0"
}
EOF

# 3. 启动服务
./jetbrains-ai-proxy --config /etc/jetbrains-ai-proxy/config.json
```

### 场景3: 使用环境变量配置

```bash
# 1. 设置环境变量
export JWT_TOKENS="jwt1,jwt2,jwt3"
export BEARER_TOKEN="your_bearer_token"
export LOAD_BALANCE_STRATEGY="random"
export SERVER_PORT="9090"

# 2. 启动服务
./jetbrains-ai-proxy

# 或者使用.env文件
cat > .env << 'EOF'
JWT_TOKENS=jwt_token_1,jwt_token_2,jwt_token_3
BEARER_TOKEN=your_bearer_token_here
LOAD_BALANCE_STRATEGY=round_robin
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
EOF

./jetbrains-ai-proxy
```

### 场景4: Docker容器部署

```bash
# 1. 创建Dockerfile（如果不存在）
cat > Dockerfile << 'EOF'
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o jetbrains-ai-proxy

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/jetbrains-ai-proxy .
EXPOSE 8080
CMD ["./jetbrains-ai-proxy"]
EOF

# 2. 构建镜像
docker build -t jetbrains-ai-proxy .

# 3. 运行容器
docker run -d \
  --name jetbrains-ai-proxy \
  -p 8080:8080 \
  -e JWT_TOKENS="jwt1,jwt2,jwt3" \
  -e BEARER_TOKEN="your_bearer_token" \
  -e LOAD_BALANCE_STRATEGY="random" \
  jetbrains-ai-proxy

# 4. 或者使用配置文件挂载
docker run -d \
  --name jetbrains-ai-proxy \
  -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  jetbrains-ai-proxy
```

## 🔧 管理和监控

### 健康检查

```bash
# 检查服务状态
curl http://localhost:8080/health

# 响应示例
{
  "status": "ok",
  "healthy_tokens": 3,
  "total_tokens": 3,
  "strategy": "round_robin",
  "server_info": {
    "host": "0.0.0.0",
    "port": 8080
  }
}
```

### 查看配置信息

```bash
# 查看当前配置
curl http://localhost:8080/config

# 响应示例
{
  "jwt_tokens_count": 3,
  "jwt_tokens": [
    {
      "name": "Primary_JWT",
      "description": "Primary JWT token",
      "priority": 1,
      "token_preview": "eyJ0eXAiOiJKV1QiLCJhbGc..."
    }
  ],
  "bearer_token_set": true,
  "load_balance_strategy": "round_robin",
  "health_check_interval": "30s",
  "server_host": "0.0.0.0",
  "server_port": 8080
}
```

### 查看统计信息

```bash
# 查看负载均衡统计
curl http://localhost:8080/stats

# 响应示例
{
  "balancer": {
    "healthy_tokens": 3,
    "total_tokens": 3,
    "strategy": "round_robin"
  },
  "config": {
    "health_check_interval": "30s",
    "server_host": "0.0.0.0",
    "server_port": 8080
  }
}
```

### 重载配置

```bash
# 重新加载配置（无需重启服务）
curl -X POST http://localhost:8080/reload

# 响应示例
{
  "message": "Configuration reloaded successfully"
}
```

## 🧪 测试API

### 发送聊天请求

```bash
# 发送非流式请求
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your_bearer_token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ],
    "stream": false
  }'

# 发送流式请求
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your_bearer_token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "user", "content": "Tell me a story"}
    ],
    "stream": true
  }'
```

### 获取支持的模型

```bash
# 获取模型列表
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer your_bearer_token"
```

## 🚨 故障排除

### 常见问题

1. **所有JWT tokens都不健康**
   ```bash
   # 检查token有效性
   curl http://localhost:8080/health
   
   # 查看日志
   tail -f /var/log/jetbrains-ai-proxy.log
   ```

2. **配置文件未找到**
   ```bash
   # 生成示例配置
   ./jetbrains-ai-proxy --generate-config
   
   # 检查配置文件路径
   ./jetbrains-ai-proxy --print-config
   ```

3. **端口被占用**
   ```bash
   # 检查端口使用情况
   lsof -i :8080
   
   # 使用不同端口启动
   ./jetbrains-ai-proxy -p 9090
   ```

### 日志分析

```bash
# 查看实时日志
tail -f jetbrains-ai-proxy.log

# 过滤健康检查日志
grep "health check" jetbrains-ai-proxy.log

# 过滤错误日志
grep -i error jetbrains-ai-proxy.log
```

## 📈 性能优化建议

1. **JWT Token数量**: 建议配置3-5个JWT tokens以获得最佳负载均衡效果
2. **健康检查间隔**: 生产环境建议设置为30-60秒
3. **负载均衡策略**: 
   - 使用`round_robin`获得均匀分布
   - 使用`random`避免可预测的请求模式
4. **监控**: 定期检查`/health`和`/stats`端点
5. **日志**: 配置适当的日志级别和轮转策略
