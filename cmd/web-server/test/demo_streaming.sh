#!/bin/bash

# PicoClaw 流式响应演示脚本
# 展示完整的流式响应功能

set -e

echo "🌊 PicoClaw 流式响应功能演示"
echo "=================================="

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查基础环境
echo -e "${BLUE}📋 1. 环境检查${NC}"

# 检查nats-cli
if ! command -v nats &> /dev/null; then
    echo -e "${YELLOW}⚠️  未找到nats-cli，正在安装...${NC}"
    echo "go install github.com/nats-io/natscli/nats@latest"
    go install github.com/nats-io/natscli/nats@latest
fi

# 检查配置文件
if [ ! -f ~/.picoclaw/config.json ]; then
    echo -e "${YELLOW}❌ 配置文件不存在${NC}"
    exit 1
fi

# 检查流式响应配置
if grep -q '"enable_streaming".*true' ~/.picoclaw/config.json; then
    echo -e "${GREEN}✅ 流式响应已启用${NC}"
else
    echo -e "${YELLOW}⚠️  流式响应未启用，请检查配置${NC}"
    echo "在 ~/.picoclaw/config.json 中设置 \"enable_streaming\": true"
fi

# 检查NATS连接
if nc -z 171.221.201.55 24222 2>/dev/null; then
    echo -e "${GREEN}✅ NATS服务器连接正常${NC}"
else
    echo -e "${YELLOW}❌ NATS服务器连接失败${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}🧪 2. 流式响应演示${NC}"

# 创建测试消息
DEMO_MESSAGES=(
    "请生成一个详细的Python类示例，包含多个方法、错误处理、类型注解和完整的文档字符串，用于测试流式响应功能"
    "写一个Go语言的并发程序，展示goroutine和channel的使用，包含详细的注释和错误处理"
    "创建一个完整的Web API设计文档，包含RESTful接口规范、认证机制、错误码定义和最佳实践"
)

for i in "${!DEMO_MESSAGES[@]}"; do
    msg_num=$((i + 1))
    message="${DEMO_MESSAGES[$i]}"
    
    echo ""
    echo -e "${BLUE}📤 演示 $msg_num: 发送长消息触发流式响应${NC}"
    echo "消息内容: ${message:0:50}..."
    
    # 创建消息JSON
    cat > "/tmp/demo_msg_$msg_num.json" << EOF
{
  "type": "message",
  "user_id": "demo_user_$msg_num",
  "chat_id": "demo_chat_$msg_num",
  "username": "流式演示用户$msg_num",
  "content": "$message",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    
    # 发送消息
    echo "发送消息..."
    nats -s nats://171.221.201.55:24222 pub picoclaw.im "$(cat "/tmp/demo_msg_$msg_num.json")"
    
    # 短暂等待
    sleep 2
done

echo ""
echo -e "${BLUE}📡 3. 监听响应指南${NC}"
echo "请在新的终端窗口中运行以下命令来监听流式响应："
echo ""
echo -e "${GREEN}nats -s nats://171.221.201.55:24222 sub picoclaw.im.out${NC}"
echo ""
echo -e "${YELLOW}💡 你应该看到：${NC}"
echo "1. 状态消息：🔄 开始生成流式响应..."
echo "2. 多个流式块：每块50字符，带chunk_index"
echo "3. 结束标记：最后一块包含 is_stream_end: true"
echo "4. 相同的stream_id：所有块共享同一个会话ID"

echo ""
echo -e "${BLUE}🧪 4. 验证清单${NC}"
echo "请在监听响应时验证以下特性："
echo ""
echo "✅ 状态通知：收到'开始生成流式响应'状态"
echo "✅ 内容分块：长内容被分成多个小块"
echo "✅ 流式标识：每块包含stream_id和chunk_index"
echo "✅ 顺序发送：chunk_index按0,1,2...顺序递增"
echo "✅ 结束标记：最后一块标记is_stream_end: true"
echo "✅ 会话一致：所有块使用相同的stream_id"
echo "✅ 时间间隔：块之间有适当的延迟效果"

echo ""
echo -e "${BLUE}🔧 5. 故障排除${NC}"
echo "如果没有收到流式响应："
echo ""
echo "1. 检查配置："
echo "   grep 'enable_streaming' ~/.picoclaw/config.json"
echo ""
echo "2. 检查服务："
echo "   ps aux | grep 'go run cmd/web-server/main.go'"
echo ""
echo "3. 检查NATS："
echo "   nc -z 171.221.201.55 24222 && echo 'NATS正常' || echo 'NATS异常'"
echo ""
echo "4. 查看日志："
echo "   tail -f ~/.picoclaw/logs/picoclaw.log | grep 'internal-im'"

# 清理临时文件
rm -f /tmp/demo_msg_*.json

echo ""
echo -e "${GREEN}🎉 流式响应演示准备完成！${NC}"
echo "现在请在另一个终端监听响应，然后观察流式效果。"
echo ""
echo -e "${BLUE}📚 更多信息请查看：${NC}"
echo "- 文档: docs/im-bot-control-guide.md"
echo "- 代码: pkg/channels/internal_im.go"
echo "- 配置: ~/.picoclaw/config.json"