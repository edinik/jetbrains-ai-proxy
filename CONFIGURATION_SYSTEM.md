# 统一配置系统说明

JetBrains AI Proxy现在采用了全新的统一配置系统，实现了自动配置发现、多种配置方式支持和配置热重载等功能。

## 🏗️ 系统架构

### 核心组件

1. **ConfigManager** (`internal/config/config.go`)
   - 统一的配置管理器
   - 支持多种配置源的合并
   - 线程安全的配置访问
   - 配置验证和默认值处理

2. **ConfigDiscovery** (`internal/config/discovery.go`)
   - 自动配置文件发现
   - 配置文件格式验证
   - 配置文件监控和热重载
   - 示例配置生成

3. **JWTBalancer** (`internal/balancer/jwt_balancer.go`)
   - JWT负载均衡器
   - 支持轮询和随机策略
   - 并发安全的token管理
   - 动态token更新

4. **HealthChecker** (`internal/balancer/health_checker.go`)
   - JWT健康检查器
   - 自动故障检测和恢复
   - 可配置的检查间隔
   - 并发健康检查

## 📁 文件结构

```
jetbrains-ai-proxy/
├── main.go                          # 主程序，支持新配置系统
├── start.sh                         # 智能启动脚本
├── CONFIGURATION_SYSTEM.md          # 配置系统说明（本文件）
├── MULTI_JWT_README.md              # 多JWT功能说明
├── internal/
│   ├── config/
│   │   ├── config.go                # 配置管理器
│   │   └── discovery.go             # 配置发现器
│   ├── balancer/
│   │   ├── jwt_balancer.go          # JWT负载均衡器
│   │   ├── health_checker.go        # 健康检查器
│   │   └── jwt_balancer_test.go     # 测试用例
│   └── jetbrains/
│       └── client.go                # 集成配置系统的客户端
└── examples/
    ├── complete_example.md          # 完整使用示例
    ├── start_with_multiple_jwt.sh   # 多JWT启动脚本
    └── .env.example                 # 环境变量示例
```

## ⚙️ 配置优先级

系统按以下优先级加载和合并配置：

1. **命令行参数** (最高优先级)
2. **环境变量**
3. **配置文件**
4. **默认值** (最低优先级)

## 🔍 配置发现机制

系统会按以下顺序搜索配置文件：

1. `CONFIG_FILE` 环境变量指定的路径
2. 当前目录：`config.json`, `jetbrains-ai-proxy.json`
3. config目录：`config/config.json`, `configs/config.json`
4. 隐藏目录：`.config/config.json`
5. 用户主目录：`$HOME/.config/jetbrains-ai-proxy/config.json`
6. 系统目录：`/etc/jetbrains-ai-proxy/config.json`

如果没有找到配置文件，系统会自动生成示例配置。

## 🚀 使用方式

### 1. 自动配置（推荐）

```bash
# 生成示例配置
./jetbrains-ai-proxy --generate-config

# 编辑配置文件
vim config/config.json

# 启动服务（自动发现配置）
./jetbrains-ai-proxy
```

### 2. 使用启动脚本

```bash
# 智能启动（自动检查配置）
./start.sh

# 生成配置
./start.sh --generate

# 查看配置
./start.sh --config
```

### 3. 指定配置文件

```bash
# 使用特定配置文件
./jetbrains-ai-proxy --config /path/to/config.json

# 或设置环境变量
export CONFIG_FILE=/path/to/config.json
./jetbrains-ai-proxy
```

## 📋 配置文件格式

### JSON配置文件示例

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
    }
  ],
  "bearer_token": "your_bearer_token_here",
  "load_balance_strategy": "round_robin",
  "health_check_interval": "30s",
  "server_port": 8080,
  "server_host": "0.0.0.0"
}
```

### 环境变量配置

```bash
# JWT Tokens（逗号分隔）
JWT_TOKENS=token1,token2,token3

# Bearer Token
BEARER_TOKEN=your_bearer_token

# 负载均衡策略
LOAD_BALANCE_STRATEGY=round_robin

# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# 配置文件路径（可选）
CONFIG_FILE=config/config.json
```

## 🔄 配置热重载

系统支持运行时配置重载，无需重启服务：

### 自动重载

配置文件监控器会自动检测配置文件变化并重新加载。

### 手动重载

```bash
# 通过API端点重载
curl -X POST http://localhost:8080/reload

# 响应
{
  "message": "Configuration reloaded successfully"
}
```

## 🛠️ 管理端点

系统提供了丰富的管理端点：

| 端点 | 方法 | 描述 |
|------|------|------|
| `/health` | GET | 健康检查和负载均衡状态 |
| `/config` | GET | 当前配置信息（隐藏敏感数据） |
| `/stats` | GET | 详细统计信息 |
| `/reload` | POST | 重新加载配置 |

## 🔧 高级功能

### 1. JWT Token元数据

支持为每个JWT token配置元数据：

```json
{
  "token": "jwt_token_here",
  "name": "Production_Primary",
  "description": "Primary production JWT token",
  "priority": 1,
  "metadata": {
    "environment": "production",
    "region": "us-east-1",
    "tier": "primary",
    "max_requests_per_minute": "1000"
  }
}
```

### 2. 配置验证

系统会自动验证配置的有效性：
- JWT token格式检查
- 必需字段验证
- 数值范围检查
- 策略有效性验证

### 3. 配置合并策略

多个配置源的合并规则：
- 数组类型：高优先级完全覆盖低优先级
- 对象类型：递归合并，高优先级字段覆盖低优先级
- 基本类型：高优先级直接覆盖低优先级

## 🚨 故障排除

### 配置问题诊断

```bash
# 查看当前配置
./jetbrains-ai-proxy --print-config

# 验证配置文件
./jetbrains-ai-proxy --config config.json --print-config

# 生成新的示例配置
./jetbrains-ai-proxy --generate-config
```

### 常见问题

1. **配置文件未找到**
   - 检查文件路径和权限
   - 使用 `--generate-config` 生成示例配置

2. **JWT tokens无效**
   - 检查token格式和有效性
   - 查看健康检查日志

3. **配置合并问题**
   - 使用 `--print-config` 查看最终配置
   - 检查配置优先级

## 📈 性能考虑

1. **配置缓存**: 配置在内存中缓存，避免重复读取
2. **并发安全**: 使用读写锁保护配置访问
3. **懒加载**: 配置发现器按需加载配置文件
4. **监控优化**: 配置文件监控使用高效的文件系统事件

## 🔮 未来扩展

系统设计支持以下扩展：
- 远程配置中心集成（如Consul、etcd）
- 配置加密和安全存储
- 配置版本管理和回滚
- 更多负载均衡策略
- 动态配置更新API
