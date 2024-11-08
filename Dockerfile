# 使用多阶段构建
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache make git

# 复制源代码
COPY . .

# 构建应用
RUN make build

# 第二阶段：运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/build/proxypool /app/proxypool

# 创建必要的目录
RUN mkdir -p /app/data

# 设置时区
ENV TZ=Asia/Shanghai

# 声明数据卷
VOLUME ["/app/data"]

# 暴露端口
EXPOSE 8080

# 设置入口命令
ENTRYPOINT ["/app/proxypool"] 