# Builder Stage - 使用明确的版本并优化缓存
FROM golang:1.24-alpine as builder

# 设置工作目录
WORKDIR /app

# 1. 仅复制依赖描述文件
COPY go.mod go.sum ./

# 2. 下载依赖项。这一步会被缓存，只有在 go.mod/go.sum 变化时才会重新运行
RUN go mod download

# 3. 复制项目源码
COPY . .

# 4. 编译应用。现在此步骤将使用缓存的依赖
RUN go build -o jetbrains-ai-proxy

# Final Stage - 保持不变
FROM alpine
LABEL maintainer="zouyq <zyqcn@live.com>"

COPY --from=builder /app/jetbrains-ai-proxy /usr/local/bin/

ENTRYPOINT ["jetbrains-ai-proxy"]