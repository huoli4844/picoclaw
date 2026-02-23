#!/bin/bash

# PicoClaw Web 测试脚本

echo "🦞 PicoClaw Web 测试脚本"
echo "========================="

# 1. 检查前端构建
if [ -d "web/dist" ]; then
    echo "✅ 前端已构建"
else
    echo "❌ 前端未构建，正在构建..."
    cd web && pnpm run build && cd ..
fi

# 2. 检查后端编译
if [ -f "picoclaw-web-server" ]; then
    echo "✅ 后端已编译"
else
    echo "❌ 后端未编译，正在编译..."
    go build -o picoclaw-web-server cmd/web-server/main.go
fi

# 3. 启动服务
echo ""
echo "🚀 启动服务..."
echo "后端将在端口 8082 启动"
echo "请访问: http://localhost:8082"
echo ""
echo "按 Ctrl+C 停止服务"
echo ""

# 启动后端服务器
PORT=8082 ./picoclaw-web-server