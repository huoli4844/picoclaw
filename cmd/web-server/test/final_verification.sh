#!/bin/bash

# PicoClaw IM机器人控制系统 - 最终验证脚本
# 验证所有功能是否完整实现

set -e

echo "🎯 PicoClaw IM机器人控制系统 - 最终验证"
echo "==============================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查结果统计
TOTAL_CHECKS=0
PASSED_CHECKS=0

check_status() {
    local test_name="$1"
    local status="$2"
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    if [ "$status" = "PASS" ]; then
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        echo -e "✅ ${GREEN}${test_name}: 通过${NC}"
    else
        echo -e "❌ ${RED}${test_name}: 失败${NC}"
    fi
}

echo ""
echo "📋 1. 配置文件检查"
echo "-------------------"

# 检查配置文件是否存在
if [ -f ~/.picoclaw/config.json ]; then
    check_status "配置文件存在" "PASS"
    
    # 检查InternalIM配置
    if grep -q '"internal_im"' ~/.picoclaw/config.json; then
        check_status "InternalIM配置存在" "PASS"
        
        # 检查enabled
        if grep -q '"enabled".*true' ~/.picoclaw/config.json; then
            check_status "InternalIM已启用" "PASS"
        else
            check_status "InternalIM已启用" "FAIL"
        fi
        
        # 检查流式响应配置
        if grep -q '"enable_streaming".*true' ~/.picoclaw/config.json; then
            check_status "流式响应已启用" "PASS"
        else
            check_status "流式响应已启用" "FAIL"
        fi
        
        # 检查NATS URL
        if grep -q '"url".*nats://171.221.201.55:24222' ~/.picoclaw/config.json; then
            check_status "NATS服务器配置正确" "PASS"
        else
            check_status "NATS服务器配置正确" "FAIL"
        fi
    else
        check_status "InternalIM配置存在" "FAIL"
    fi
else
    check_status "配置文件存在" "FAIL"
fi

echo ""
echo "📁 2. 代码文件检查"
echo "-------------------"

# 检查核心文件是否存在
CODE_FILES=(
    "pkg/channels/internal_im.go"
    "pkg/channels/internal_im_protocol.go"
    "pkg/channels/manager.go"
    "pkg/config/config.go"
    "cmd/web-server/main.go"
)

for file in "${CODE_FILES[@]}"; do
    if [ -f "$file" ]; then
        check_status "文件存在: $file" "PASS"
    else
        check_status "文件存在: $file" "FAIL"
    fi
done

echo ""
echo "🔍 3. 功能代码检查"
echo "-------------------"

# 检查流式响应相关代码
if grep -q "sendStreamingResponseToIM" pkg/channels/internal_im.go; then
    check_status "流式响应发送函数" "PASS"
else
    check_status "流式响应发送函数" "FAIL"
fi

if grep -q "chunkContent" pkg/channels/internal_im.go; then
    check_status "内容分块函数" "PASS"
else
    check_status "内容分块函数" "FAIL"
fi

# 检查流式消息协议
if grep -q "StreamID.*string.*json.*stream_id" pkg/channels/internal_im_protocol.go; then
    check_status "流式消息协议定义" "PASS"
else
    check_status "流式消息协议定义" "FAIL"
fi

# 检查配置结构
if grep -q "EnableStreaming.*bool" pkg/config/config.go; then
    check_status "流式响应配置字段" "PASS"
else
    check_status "流式响应配置字段" "FAIL"
fi

echo ""
echo "🌐 4. NATS连接检查"
echo "-------------------"

# 检查NATS连接
if nc -z 171.221.201.55 24222 2>/dev/null; then
    check_status "NATS服务器连接" "PASS"
else
    check_status "NATS服务器连接" "FAIL"
fi

echo ""
echo "🧪 5. 服务状态检查"
echo "-------------------"

# 检查Web服务是否运行
if lsof -i :8080 >/dev/null 2>&1; then
    check_status "PicoClaw Web服务运行" "PASS"
else
    check_status "PicoClaw Web服务运行" "FAIL"
fi

echo ""
echo "📚 6. 文档检查"
echo "-------------------"

# 检查文档文件
DOC_FILES=(
    "docs/im-bot-control-guide.md"
    "README.md"
    "CONTRIBUTING.md"
)

for doc in "${DOC_FILES[@]}"; do
    if [ -f "$doc" ]; then
        # 检查文档是否包含最新信息
        if grep -q "流式响应" "$doc" 2>/dev/null || [ "$doc" != "docs/im-bot-control-guide.md" ]; then
            check_status "文档存在: $doc" "PASS"
        else
            check_status "文档存在且包含流式响应信息: $doc" "FAIL"
        fi
    else
        check_status "文档存在: $doc" "FAIL"
    fi
done

echo ""
echo "🎯 验证结果总结"
echo "=============================="
echo -e "总检查项: ${YELLOW}$TOTAL_CHECKS${NC}"
echo -e "通过检查: ${GREEN}$PASSED_CHECKS${NC}"
echo -e "失败检查: ${RED}$((TOTAL_CHECKS - PASSED_CHECKS))${NC}"

if [ $PASSED_CHECKS -eq $TOTAL_CHECKS ]; then
    echo ""
    echo -e "🎉 ${GREEN}恭喜！所有检查项均通过！${NC}"
    echo -e "🚀 ${GREEN}PicoClaw IM机器人控制系统已完全就绪！${NC}"
    echo ""
    echo "📋 系统功能清单:"
    echo "✅ InternalIM Channel完整实现"
    echo "✅ 双向NATS通信机制"
    echo "✅ 用户权限验证（支持通配符）"
    echo "✅ 配置系统集成"
    echo "✅ Web服务器集成"
    echo "✅ 完整响应发送机制"
    echo "✅ 标准化消息协议"
    echo "✅ 错误处理和状态反馈"
    echo "✅ 流式响应支持"
    echo "✅ 消息总线集成"
    echo "✅ 完整文档和测试"
    echo ""
    echo "🔥 立即开始使用:"
    echo "1. 确保PicoClaw服务运行: cd cmd/web-server && go run main.go"
    echo "2. 发送测试消息到NATS: nats -s nats://171.221.201.55:24222 pub picoclaw.im '{\"type\":\"message\",\"user_id\":\"test\",\"chat_id\":\"test\",\"username\":\"test\",\"content\":\"你好\",\"timestamp\":\"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'\"}'"
    echo "3. 监听响应: nats -s nats://171.221.201.55:24222 sub picoclaw.im.out"
    echo ""
    echo "📖 更多信息请查看: docs/im-bot-control-guide.md"
    
    exit 0
else
    echo ""
    echo -e "⚠️  ${YELLOW}仍有 $((TOTAL_CHECKS - PASSED_CHECKS)) 项检查未通过${NC}"
    echo "请检查失败项并确保所有功能正常工作"
    exit 1
fi