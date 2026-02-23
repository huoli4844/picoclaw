#!/bin/bash

# PicoClaw Web 启动脚本

set -e

echo "🦞 PicoClaw Web 启动脚本"
echo "=========================="

# 检查是否已安装 Node.js
if ! command -v node &> /dev/null; then
    echo "❌ 错误: 请先安装 Node.js (推荐 18+ 版本)"
    echo "   访问: https://nodejs.org/"
    exit 1
fi

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 请先安装 Go"
    echo "   访问: https://golang.org/"
    exit 1
fi

echo "📦 安装前端依赖..."
cd web
if [ ! -d "node_modules" ]; then
    npm install
else
    echo "✅ 前端依赖已安装"
fi

echo "🔨 构建前端..."
npm run build

echo "🚀 启动后端服务器..."
cd ..
go run cmd/web-server/main.go &

BACKEND_PID=$!
echo "✅ 后端服务器已启动 (PID: $BACKEND_PID)"

# 等待后端启动
sleep 3

echo "🌐 服务地址:"
echo "   前端界面: http://localhost:8080"
echo "   API 接口:  http://localhost:8080/api"
echo ""
echo "💡 提示:"
echo "   - 请确保已配置 ~/.picoclaw/config.json"
echo "   - 按 Ctrl+C 停止服务器"
echo ""

# 等待用户中断
wait $BACKEND_PID