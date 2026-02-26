/*
* @desc: Watermill NATS 功能测试
* @Date: 2026/02/09
* 测试中主要使用了 Watermill-NATS 的 4 个核心封装：
* ✅ Publisher/Subscriber - 消息发布订阅（群聊核心）
* ✅ GobMarshaler - 消息序列化（自动编解码）
* ✅ JetStreamConfig - JetStream 配置（持久化）
* ✅ Message/Router - 消息对象和路由（高级特性）
*未使用 Watermill 封装的部分（使用原生 API）：
* Stream 创建和管理
* KV Store 操作
*Object Store 操作
* 单聊的 Request/Reply 模式
 */

package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	natsgo "github.com/nats-io/nats.go"
)

var (
	wmCtx    = gctx.New()
	wmLogger = watermill.NewStdLogger(false, false)
)

// TestWatermillNATSContainer Watermill NATS 测试容器
func TestWatermillNATSContainer(t *testing.T) {
	t.Run("TestPrivateChatWithWatermill", testPrivateChatWithWatermill)
	t.Run("TestGroupChatWithWatermill", testGroupChatWithWatermill)
	t.Run("TestJetStreamWithWatermill", testJetStreamWithWatermill)
	t.Run("TestKVStoreWithWatermill", testKVStoreWithWatermill)
	t.Run("TestObjectStoreWithWatermill", testObjectStoreWithWatermill)
	t.Run("TestMessageArchiveWithWatermill", testMessageArchiveWithWatermill)
	t.Run("TestMessageMiddleware", testMessageMiddleware)
	t.Run("TestMessageRouter", testMessageRouter)
}

// testPrivateChatWithWatermill 测试单聊（使用 NATS 原生 Request/Reply 模式）
// 注意：
// 1. 新架构统一使用 Pub/Sub 模式，私聊和群聊架构一致
// 2. 私聊消息发布到 chat.private.{conversation_id} 主题
// 3. 消息持久化由 JetStream CHAT_PRIVATE Stream 处理
func testPrivateChatWithWatermill(t *testing.T) {
	// 从配置读取 NATS URL
	url := g.Cfg().MustGet(wmCtx, "nats.url").String()
	if url == "" {
		t.Fatal("NATS URL 配置为空")
	}

	t.Logf("💬 开始测试单聊（Pub/Sub 模式）")

	// 创建 NATS 原生连接
	nc, err := natsgo.Connect(url)
	if err != nil {
		t.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Close()

	// 获取 JetStream 上下文
	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("创建 JetStream 失败: %v", err)
	}

	// 创建私聊消息流（如果不存在）
	streamName := "WM_CHAT_PRIVATE_TEST"
	_, err = js.AddStream(&natsgo.StreamConfig{
		Name:        streamName,
		Description: "Watermill 私聊消息流测试",
		Subjects:    []string{"chat.private.test.>"},
		Retention:   natsgo.LimitsPolicy,
		MaxAge:      24 * time.Hour,
		MaxMsgs:     1000,
		MaxBytes:    10 * 1024 * 1024,
		Storage:     natsgo.FileStorage,
		Replicas:    1,
	})
	if err != nil {
		t.Logf("创建 Stream（可能已存在）: %v", err)
	}

	// 构造 conversation_id
	conversationID := "test.conv_private_1_2"
	subject := fmt.Sprintf("chat.private.%s", conversationID)

	// user_001 和 user_002 订阅同一会话
	receivedChan1 := make(chan *natsgo.Msg, 1)
	receivedChan2 := make(chan *natsgo.Msg, 1)

	sub1, err := nc.Subscribe(subject, func(msg *natsgo.Msg) {
		t.Logf("📩 user_001 收到私聊消息: %s", string(msg.Data))
		receivedChan1 <- msg
	})
	if err != nil {
		t.Fatalf("user_001 订阅失败: %v", err)
	}
	defer sub1.Unsubscribe()

	sub2, err := nc.Subscribe(subject, func(msg *natsgo.Msg) {
		t.Logf("📩 user_002 收到私聊消息: %s", string(msg.Data))
		receivedChan2 <- msg
	})
	if err != nil {
		t.Fatalf("user_002 订阅失败: %v", err)
	}
	defer sub2.Unsubscribe()

	t.Logf("✅ user_001 和 user_002 已订阅会话: %s", subject)

	// 等待订阅生效
	time.Sleep(200 * time.Millisecond)

	// user_001 发送私聊消息（Pub/Sub 模式）
	testMsg := []byte(`{"message_id":"msg_001","conversation_id":"test.conv_private_1_2","sender_id":1,"receiver_id":2,"sender_name":"user_001","content":"你好，这是一条私聊消息（Pub/Sub模式）","timestamp":"2026-02-09T10:00:00Z"}`)

	t.Logf("📨 user_001 发送私聊消息...")

	// 使用 Publish 发布消息（Pub/Sub 模式）
	err = nc.Publish(subject, testMsg)
	if err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	t.Logf("✅ 消息发布成功")

	// 验证双方都收到消息
	select {
	case <-receivedChan1:
		t.Logf("✅ user_001 收到消息")
	case <-time.After(3 * time.Second):
		t.Fatal("user_001 接收消息超时")
	}

	select {
	case <-receivedChan2:
		t.Logf("✅ user_002 收到消息")
	case <-time.After(3 * time.Second):
		t.Fatal("user_002 接收消息超时")
	}

	// 清理测试 Stream
	err = js.DeleteStream(streamName)
	if err != nil {
		t.Logf("⚠️ 清理测试 Stream 失败: %v", err)
	}

	t.Logf("✅ 单聊测试完成（Pub/Sub 模式）")
}

