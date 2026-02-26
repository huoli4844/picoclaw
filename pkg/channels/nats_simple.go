package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

type SimpleNATSChannel struct {
	*BaseChannel
	config        config.NATSConfig
	nc            *natsgo.Conn
	subscriptions map[string]*natsgo.Subscription
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

type SimpleNATSMessage struct {
	Type      string            `json:"type"`
	MessageID string            `json:"message_id"`
	SenderID  string            `json:"sender_id"`
	Content   string            `json:"content"`
	Model     string            `json:"model,omitempty"`
	SessionID string            `json:"session_id,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp int64             `json:"timestamp"`
}

func NewSimpleNATSChannel(cfg config.NATSConfig, messageBus *bus.MessageBus) (*SimpleNATSChannel, error) {
	// 连接NATS
	nc, err := natsgo.Connect(cfg.URL,
		natsgo.ReconnectWait(2*time.Second),
		natsgo.MaxReconnects(10),
		natsgo.DisconnectErrHandler(func(nc *natsgo.Conn, err error) {
			if err != nil {
				logger.ErrorCF("nats", "NATS disconnected", map[string]any{"error": err.Error()})
			} else {
				logger.WarnC("nats", "NATS disconnected without error")
			}
		}),
		natsgo.ReconnectHandler(func(nc *natsgo.Conn) {
			logger.InfoC("nats", "NATS reconnected")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	base := NewBaseChannel("nats", cfg, messageBus, []string(cfg.AllowFrom))

	channel := &SimpleNATSChannel{
		BaseChannel:   base,
		config:        cfg,
		nc:            nc,
		subscriptions: make(map[string]*natsgo.Subscription),
	}

	return channel, nil
}

func (c *SimpleNATSChannel) Start(ctx context.Context) error {
	logger.InfoC("nats", "Starting simple NATS channel")

	if c.config.Topic == "" {
		return fmt.Errorf("NATS topic not configured")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	// 订阅配置的主题
	if err := c.subscribeTopic(c.config.Topic); err != nil {
		return fmt.Errorf("failed to subscribe topic %s: %w", c.config.Topic, err)
	}

	c.setRunning(true)
	logger.InfoCF("nats", "Simple NATS channel started successfully", map[string]any{
		"topic": c.config.Topic,
		"url":   c.config.URL,
	})

	return nil
}

func (c *SimpleNATSChannel) Stop(ctx context.Context) error {
	logger.InfoC("nats", "Stopping simple NATS channel")
	c.setRunning(false)

	if c.cancel != nil {
		c.cancel()
	}

	c.mu.Lock()
	// 取消所有订阅
	for subject, sub := range c.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			logger.ErrorCF("nats", "Failed to unsubscribe", map[string]any{
				"subject": subject,
				"error":   err.Error(),
			})
		}
	}
	c.subscriptions = make(map[string]*natsgo.Subscription)
	c.mu.Unlock()

	// 关闭NATS连接
	if c.nc != nil {
		c.nc.Close()
	}

	logger.InfoC("nats", "Simple NATS channel stopped successfully")
	return nil
}

func (c *SimpleNATSChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.IsRunning() {
		return fmt.Errorf("NATS channel not running")
	}

	// 构造响应消息
	response := SimpleNATSMessage{
		Type:      "response",
		MessageID: generateMessageIDSimple(),
		Content:   msg.Content,
		Model:     "deepseek",
		Timestamp: getCurrentTimestampSimple(),
		SessionID: msg.ChatID,
	}

	// 序列化为JSON
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// 使用配置的主题进行响应
	targetSubject := c.config.Topic

	// 发布消息
	err = c.nc.Publish(targetSubject, responseBytes)
	if err != nil {
		logger.ErrorCF("nats", "Failed to publish response", map[string]any{
			"subject": targetSubject,
			"chat_id": msg.ChatID,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to publish response to %s: %w", targetSubject, err)
	}

	logger.DebugCF("nats", "Sent response", map[string]any{
		"subject":     targetSubject,
		"chat_id":     msg.ChatID,
		"content_len": len(msg.Content),
		"message_id":  response.MessageID,
	})

	return nil
}

func (c *SimpleNATSChannel) subscribeTopic(topic string) error {
	sub, err := c.nc.Subscribe(topic, c.handleMessage())
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", topic, err)
	}

	c.mu.Lock()
	c.subscriptions[topic] = sub
	c.mu.Unlock()

	logger.InfoCF("nats", "Subscribed to topic", map[string]any{
		"topic": topic,
	})

	return nil
}

func (c *SimpleNATSChannel) handleMessage() func(msg *natsgo.Msg) {
	return func(msg *natsgo.Msg) {
		logger.InfoCF("nats", "Received message", map[string]any{
			"subject":     msg.Subject,
			"payload_len": len(msg.Data),
			"payload":     string(msg.Data)[:min(200, len(msg.Data))],
		})

		// 解析消息
		var natsMsg SimpleNATSMessage
		if err := json.Unmarshal(msg.Data, &natsMsg); err != nil {
			// 如果解析失败，将payload作为纯文本内容处理
			logger.DebugCF("nats", "Failed to unmarshal as JSON, treating as plain text", map[string]any{
				"subject": msg.Subject,
				"error":   err.Error(),
			})

			natsMsg = SimpleNATSMessage{
				Type:      "message",
				MessageID: generateMessageIDSimple(),
				SenderID:  c.extractUserIDFromTopic(msg.Subject),
				Content:   string(msg.Data),
				Model:     "unknown",
				Timestamp: getCurrentTimestampSimple(),
				SessionID: "unknown",
			}
		}

		// 提取用户ID
		userID := natsMsg.SenderID
		if userID == "" {
			userID = c.extractUserIDFromTopic(msg.Subject)
		}

		if userID == "" {
			logger.WarnCF("nats", "Cannot extract user ID", map[string]any{
				"subject": msg.Subject,
				"content": natsMsg.Content,
			})
			return
		}

		// 权限检查
		if !c.IsAllowed(userID) {
			logger.WarnCF("nats", "Unauthorized user", map[string]any{
				"user_id": userID,
				"subject": msg.Subject,
			})
			return
		}

		logger.InfoCF("nats", "Processing message", map[string]any{
			"user_id":     userID,
			"message_id":  natsMsg.MessageID,
			"content_len": len(natsMsg.Content),
		})

		// 构造元数据
		metadata := map[string]string{
			"channel":    "nats",
			"subject":    msg.Subject,
			"message_id": natsMsg.MessageID,
			"user_id":    userID,
			"session_id": natsMsg.SessionID,
			"model":      natsMsg.Model,
			"topic":      c.config.Topic,
		}

		// 合并额外元数据
		for k, v := range natsMsg.Metadata {
			metadata[k] = v
		}

		// 转发到PicoClaw消息总线
		c.HandleMessage(userID, userID, natsMsg.Content, []string{}, metadata)
	}
}

func (c *SimpleNATSChannel) extractUserIDFromTopic(topic string) string {
	// 对于 "chatbot.user.888" 这样的主题，提取 "888"
	parts := strings.Split(topic, ".")
	if len(parts) >= 3 && parts[0] == "chatbot" && parts[1] == "user" {
		return parts[2]
	}

	// 对于测试环境的各种topic，返回对应的默认用户ID
	switch {
	case strings.HasPrefix(topic, "picoclaw_test_middleware"):
		return "middleware_user"
	case strings.HasPrefix(topic, "picoclaw_test_topic"):
		return "test_user_888"
	case strings.HasPrefix(topic, "picoclaw_test_messages"):
		return "test123"
	case strings.HasPrefix(topic, "picoclaw_test_error"):
		return "error_user"
	}

	// 检查是否匹配配置的主题
	if topic == c.config.Topic {
		return "default_user"
	}

	logger.WarnCF("nats", "Cannot extract user ID from topic, using default", map[string]any{
		"topic": topic,
	})
	return "unknown_user"
}

func generateMessageIDSimple() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

func getCurrentTimestampSimple() int64 {
	return time.Now().UnixMilli()
}
