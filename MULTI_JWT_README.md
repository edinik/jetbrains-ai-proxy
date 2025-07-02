# JetBrains AI Proxy - 多JWT负载均衡系统

本项目提供了一个功能完整的多JWT负载均衡系统，支持自动配置发现、健康检查和故障转移。

## 🎯 核心特性

- ✅ **智能配置管理**: 自动发现和加载配置文件，支持多种配置方式
- ✅ **多JWT支持**: 支持配置多个JWT tokens进行负载均衡
- ✅ **负载均衡策略**: 支持轮询(round_robin)和随机(random)两种策略
- ✅ **健康检查**: 自动检测失效的tokens并从负载均衡池中移除
- ✅ **故障转移**: 当某个token失效时自动切换到其他健康的token
- ✅ **配置热重载**: 支持运行时重新加载配置
- ✅ **并发安全**: 支持高并发环境下的安全使用
- ✅ **管理端点**: 提供健康检查、配置查看、统计信息等管理接口
- ✅ **优雅关闭**: 支持优雅关闭和资源清理

## 🚀 快速开始

### 方式1: 使用启动脚本（推荐）

```bash
# 1. 生成示例配置
./start.sh --generate

# 2. 编辑配置文件
vim config/config.json
# 或编辑环境变量文件
cp .env.example .env && vim .env

# 3. 启动服务
./start.sh
```

### 方式2: 直接使用可执行文件

```bash
# 生成示例配置
./jetbrains-ai-proxy --generate-config

# 查看当前配置
./jetbrains-ai-proxy --print-config

# 启动服务
./jetbrains-ai-proxy
```

## ⚙️ 配置方式

系统支持多种配置方式，优先级从高到低：

1. **命令行参数** (最高优先级)
2. **环境变量**
3. **配置文件**
4. **默认值** (最低优先级)

### 1. 配置文件方式（推荐）

系统会自动搜索以下路径的配置文件：

- `config.json`
- `config/config.json`
- `configs/config.json`
- `.config/jetbrains-ai-proxy.json`
- `$HOME/.config/jetbrains-ai-proxy/config.json`

配置文件示例：

```json
{
  "jetbrains_tokens": [
    {
      "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
      "name": "Primary_JWT",
      "description": "Primary JWT token for JetBrains AI",
      "priority": 1,
      "metadata": {
        "environment": "production",
        "region": "us-east-1"
      }
    },
    {
      "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
      "name": "Secondary_JWT",
      "description": "Secondary JWT token for load balancing",
      "priority": 2,
      "metadata": {
        "environment": "production",
        "region": "us-west-2"
      }
    }
  ],
  "bearer_token": "your_bearer_token_here",
  "load_balance_strategy": "round_robin",
  "health_check_interval": "30s",
  "server_port": 8080,
  "server_host": "0.0.0.0"
}
```

### 2. 环境变量配置

```bash
# 多个JWT tokens，用逗号分隔
export JWT_TOKENS="jwt_token_1,jwt_token_2,jwt_token_3"

# 或者使用旧的单token配置（向后兼容）
export JWT_TOKEN="single_jwt_token"

# Bearer token
export BEARER_TOKEN="your_bearer_token"

# 负载均衡策略（可选，默认为round_robin）
export LOAD_BALANCE_STRATEGY="random"

# 服务器配置
export SERVER_HOST="0.0.0.0"
export SERVER_PORT="8080"

# 指定配置文件路径（可选）
export CONFIG_FILE="path/to/config.json"
```

### 3. 命令行参数配置

```bash
# 查看所有选项
./jetbrains-ai-proxy --help

# 使用命令行参数启动
./jetbrains-ai-proxy \
  -c "jwt1,jwt2,jwt3" \
  -k "bearer_token" \
  -s "round_robin" \
  -p 8080 \
  -h "0.0.0.0"

# 指定配置文件
./jetbrains-ai-proxy --config config/my-config.json
```

## 负载均衡策略

### 轮询策略 (round_robin)

- **默认策略**
- 按顺序依次使用每个健康的JWT token
- 确保负载均匀分布
- 适合大多数场景

### 随机策略 (random)

- 随机选择一个健康的JWT token
- 避免可预测的请求模式
- 适合需要随机分布的场景

## 健康检查机制

系统会自动进行JWT token健康检查：

- **检查间隔**: 每30秒检查一次
- **检查方式**: 发送测试请求到JetBrains AI API
- **故障处理**: 自动标记失效的tokens为不健康状态
- **恢复机制**: 定期重新检查不健康的tokens

## 监控端点

访问 `/health` 端点可以查看负载均衡器状态：

```bash
curl http://localhost:8080/health
```

响应示例：

```json
{
  "status": "ok",
  "healthy_tokens": 2,
  "total_tokens": 3,
  "strategy": "round_robin"
}
```

## 使用示例

### 启动服务

```bash
# 使用3个JWT tokens，轮询策略
./jetbrains-ai-proxy \
  -p 8080 \
  -c "eyJ0eXAiOiJKV1QiLCJhbGc...,eyJ0eXAiOiJKV1QiLCJhbGc...,eyJ0eXAiOiJKV1QiLCJhbGc..." \
  -k "your_bearer_token" \
  -s "round_robin"
```

### 发送请求

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your_bearer_token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "user", "content": "Hello, world!"}
    ],
    "stream": false
  }'
```

## 日志输出

启动时会显示负载均衡器配置信息：

```
2024/01/01 12:00:00 JWT balancer initialized with 3 tokens, strategy: round_robin
2024/01/01 12:00:00 JWT health checker started
2024/01/01 12:00:00 Server starting on 0.0.0.0:8080
2024/01/01 12:00:00 JWT tokens configured: 3
2024/01/01 12:00:00 Load balance strategy: round_robin
```

运行时会显示健康检查和token使用情况：

```
2024/01/01 12:01:00 Performing JWT health check...
2024/01/01 12:01:01 Health check completed: 3/3 tokens healthy
2024/01/01 12:01:30 JWT token marked as unhealthy: eyJ0eXAiOi... (errors: 1)
2024/01/01 12:02:00 JWT token marked as healthy: eyJ0eXAiOi...
```

## 故障排除

### 1. 所有tokens都不健康

如果所有JWT tokens都被标记为不健康，请求会返回错误：

```json
{
  "error": "no available JWT tokens: no healthy JWT tokens available"
}
```

**解决方案**:

- 检查JWT tokens是否有效
- 检查网络连接
- 查看健康检查日志

### 2. 部分tokens不健康

系统会自动使用健康的tokens，但建议：

- 检查不健康tokens的有效性
- 考虑更换失效的tokens
- 监控健康检查日志

### 3. 性能优化

- 根据实际负载调整JWT tokens数量
- 选择合适的负载均衡策略
- 监控 `/health` 端点的响应

## 技术实现

### 核心组件

1. **JWTBalancer**: 负载均衡器接口和实现
2. **HealthChecker**: JWT健康检查器
3. **Config**: 配置管理，支持多JWT配置
4. **Client**: 集成负载均衡器的HTTP客户端

### 并发安全

- 使用读写锁保护共享状态
- 原子操作处理计数器
- 线程安全的随机数生成器

### 错误处理

- 自动重试机制
- 优雅的错误降级
- 详细的错误日志记录