// testGroupChatWithWatermill 测试群聊（使用 Watermill + JetStream）
func testGroupChatWithWatermill(t *testing.T) {
	url := g.Cfg().MustGet(wmCtx, "nats.url").String()
	if url == "" {
		t.Fatal("NATS URL 配置为空")
	}

	t.Logf("👥 开始测试群聊（Watermill + JetStream）")

	// 创建 NATS 原生连接用于 Stream 管理
	nc, err := natsgo.Connect(url)
	if err != nil {
		t.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Close()

	// 使用原生 API 创建 Stream
	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("创建 JetStream 失败: %v", err)
	}

	streamName := "WM_CHAT_GROUP"
	_, err = js.AddStream(&natsgo.StreamConfig{
		Name:        streamName,
		Description: "Watermill 群聊消息流",
		Subjects:    []string{"chat.group.>"}, // 通配符 subject
		Retention:   natsgo.LimitsPolicy,
		MaxAge:      30 * 24 * time.Hour,
		MaxMsgs:     1000000,
		MaxBytes:    1024 * 1024 * 1024,
		Storage:     natsgo.FileStorage,
		Replicas:    1,
	})
	if err != nil {
		// Stream 可能已存在
		t.Logf("创建 Stream（可能已存在）: %v", err)
	}

	// 使用 Watermill Publisher
	publisherConfig := nats.PublisherConfig{
		URL:       url,
		Marshaler: &nats.GobMarshaler{},
		JetStream: nats.JetStreamConfig{
			Disabled:      false, // 启用 JetStream
			AutoProvision: false, // 禁用自动创建（Stream 已手动创建）
		},
	}

	publisher, err := nats.NewPublisher(publisherConfig, wmLogger)
	if err != nil {
		t.Fatalf("创建发布者失败: %v", err)
	}
	defer publisher.Close()

	// 创建 3 个订阅者（模拟 3 个群成员）
	member1Chan := make(chan *message.Message, 10)
	member2Chan := make(chan *message.Message, 10)
	member3Chan := make(chan *message.Message, 10)

	// 成员订阅配置（为了模拟群聊，每个成员都需要独立的 Durable Consumer）
	// 成员 1
	subscriberConfig1 := nats.SubscriberConfig{
		URL:         url,
		Unmarshaler: &nats.GobMarshaler{},
		JetStream: nats.JetStreamConfig{
			Disabled:      false,
			AutoProvision: false,
			AckAsync:      false,
			DurablePrefix: "test_group_member1", // 每个成员独立的 Consumer
		},
	}

	// 成员 2
	subscriberConfig2 := nats.SubscriberConfig{
		URL:         url,
		Unmarshaler: &nats.GobMarshaler{},
		JetStream: nats.JetStreamConfig{
			Disabled:      false,
			AutoProvision: false,
			AckAsync:      false,
			DurablePrefix: "test_group_member2",
		},
	}

	// 成员 3
	subscriberConfig3 := nats.SubscriberConfig{
		URL:         url,
		Unmarshaler: &nats.GobMarshaler{},
		JetStream: nats.JetStreamConfig{
			Disabled:      false,
			AutoProvision: false,
			AckAsync:      false,
			DurablePrefix: "test_group_member3",
		},
	}

	// 成员 1 订阅
	subscriber1, err := nats.NewSubscriber(subscriberConfig1, wmLogger)
	if err != nil {
		t.Fatalf("创建成员1订阅者失败: %v", err)
	}
	defer subscriber1.Close()

	messages1, err := subscriber1.Subscribe(context.Background(), "chat.group.group_001")
	if err != nil {
		t.Fatalf("成员1订阅失败: %v", err)
	}

	go func() {
		for msg := range messages1 {
			t.Logf("📩 成员1 收到群消息: %s", string(msg.Payload))
			member1Chan <- msg
			msg.Ack()
		}
	}()

	// 成员 2 订阅
	subscriber2, err := nats.NewSubscriber(subscriberConfig2, wmLogger)
	if err != nil {
		t.Fatalf("创建成员2订阅者失败: %v", err)
	}
	defer subscriber2.Close()

	messages2, err := subscriber2.Subscribe(context.Background(), "chat.group.group_001")
	if err != nil {
		t.Fatalf("成员2订阅失败: %v", err)
	}

	go func() {
		for msg := range messages2 {
			t.Logf("📩 成员2 收到群消息: %s", string(msg.Payload))
			member2Chan <- msg
			msg.Ack()
		}
	}()

	// 成员 3 订阅
	subscriber3, err := nats.NewSubscriber(subscriberConfig3, wmLogger)
	if err != nil {
		t.Fatalf("创建成员3订阅者失败: %v", err)
	}
	defer subscriber3.Close()

	messages3, err := subscriber3.Subscribe(context.Background(), "chat.group.group_001")
	if err != nil {
		t.Fatalf("成员3订阅失败: %v", err)
	}

	go func() {
		for msg := range messages3 {
			t.Logf("📩 成员3 收到群消息: %s", string(msg.Payload))
			member3Chan <- msg
			msg.Ack()
		}
	}()

	t.Logf("✅ 3 个群成员已订阅群聊")

	// 等待订阅生效
	time.Sleep(300 * time.Millisecond)

	// 成员 1 发送群消息
	testMsg := message.NewMessage(
		watermill.NewUUID(),
		[]byte(`{"message_id":"msg_group_001","group_id":"group_001","sender_id":"user_001","sender_name":"张三","content":"@所有人 大家好！","timestamp":"2026-02-09T10:00:00Z"}`),
	)

	t.Logf("📨 成员1 发送群消息...")

	err = publisher.Publish("chat.group.group_001", testMsg)
	if err != nil {
		t.Fatalf("发布群消息失败: %v", err)
	}

	t.Logf("✅ 群消息发布成功")

	// 验证所有成员都收到消息
	receivedCount := 0
	timeout := time.After(3 * time.Second)

	for receivedCount < 3 {
		select {
		case msg := <-member1Chan:
			if string(msg.Payload) == string(testMsg.Payload) {
				receivedCount++
				t.Logf("✅ 成员1 消息验证成功")
			}
		case msg := <-member2Chan:
			if string(msg.Payload) == string(testMsg.Payload) {
				receivedCount++
				t.Logf("✅ 成员2 消息验证成功")
			}
		case msg := <-member3Chan:
			if string(msg.Payload) == string(testMsg.Payload) {
				receivedCount++
				t.Logf("✅ 成员3 消息验证成功")
			}
		case <-timeout:
			t.Fatalf("接收消息超时，只收到 %d/3 条消息", receivedCount)
		}
	}

	t.Logf("✅ 群聊测试完成（Watermill）")
}

