#!/bin/bash

# 创建必要的目录
mkdir -p data

# 启动服务
docker-compose up -d

# 显示服务状态
docker-compose ps

# 等待服务启动
echo "等待服务启动..."
sleep 5

# 检查服务健康状态
curl http://localhost:8080/health