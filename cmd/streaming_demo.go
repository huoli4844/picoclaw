package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	natsgo "github.com/nats-io/nats.go"
)

// IMMessage 匹配PicoClaw的消息格式
type IMMessage struct {
	Type         string `json:"type"`
	UserID       string `json:"user_id"`
	ChatID       string `json:"chat_id"`
	Username     string `json:"username,omitempty"`
	Content      string `json:"content,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	StatusCode   string `json:"status_code,omitempty"`
	Timestamp    string `json:"timestamp"`
	// 流式响应字段
	StreamID    string `json:"stream_id,omitempty"`
	IsStreamEnd bool   `json:"is_stream_end,omitempty"`
	ChunkIndex  int    `json:"chunk_index,omitempty"`
}

func main() {
	// 连接NATS
	nc, err := natsgo.Connect("nats://171.221.201.55:24222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	fmt.Println("🧪 开始流式响应测试...")
	fmt.Println("📡 已连接到NATS服务器")

	// 订阅响应主题
	sub, err := nc.Subscribe("picoclaw.im.out", func(msg *natsgo.Msg) {
		var imMsg IMMessage
		if err := json.Unmarshal(msg.Data, &imMsg); err != nil {
			fmt.Printf("❌ 解析响应消息失败: %v\n", err)
			return
		}

		switch imMsg.Type {
		case "status":
			fmt.Printf("📊 状态更新: %s - %s\n", imMsg.StatusCode, imMsg.Content)
		case "response":
			fmt.Printf("💬 普通响应: %s\n", imMsg.Content)
		case "stream":
			fmt.Printf("🌊 流式块[%d]: %s%s\n", imMsg.ChunkIndex, imMsg.Content, func() string {
				if imMsg.IsStreamEnd {
					return " [结束]"
				}
				return ""
			}())
		case "error":
			fmt.Printf("❌ 错误: %s - %s\n", imMsg.StatusCode, imMsg.ErrorMessage)
		}
	})
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	// 等待订阅建立
	time.Sleep(1 * time.Second)

	// 发送测试消息
	testUserID := "stream_test_user"
	testChatID := "stream_test_chat"
	testMessage := "请生成一段较长的代码示例，包含详细的注释和说明，用于测试流式响应功能"

	imMsg := IMMessage{
		Type:      "message",
		UserID:    testUserID,
		ChatID:    testChatID,
		Username:  "流式测试用户",
		Content:   testMessage,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	msgData, err := json.Marshal(imMsg)
	if err != nil {
		log.Fatalf("Failed to marshal test message: %v", err)
	}

	fmt.Printf("📤 发送测试消息: %s\n", testMessage)
	if err := nc.Publish("picoclaw.im", msgData); err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	fmt.Println("⏳ 等待响应...")
	fmt.Println("💡 提示: 观察是否有流式响应块出现")

	// 等待信号或超时
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 30秒超时
	timeout := time.After(30 * time.Second)

	select {
	case <-sigChan:
		fmt.Println("\n🛑 收到停止信号")
	case <-timeout:
		fmt.Println("\n⏰ 测试超时")
	case <-time.After(15 * time.Second):
		fmt.Println("\n✅ 测试完成")
	}

	fmt.Println("📊 测试总结:")
	fmt.Println("- 如果看到流式块消息，说明流式响应功能正常")
	fmt.Println("- 如果只看到普通响应，可能是配置问题")
	fmt.Println("- 如果没有响应，请检查PicoClaw服务状态")
}