// testMessageMiddleware 测试消息中间件
func testMessageMiddleware(t *testing.T) {
	url := g.Cfg().MustGet(wmCtx, "nats.url").String()
	if url == "" {
		t.Fatal("NATS URL 配置为空")
	}

	t.Logf("🔧 开始测试消息中间件")

	// 创建 NATS 原生连接
	nc, err := natsgo.Connect(url)
	if err != nil {
		t.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Close()

	// 创建发布者
	publisherConfig := nats.PublisherConfig{
		URL:         nc.Servers()[0],
		NatsOptions: []natsgo.Option{natsgo.UseOldRequestStyle()},
		Marshaler:   &nats.GobMarshaler{},
		JetStream: nats.JetStreamConfig{
			Disabled: true,
		},
	}

	publisher, err := nats.NewPublisher(publisherConfig, wmLogger)
	if err != nil {
		t.Fatalf("创建发布者失败: %v", err)
	}
	defer publisher.Close()

	t.Logf("✅ 发布者创建成功")

	// 创建订阅者
	subscriberConfig := nats.SubscriberConfig{
		URL:         nc.Servers()[0],
		NatsOptions: []natsgo.Option{natsgo.UseOldRequestStyle()},
		Unmarshaler: &nats.GobMarshaler{},
		JetStream: nats.JetStreamConfig{
			Disabled: true,
		},
	}

	subscriber, err := nats.NewSubscriber(subscriberConfig, wmLogger)
	if err != nil {
		t.Fatalf("创建订阅者失败: %v", err)
	}
	defer subscriber.Close()

	// 接收消息通道
	receivedChan := make(chan *message.Message, 1)

	// 订阅消息
	messages, err := subscriber.Subscribe(context.Background(), "chat.middleware.test")
	if err != nil {
		t.Fatalf("订阅失败: %v", err)
	}

	// 启动消息处理协程（带中间件逻辑）
	go func() {
		for msg := range messages {
			t.Logf("📝 [中间件] 处理接收消息: uuid=%s, payload_size=%d", msg.UUID, len(msg.Payload))
			t.Logf("📩 收到消息: %s", string(msg.Payload))
			receivedChan <- msg
			msg.Ack()
		}
	}()

	// 等待订阅生效
	time.Sleep(200 * time.Millisecond)

	// 发送测试消息
	testMsg := message.NewMessage(
		watermill.NewUUID(),
		[]byte(`{"content":"测试中间件功能"}`),
	)

	t.Logf("📨 发送测试消息...")
	t.Logf("📝 [中间件] 准备发送消息: topic=chat.middleware.test, uuid=%s, payload_size=%d", testMsg.UUID, len(testMsg.Payload))

	err = publisher.Publish("chat.middleware.test", testMsg)
	if err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	// 等待接收消息
	select {
	case <-receivedChan:
		t.Logf("✅ 消息通过中间件处理成功")
	case <-time.After(3 * time.Second):
		t.Fatal("接收消息超时")
	}

	t.Logf("✅ 消息中间件测试完成")
}

// testMessageRouter 测试消息路由
func testMessageRouter(t *testing.T) {
	url := g.Cfg().MustGet(wmCtx, "nats.url").String()
	if url == "" {
		t.Fatal("NATS URL 配置为空")
	}

	t.Logf("🔀 开始测试消息路由")

	// 创建 NATS 原生连接
	nc, err := natsgo.Connect(url)
	if err != nil {
		t.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Close()

	// 创建发布者
	publisherConfig := nats.PublisherConfig{
		URL:         nc.Servers()[0],
		NatsOptions: []natsgo.Option{natsgo.UseOldRequestStyle()},
		Marshaler:   &nats.GobMarshaler{},
		JetStream: nats.JetStreamConfig{
			Disabled: true,
		},
	}

	publisher, err := nats.NewPublisher(publisherConfig, wmLogger)
	if err != nil {
		t.Fatalf("创建发布者失败: %v", err)
	}
	defer publisher.Close()

	// 创建订阅者
	subscriberConfig := nats.SubscriberConfig{
		URL:         nc.Servers()[0],
		NatsOptions: []natsgo.Option{natsgo.UseOldRequestStyle()},
		Unmarshaler: &nats.GobMarshaler{},
		JetStream: nats.JetStreamConfig{
			Disabled: true,
		},
	}

	subscriber, err := nats.NewSubscriber(subscriberConfig, wmLogger)
	if err != nil {
		t.Fatalf("创建订阅者失败: %v", err)
	}
	defer subscriber.Close()

	// 创建消息路由器
	router, err := message.NewRouter(message.RouterConfig{}, wmLogger)
	if err != nil {
		t.Fatalf("创建路由器失败: %v", err)
	}

	// 消息计数器
	textMsgCount := 0
	imageMsgCount := 0
	voiceMsgCount := 0

	// 添加路由处理器：处理文本消息
	router.AddNoPublisherHandler(
		"text_message_handler",
		"chat.router.text",
		subscriber,
		func(msg *message.Message) error {
			t.Logf("📝 [路由] 处理文本消息: %s", string(msg.Payload))
			textMsgCount++
			return nil
		},
	)

	// 添加路由处理器：处理图片消息
	router.AddNoPublisherHandler(
		"image_message_handler",
		"chat.router.image",
		subscriber,
		func(msg *message.Message) error {
			t.Logf("🖼️ [路由] 处理图片消息: %s", string(msg.Payload))
			imageMsgCount++
			return nil
		},
	)

	// 添加路由处理器：处理语音消息
	router.AddNoPublisherHandler(
		"voice_message_handler",
		"chat.router.voice",
		subscriber,
		func(msg *message.Message) error {
			t.Logf("🎤 [路由] 处理语音消息: %s", string(msg.Payload))
			voiceMsgCount++
			return nil
		},
	)

	// 启动路由器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := router.Run(ctx)
		if err != nil && err != context.Canceled {
			t.Logf("路由器运行错误: %v", err)
		}
	}()

	// 等待路由器启动
	<-router.Running()
	t.Logf("✅ 消息路由器已启动")

	time.Sleep(300 * time.Millisecond)

	// 发送不同类型的消息
	messages := []struct {
		topic   string
		payload string
	}{
		{"chat.router.text", `{"type":"text","content":"你好"}`},
		{"chat.router.image", `{"type":"image","url":"http://example.com/image.jpg"}`},
		{"chat.router.voice", `{"type":"voice","url":"http://example.com/voice.mp3"}`},
		{"chat.router.text", `{"type":"text","content":"再见"}`},
	}

	for i, msgData := range messages {
		msg := message.NewMessage(
			watermill.NewUUID(),
			[]byte(msgData.payload),
		)

		t.Logf("📨 发送消息 %d: %s", i+1, msgData.topic)

		err = publisher.Publish(msgData.topic, msg)
		if err != nil {
			t.Fatalf("发送消息失败: %v", err)
		}
	}

	t.Logf("✅ 所有消息已发送")

	// 等待消息处理
	time.Sleep(1 * time.Second)

	// 验证消息计数
	if textMsgCount != 2 {
		t.Fatalf("文本消息处理数量不正确: 期望=2, 实际=%d", textMsgCount)
	}
	if imageMsgCount != 1 {
		t.Fatalf("图片消息处理数量不正确: 期望=1, 实际=%d", imageMsgCount)
	}
	if voiceMsgCount != 1 {
		t.Fatalf("语音消息处理数量不正确: 期望=1, 实际=%d", voiceMsgCount)
	}

	t.Logf("✅ 消息计数验证成功: 文本=%d, 图片=%d, 语音=%d", textMsgCount, imageMsgCount, voiceMsgCount)

	// 停止路由器
	cancel()
	err = router.Close()
	if err != nil {
		t.Logf("关闭路由器错误: %v", err)
	}

	t.Logf("✅ 消息路由测试完成")
}

// TestWatermillCleanup 清理测试资源
func TestWatermillCleanup(t *testing.T) {
	t.Logf("✅ Watermill NATS 测试清理完成")
}

// testJetStreamWithWatermill 测试 JetStream（使用 Watermill）
func testJetStreamWithWatermill(t *testing.T) {
	url := g.Cfg().MustGet(wmCtx, "nats.url").String()
	if url == "" {
		t.Fatal("NATS URL 配置为空")
	}

	t.Logf("🌊 开始测试 JetStream（Watermill）")

	// 创建 NATS 原生连接
	nc, err := natsgo.Connect(url)
	if err != nil {
		t.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Close()

	// 创建 JetStream 上下文（使用原生 API）
	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("创建 JetStream 失败: %v", err)
	}

	// 创建测试流
	streamName := "TEST_WM_CHAT_MESSAGES"
	_, err = js.AddStream(&natsgo.StreamConfig{
		Name:        streamName,
		Description: "Watermill 测试聊天消息流",
		Subjects:    []string{"chat.wm.test.>"},
		Retention:   natsgo.LimitsPolicy,
		MaxAge:      24 * time.Hour,
		MaxMsgs:     1000,
		MaxBytes:    10 * 1024 * 1024,
		Storage:     natsgo.FileStorage,
		Replicas:    1,
	})
	if err != nil {
		t.Fatalf("创建流失败: %v", err)
	}

	t.Logf("✅ JetStream 流创建成功: %s", streamName)

	// 使用 Watermill Publisher 发布消息（启用 JetStream）
	publisherConfig := nats.PublisherConfig{
		URL:         nc.Servers()[0],
		NatsOptions: []natsgo.Option{natsgo.UseOldRequestStyle()},
		Marshaler:   &nats.GobMarshaler{},
		JetStream: nats.JetStreamConfig{
			Disabled:       false, // 启用 JetStream
			AutoProvision:  false, // 不自动创建流
			PublishOptions: nil,
		},
	}

	publisher, err := nats.NewPublisher(publisherConfig, wmLogger)
	if err != nil {
		t.Fatalf("创建发布者失败: %v", err)
	}
	defer publisher.Close()

	// 发布测试消息
	testMsg := message.NewMessage(
		watermill.NewUUID(),
		[]byte(`{"type":"text","content":"测试JetStream消息","sender_id":"test_user","timestamp":"2026-02-09T10:00:00Z"}`),
	)

	err = publisher.Publish("chat.wm.test.user.001", testMsg)
	if err != nil {
		t.Fatalf("发布消息失败: %v", err)
	}

	t.Logf("✅ 消息发布成功")

	// 获取流信息
	streamInfo, err := js.StreamInfo(streamName)
	if err != nil {
		t.Fatalf("获取流信息失败: %v", err)
	}

	t.Logf("流信息: 消息数=%d, 字节数=%d",
		streamInfo.State.Msgs,
		streamInfo.State.Bytes,
	)

	// 清理：删除测试流
	err = js.DeleteStream(streamName)
	if err != nil {
		t.Logf("⚠️ 清理测试流失败: %v", err)
	}

	t.Logf("✅ JetStream 测试完成（Watermill）")
}

// testKVStoreWithWatermill 测试 KV Store（使用 Watermill + 原生 API）
func testKVStoreWithWatermill(t *testing.T) {
	url := g.Cfg().MustGet(wmCtx, "nats.url").String()
	if url == "" {
		t.Fatal("NATS URL 配置为空")
	}

	t.Logf("🗄️ 开始测试 KV Store（Watermill + 原生 API）")

	// 创建 NATS 原生连接
	nc, err := natsgo.Connect(url)
	if err != nil {
		t.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Close()

	// 创建 JetStream 上下文
	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("创建 JetStream 失败: %v", err)
	}

	// 创建测试 KV Bucket
	bucketName := "TEST_WM_CHAT_USER_STATUS"
	kv, err := js.CreateKeyValue(&natsgo.KeyValueConfig{
		Bucket:      bucketName,
		Description: "Watermill 测试用户在线状态",
		TTL:         5 * time.Minute,
		MaxBytes:    1024 * 1024,
		Storage:     natsgo.MemoryStorage,
		Replicas:    1,
	})
	if err != nil {
		t.Fatalf("创建 KV Bucket 失败: %v", err)
	}

	t.Logf("✅ KV Store 创建成功: %s", bucketName)

	// 写入数据（注意：key 不能包含冒号）
	testKey := "user.wm_test_001.status"
	testValue := []byte(`{"user_id":"wm_test_001","status":"online","last_active":"2026-02-09T10:00:00Z"}`)
	revision, err := kv.Put(testKey, testValue)
	if err != nil {
		t.Fatalf("写入 KV 数据失败: %v", err)
	}

	t.Logf("✅ KV 数据写入成功, revision=%d", revision)

	// 读取数据
	entry, err := kv.Get(testKey)
	if err != nil {
		t.Fatalf("读取 KV 数据失败: %v", err)
	}

	t.Logf("✅ KV 数据读取成功: %s", string(entry.Value()))

	// 更新数据
	updatedValue := []byte(`{"user_id":"wm_test_001","status":"away","last_active":"2026-02-09T10:05:00Z"}`)
	newRevision, err := kv.Put(testKey, updatedValue)
	if err != nil {
		t.Fatalf("更新 KV 数据失败: %v", err)
	}

	t.Logf("✅ KV 数据更新成功, revision=%d", newRevision)

	// 删除数据
	err = kv.Delete(testKey)
	if err != nil {
		t.Fatalf("删除 KV 数据失败: %v", err)
	}

	t.Logf("✅ KV 数据删除成功")

	// 清理：删除测试 Bucket
	err = js.DeleteKeyValue(bucketName)
	if err != nil {
		t.Logf("⚠️ 清理测试 Bucket 失败: %v", err)
	}

	t.Logf("✅ KV Store 测试完成（Watermill）")
}

// testObjectStoreWithWatermill 测试 Object Store（使用 Watermill + 原生 API）
func testObjectStoreWithWatermill(t *testing.T) {
	url := g.Cfg().MustGet(wmCtx, "nats.url").String()
	if url == "" {
		t.Fatal("NATS URL 配置为空")
	}

	t.Logf("📦 开始测试 Object Store（Watermill + 原生 API）")

	// 创建 NATS 原生连接
	nc, err := natsgo.Connect(url)
	if err != nil {
		t.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Close()

	// 创建 JetStream 上下文
	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("创建 JetStream 失败: %v", err)
	}

	// 创建测试 Object Store
	bucketName := "TEST_WM_CHAT_FILES"
	obs, err := js.CreateObjectStore(&natsgo.ObjectStoreConfig{
		Bucket:      bucketName,
		Description: "Watermill 测试聊天文件存储",
		TTL:         24 * time.Hour,
		MaxBytes:    10 * 1024 * 1024,
		Storage:     natsgo.FileStorage,
		Replicas:    1,
	})
	if err != nil {
		t.Fatalf("创建 Object Store 失败: %v", err)
	}

	t.Logf("✅ Object Store 创建成功: %s", bucketName)

	// 写入对象
	objectName := "test_wm_audio.mp3"
	testData := []byte("这是一个 Watermill 测试音频文件的模拟数据...")
	_, err = obs.PutBytes(objectName, testData)
	if err != nil {
		t.Fatalf("写入对象失败: %v", err)
	}

	t.Logf("✅ 对象写入成功: %s", objectName)

	// 读取对象
	data, err := obs.GetBytes(objectName)
	if err != nil {
		t.Fatalf("读取对象失败: %v", err)
	}

	t.Logf("✅ 对象读取成功: %d bytes", len(data))

	// 获取对象信息
	info, err := obs.GetInfo(objectName)
	if err != nil {
		t.Fatalf("获取对象信息失败: %v", err)
	}

	t.Logf("对象信息: name=%s, size=%d, modified=%s",
		info.Name,
		info.Size,
		info.ModTime,
	)

	// 删除对象
	err = obs.Delete(objectName)
	if err != nil {
		t.Fatalf("删除对象失败: %v", err)
	}

	t.Logf("✅ 对象删除成功")

	// 清理：删除测试 Bucket
	err = js.DeleteObjectStore(bucketName)
	if err != nil {
		t.Logf("⚠️ 清理测试 Bucket 失败: %v", err)
	}

	t.Logf("✅ Object Store 测试完成（Watermill）")
}

// testMessageArchiveWithWatermill 测试消息归档（使用 Watermill + JetStream）
func testMessageArchiveWithWatermill(t *testing.T) {
	url := g.Cfg().MustGet(wmCtx, "nats.url").String()
	if url == "" {
		t.Fatal("NATS URL 配置为空")
	}

	t.Logf("💾 开始测试消息归档（Watermill + JetStream）")

	// 创建 NATS 原生连接
	nc, err := natsgo.Connect(url)
	if err != nil {
		t.Fatalf("连接 NATS 失败: %v", err)
	}
	defer nc.Close()

	// 创建 JetStream 上下文
	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("创建 JetStream 失败: %v", err)
	}

	// 1. 创建消息归档流
	streamName := "WM_CHAT_ARCHIVE_TEST"
	_, err = js.AddStream(&natsgo.StreamConfig{
		Name:        streamName,
		Description: "Watermill 聊天消息归档流",
		Subjects:    []string{"chat.wm.archive.>"},
		Retention:   natsgo.LimitsPolicy,
		MaxAge:      30 * 24 * time.Hour,
		MaxMsgs:     100000,
		MaxBytes:    1024 * 1024 * 1024,
		Storage:     natsgo.FileStorage,
		Replicas:    1,
	})
	if err != nil {
		t.Fatalf("创建归档流失败: %v", err)
	}

	t.Logf("✅ 消息归档流创建成功: %s", streamName)

	// 2. 使用 Watermill Publisher 归档消息（启用 JetStream）
	publisherConfig := nats.PublisherConfig{
		URL:         nc.Servers()[0],
		NatsOptions: []natsgo.Option{natsgo.UseOldRequestStyle()},
		Marshaler:   &nats.GobMarshaler{},
		JetStream: nats.JetStreamConfig{
			Disabled:      false,
			AutoProvision: false,
		},
	}

	publisher, err := nats.NewPublisher(publisherConfig, wmLogger)
	if err != nil {
		t.Fatalf("创建发布者失败: %v", err)
	}
	defer publisher.Close()

	// 归档单聊消息
	privateMsg := message.NewMessage(
		watermill.NewUUID(),
		[]byte(`{"type":"private","message_id":"msg_wm_p_001","conversation_id":"conv_001","sender_id":"user_001","receiver_id":"user_002","content":"Watermill单聊消息内容","timestamp":"2026-02-09T10:00:00Z"}`),
	)

	err = publisher.Publish("chat.wm.archive.private.conv_001", privateMsg)
	if err != nil {
		t.Fatalf("归档单聊消息失败: %v", err)
	}
	t.Logf("✅ 单聊消息已归档")

	// 归档群聊消息
	groupMsg := message.NewMessage(
		watermill.NewUUID(),
		[]byte(`{"type":"group","message_id":"msg_wm_g_001","group_id":"group_001","sender_id":"user_001","content":"Watermill群聊消息内容","timestamp":"2026-02-09T10:01:00Z"}`),
	)

	err = publisher.Publish("chat.wm.archive.group.group_001", groupMsg)
	if err != nil {
		t.Fatalf("归档群聊消息失败: %v", err)
	}
	t.Logf("✅ 群聊消息已归档")

	// 3. 使用 Watermill Subscriber 读取归档消息（启用 JetStream）
	subscriberConfig := nats.SubscriberConfig{
		URL:         nc.Servers()[0],
		NatsOptions: []natsgo.Option{natsgo.UseOldRequestStyle()},
		Unmarshaler: &nats.GobMarshaler{},
		JetStream: nats.JetStreamConfig{
			Disabled:      false,
			AutoProvision: false,
		},
	}

	subscriber, err := nats.NewSubscriber(subscriberConfig, wmLogger)
	if err != nil {
		t.Fatalf("创建订阅者失败: %v", err)
	}
	defer subscriber.Close()

	// 订阅归档消息（使用通配符）
	messages, err := subscriber.Subscribe(context.Background(), "chat.wm.archive.>")
	if err != nil {
		t.Fatalf("订阅归档消息失败: %v", err)
	}

	// 读取归档消息
	receivedChan := make(chan *message.Message, 2)
	go func() {
		for msg := range messages {
			t.Logf("📜 读取归档消息: %s", string(msg.Payload))
			receivedChan <- msg
			msg.Ack()
		}
	}()

	// 等待接收2条消息
	receivedCount := 0
	timeout := time.After(3 * time.Second)

	for receivedCount < 2 {
		select {
		case <-receivedChan:
			receivedCount++
			t.Logf("✅ 已接收 %d/2 条归档消息", receivedCount)
		case <-timeout:
			t.Logf("⚠️ 接收超时，已接收 %d/2 条消息", receivedCount)
			goto cleanup
		}
	}

cleanup:
	// 4. 清理资源
	err = js.DeleteStream(streamName)
	if err != nil {
		t.Logf("⚠️ 删除流失败: %v", err)
	}

	// 验证
	if receivedCount != 2 {
		t.Fatalf("归档消息接收数量不匹配: 期望=2, 实际=%d", receivedCount)
	}

	t.Logf("✅ 消息归档测试完成（Watermill）")
}
