#!/bin/bash
set -e

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "Docker 未安装，请先安装 Docker"
    echo "运行: curl -fsSL https://get.docker.com | sh"
    exit 1
fi

# 检查 Docker Compose
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "Docker Compose 未安装"
    exit 1
fi

# 构建并启动
cd /home/prj/rss-reader
echo "构建并启动服务..."
docker-compose up -d --build

echo "服务已启动在 http://localhost:80"
