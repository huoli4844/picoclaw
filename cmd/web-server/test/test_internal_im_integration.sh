#!/bin/bash

# InternalIM集成测试脚本
# 这个脚本测试完整的IM机器人控制流程

set -e

echo "🧪 开始InternalIM集成测试..."

# 检查NATS是否运行（使用远程服务器）
NATS_URL="nats://171.221.201.55:24222"
echo "📡 检查NATS连接 ($NATS_URL)..."
if ! nc -z 171.221.201.55 24222 2>/dev/null; then
    echo "❌ NATS服务器未运行，请检查远程服务器：$NATS_URL"
    exit 1
fi
echo "✅ NATS连接正常"

# 构建测试消息
TEST_USER_ID="test_user_$(date +%s)"
TEST_CHAT_ID="test_chat_$(date +%s)"
TEST_MESSAGE="你好，请帮我分析一下当前系统的状态"

# 创建测试消息文件
TEST_MSG_FILE="/tmp/picoclaw_test_msg.json"
cat > "$TEST_MSG_FILE" << EOF
{
  "type": "message",
  "user_id": "$TEST_USER_ID",
  "chat_id": "$TEST_CHAT_ID",
  "username": "测试用户",
  "content": "$TEST_MESSAGE",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

echo "📝 测试消息已生成："
cat "$TEST_MSG_FILE"
echo ""

# 检查是否有nats-cli工具
if command -v nats &> /dev/null; then
    echo "📤 通过nats-cli发送测试消息..."
    nats -s "$NATS_URL" pub picoclaw.im "$(cat "$TEST_MSG_FILE")"
    echo "✅ 消息已发送到picoclaw.im主题"
else
    echo "⚠️  未找到nats-cli工具，跳过消息发送测试"
fi

# 等待处理
echo "⏳ 等待消息处理..."
sleep 3

# 创建配置检查脚本
CONFIG_CHECK_SCRIPT="/tmp/check_picoclaw_config.go"
cat > "$CONFIG_CHECK_SCRIPT" << 'EOF'
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ChannelsConfig struct {
	InternalIM struct {
		Enabled bool `json:"enabled"`
		URL     string `json:"url"`
		Topic   string `json:"topic"`
	} `json:"internal_im"`
}

type Config struct {
	Channels ChannelsConfig `json:"channels"`
}

func main() {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".picoclaw", "config.json")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("❌ 无法读取配置文件: %v\n", err)
		return
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("❌ 无法解析配置文件: %v\n", err)
		return
	}
	
	if !config.Channels.InternalIM.Enabled {
		fmt.Println("⚠️  InternalIM通道未启用")
		return
	}
	
	expectedURL := "nats://171.221.201.55:24222"
	if config.Channels.InternalIM.URL != expectedURL {
		fmt.Printf("⚠️  InternalIM URL配置不匹配\n")
		fmt.Printf("   当前: %s\n", config.Channels.InternalIM.URL)
		fmt.Printf("   期望: %s\n", expectedURL)
		return
	}
	
	fmt.Printf("✅ InternalIM配置正常\n")
	fmt.Printf("   URL: %s\n", config.Channels.InternalIM.URL)
	fmt.Printf("   Topic: %s\n", config.Channels.InternalIM.Topic)
}
EOF

# 检查配置
echo "🔍 检查PicoClaw配置..."
if command -v go &> /dev/null; then
    go run "$CONFIG_CHECK_SCRIPT"
else
    echo "⚠️  未找到go工具，跳过配置检查"
fi

# 清理临时文件
rm -f "$TEST_MSG_FILE" "$CONFIG_CHECK_SCRIPT"

echo ""
echo "🎯 测试完成！"
echo ""
echo "📋 下一步操作："
echo "1. 确保~/.picoclaw/config.json中启用了internal_im通道"
echo "2. 启动PicoClaw Web服务器: cd cmd/web-server && go run main.go"
echo "3. 使用IM机器人发送消息测试完整流程"
echo ""
echo "💡 提示：查看PicoClaw日志了解消息处理状态"
echo "   tail -f ~/.picoclaw/logs/picoclaw.log"