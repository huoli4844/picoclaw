package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/channels"
	"github.com/sipeed/picoclaw/pkg/config"
)

// TestPicoClawNATSChannel 测试PicoClaw的NATS Channel功能
func TestPicoClawNATSChannel(t *testing.T) {
	// 检查环境变量
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://171.221.201.55:24222"
	}

	t.Run("ChannelLifecycle", func(t *testing.T) {
		t.Log("🔄 测试NATS Channel生命周期")

		// 创建NATS配置
		cfg := config.NATSConfig{
			Enabled:         true,
			URL:             natsURL,
			WebSocket:       "wss://171.221.201.55:28444",
			Topic:           "picoclaw_test_topic",
			EnableJetStream: true,
			AllowFrom:       config.FlexibleStringSlice{"user_888"},
		}

		// 创建消息总线
		messageBus := bus.NewMessageBus()

		// 创建NATS Channel
		channel, err := channels.NewSimpleNATSChannel(cfg, messageBus)
		if err != nil {
			t.Fatalf("❌ 创建NATS Channel失败: %v", err)
		}

		// 测试启动
		ctx := context.Background()
		err = channel.Start(ctx)
		if err != nil {
			t.Fatalf("❌ 启动NATS Channel失败: %v", err)
		}

		t.Log("✅ NATS Channel 启动成功")

		// 等待连接建立
		time.Sleep(100 * time.Millisecond)

		// 测试发送消息
		outboundMsg := bus.OutboundMessage{
			ChatID:  "user_888",
			Content: "你好，这是来自PicoClaw的测试消息",
		}

		err = channel.Send(ctx, outboundMsg)
		if err != nil {
			t.Fatalf("❌ 发送消息失败: %v", err)
		}

		t.Log("✅ 消息发送成功")

		// 测试停止
		err = channel.Stop(ctx)
		if err != nil {
			t.Fatalf("❌ 停止NATS Channel失败: %v", err)
		}

		t.Log("✅ NATS Channel 停止成功")
	})

	t.Run("MessageProcessing", func(t *testing.T) {
		t.Log("📨 测试消息处理")

		// 创建NATS配置
		cfg := config.NATSConfig{
			Enabled:         true,
			URL:             natsURL,
			Topic:           "picoclaw_test_messages",
			EnableJetStream: true,
			AllowFrom:       config.FlexibleStringSlice{"test123"},
		}

		// 创建消息总线
		messageBus := bus.NewMessageBus()

		// 监听消息总线消息
		receivedMessages := make(chan bus.InboundMessage, 10)

		// 启动goroutine来监听入站消息
		go func() {
			defer close(receivedMessages)
			timeout := time.After(3 * time.Second) // 增加超时保护
			for {
				select {
				case <-timeout:
					return
				default:
					// 使用带超时的context避免无限阻塞
					msgCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
					msg, ok := messageBus.ConsumeInbound(msgCtx)
					cancel()

					if !ok {
						time.Sleep(10 * time.Millisecond)
						continue
					}
					receivedMessages <- msg
				}
			}
		}()

		// 创建NATS Channel
		channel, err := channels.NewSimpleNATSChannel(cfg, messageBus)
		if err != nil {
			t.Fatalf("❌ 创建NATS Channel失败: %v", err)
		}

		// 启动Channel
		ctx := context.Background()
		err = channel.Start(ctx)
		if err != nil {
			t.Fatalf("❌ 启动NATS Channel失败: %v", err)
		}

		// 等待订阅建立
		time.Sleep(200 * time.Millisecond)

		// 创建外部NATS连接来发送测试消息
		nc, err := nats.Connect(natsURL)
		if err != nil {
			t.Fatalf("❌ 连接NATS失败: %v", err)
		}
		defer nc.Close()

		// 发送测试消息
		testMsg := map[string]interface{}{
			"type":       "request",
			"message_id": fmt.Sprintf("test_%d", time.Now().UnixNano()),
			"sender_id":  "test123",
			"content":    "这是一条测试消息",
			"model":      "deepseek",
			"stream":     false,
			"session_id": "session_test_123",
			"metadata":   map[string]string{"source": "test"},
			"timestamp":  time.Now().UnixMilli(),
		}

		msgBytes, err := json.Marshal(testMsg)
		if err != nil {
			t.Fatalf("❌ 序列化消息失败: %v", err)
		}

		err = nc.Publish(cfg.Topic, msgBytes)
		if err != nil {
			t.Fatalf("❌ 发布消息失败: %v", err)
		}

		t.Log("📤 测试消息已发送")

		// 等待消息处理
		select {
		case receivedMsg := <-receivedMessages:
			t.Logf("✅ 收到处理后的消息: ChatID=%s, Content=%s", receivedMsg.ChatID, receivedMsg.Content)

			// 验证消息内容
			if receivedMsg.ChatID != "test123" {
				t.Errorf("❌ ChatID不匹配: 期望=test123, 实际=%s", receivedMsg.ChatID)
			}

			if receivedMsg.Content != "这是一条测试消息" {
				t.Errorf("❌ Content不匹配: 期望='这是一条测试消息', 实际=%s", receivedMsg.Content)
			}

			// 验证元数据
			if receivedMsg.Metadata["channel"] != "nats" {
				t.Errorf("❌ Channel元数据不匹配: 期望=nats, 实际=%s", receivedMsg.Metadata["channel"])
			}

			if receivedMsg.Metadata["user_id"] != "test123" {
				t.Errorf("❌ UserID元数据不匹配: 期望=test123, 实际=%s", receivedMsg.Metadata["user_id"])
			}

		case <-time.After(5 * time.Second):
			t.Fatal("❌ 等待消息处理超时")
		}

		// 停止Channel
		err = channel.Stop(ctx)
		if err != nil {
			t.Fatalf("❌ 停止NATS Channel失败: %v", err)
		}

		t.Log("✅ 消息处理测试完成")
	})

	t.Run("WatermillMiddleware", func(t *testing.T) {
		t.Log("⚙️ 测试Watermill中间件功能")

		// 创建NATS配置
		cfg := config.NATSConfig{
			Enabled:         true,
			URL:             natsURL,
			Topic:           "picoclaw_test_middleware_simple",
			EnableJetStream: false, // 禁用JetStream避免重复投递问题
			AllowFrom:       config.FlexibleStringSlice{"middleware_user"},
		}

		// 创建消息总线
		messageBus := bus.NewMessageBus()

		// 监听消息总线消息
		receivedCount := 0
		done := make(chan bool)

		// 启动goroutine来监听入站消息
		go func() {
			defer close(done)
			timeout := time.After(2 * time.Second)
			for {
				select {
				case <-timeout:
					return
				default:
					// 使用带超时的context避免阻塞
					msgCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
					msg, ok := messageBus.ConsumeInbound(msgCtx)
					cancel()

					if !ok {
						time.Sleep(10 * time.Millisecond) // 短暂休眠避免CPU占用过高
						continue
					}
					// 只计数有效消息
					if msg.ChatID != "" && msg.Content != "" {
						receivedCount++
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()

		// 创建NATS Channel
		channel, err := channels.NewSimpleNATSChannel(cfg, messageBus)
		if err != nil {
			t.Fatalf("❌ 创建NATS Channel失败: %v", err)
		}

		// 启动Channel
		ctx := context.Background()
		err = channel.Start(ctx)
		if err != nil {
			t.Fatalf("❌ 启动NATS Channel失败: %v", err)
		}

		// 等待订阅建立
		time.Sleep(200 * time.Millisecond)

		// 创建外部NATS连接来发送多条测试消息
		nc, err := nats.Connect(natsURL)
		if err != nil {
			t.Fatalf("❌ 连接NATS失败: %v", err)
		}
		defer nc.Close()

		// 发送多条测试消息
		for i := 0; i < 5; i++ {
			testMsg := map[string]interface{}{
				"type":       "request",
				"message_id": fmt.Sprintf("middleware_test_%d_%d", i, time.Now().UnixNano()),
				"sender_id":  "middleware_user",
				"content":    fmt.Sprintf("中间件测试消息 %d", i+1),
				"timestamp":  time.Now().UnixMilli(),
			}

			msgBytes, _ := json.Marshal(testMsg)
			nc.Publish(cfg.Topic, msgBytes)
			time.Sleep(50 * time.Millisecond) // 短暂间隔
		}

		t.Log("📤 多条测试消息已发送")

		// 等待所有消息处理完成
		time.Sleep(1 * time.Second)

		// 验证消息处理结果
		if receivedCount != 5 {
			t.Errorf("❌ 消息处理数量不匹配: 期望=5, 实际=%d", receivedCount)
		} else {
			t.Logf("✅ 中间件成功处理了 %d 条消息", receivedCount)
		}

		// 停止Channel
		err = channel.Stop(ctx)
		if err != nil {
			t.Fatalf("❌ 停止NATS Channel失败: %v", err)
		}

		// 等待消息接收goroutine结束
		<-done

		t.Log("✅ Watermill中间件测试完成")
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		t.Log("🚨 测试错误处理")

		// 创建NATS配置
		cfg := config.NATSConfig{
			Enabled:         true,
			URL:             natsURL,
			Topic:           "picoclaw_test_error",
			EnableJetStream: true,
			AllowFrom:       config.FlexibleStringSlice{"error_user"}, // 只允许error_user
		}

		// 创建消息总线
		messageBus := bus.NewMessageBus()

		// 监听消息总线消息
		unauthorizedCount := 0
		done := make(chan bool)

		// 启动goroutine来监听入站消息
		go func() {
			defer close(done)
			timeout := time.After(2 * time.Second)
			noMessageCount := 0

			for {
				select {
				case <-timeout:
					// 超时，退出goroutine
					return
				default:
					// 尝试消费消息，设置更短的超时避免阻塞
					msgCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
					_, ok := messageBus.ConsumeInbound(msgCtx)
					cancel()

					if !ok {
						// 没有消息，增加计数
						noMessageCount++
						if noMessageCount > 10 { // 连续10次没有消息就稍微休息一下
							time.Sleep(100 * time.Millisecond)
							noMessageCount = 0
						}
						continue
					}
					unauthorizedCount++
					noMessageCount = 0 // 重置计数
				}
			}
		}()

		// 创建NATS Channel
		channel, err := channels.NewSimpleNATSChannel(cfg, messageBus)
		if err != nil {
			t.Fatalf("❌ 创建NATS Channel失败: %v", err)
		}

		// 启动Channel
		ctx := context.Background()
		err = channel.Start(ctx)
		if err != nil {
			t.Fatalf("❌ 启动NATS Channel失败: %v", err)
		}

		// 等待订阅建立
		time.Sleep(200 * time.Millisecond)

		// 创建外部NATS连接来发送未授权用户的消息
		nc, err := nats.Connect(natsURL)
		if err != nil {
			t.Fatalf("❌ 连接NATS失败: %v", err)
		}
		defer nc.Close()

		// 发送未授权用户的消息
		testMsg := map[string]interface{}{
			"type":       "request",
			"message_id": fmt.Sprintf("unauthorized_%d", time.Now().UnixNano()),
			"sender_id":  "unauthorized_user", // 不在允许列表中
			"content":    "这是未授权用户的消息",
			"timestamp":  time.Now().UnixMilli(),
		}

		msgBytes, _ := json.Marshal(testMsg)
		err = nc.Publish(cfg.Topic, msgBytes)
		if err != nil {
			t.Fatalf("❌ 发布未授权消息失败: %v", err)
		}

		t.Log("📤 未授权用户消息已发送")

		// 等待一段时间确保消息被处理（或拒绝）
		time.Sleep(1 * time.Second)

		// 验证未授权消息被拒绝
		if unauthorizedCount > 0 {
			t.Errorf("❌ 未授权消息被错误处理: 期望=0, 实际=%d", unauthorizedCount)
		} else {
			t.Log("✅ 未授权消息被正确拒绝")
		}

		// 停止Channel
		err = channel.Stop(ctx)
		if err != nil {
			t.Fatalf("❌ 停止NATS Channel失败: %v", err)
		}

		// 等待消息接收goroutine结束，减少超时时间
		select {
		case <-done:
			t.Log("✅ 监听goroutine正常结束")
		case <-time.After(1 * time.Second):
			t.Log("⚠️ 监听goroutine超时，但继续执行")
		}

		t.Log("✅ 错误处理测试完成")
	})
}
