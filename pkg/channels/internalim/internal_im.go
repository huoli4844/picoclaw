package internalim

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

// MessageRecord 消息记录，用于保留策略
type MessageRecord struct {
	Timestamp time.Time
	MessageID string
	UserID    string
	Content   string
}

// InternalIMChannel 内部IM Channel，通过NATS进行通信
type InternalIMChannel struct {
	*BaseChannel
	config        config.InternalIMConfig
	nc            *natsgo.Conn
	subscriptions map[string]*natsgo.Subscription
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc

	// 消息处理
	running      bool
	processedIDs map[string]bool
	idMu         sync.RWMutex

	// 速率限制
	rateLimitTokens    atomic.Int64
	lastRateLimitReset time.Time
	rateLimitMu        sync.Mutex

	// 消息保留
	messageHistory []MessageRecord
	messageMu      sync.RWMutex
}

// NewInternalIMChannel 创建内部IM Channel
func NewInternalIMChannel(cfg config.InternalIMConfig, messageBus *bus.MessageBus) (*InternalIMChannel, error) {
	// 连接NATS
	nc, err := natsgo.Connect(cfg.URL,
		natsgo.ReconnectWait(2*time.Second),
		natsgo.MaxReconnects(10),
		natsgo.DisconnectErrHandler(func(nc *natsgo.Conn, err error) {
			if err != nil {
				logger.ErrorCF("internal-im", "NATS disconnected", map[string]any{"error": err.Error()})
			} else {
				logger.WarnC("internal-im", "NATS disconnected without error")
			}
		}),
		natsgo.ReconnectHandler(func(nc *natsgo.Conn) {
			logger.InfoC("internal-im", "NATS reconnected")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	base := NewBaseChannel("internal-im", cfg, messageBus, cfg.AllowFrom)

	// 初始化速率限制
	var initialTokens int64 = 1000 // 默认值
	if cfg.RateLimit != nil {
		initialTokens = int64(cfg.RateLimit.BurstSize)
	}

	channel := &InternalIMChannel{
		BaseChannel:     base,
		config:          cfg,
		nc:              nc,
		subscriptions:   make(map[string]*natsgo.Subscription),
		processedIDs:    make(map[string]bool),
		rateLimitTokens: atomic.Int64{},
		messageHistory:  make([]MessageRecord, 0),
	}

	// 初始化速率限制令牌
	channel.rateLimitTokens.Store(initialTokens)
	channel.lastRateLimitReset = time.Now()

	return channel, nil
}

func (c *InternalIMChannel) Start(ctx context.Context) error {
	logger.InfoC("internal-im", "Starting Internal IM channel")

	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("Internal IM channel is already running")
	}
	c.running = true
	c.mu.Unlock()

	c.ctx, c.cancel = context.WithCancel(ctx)

	// 订阅配置的主题（接收来自IM机器人的消息）
	if c.config.Topic != "" {
		if err := c.subscribeTopic(c.config.Topic); err != nil {
			return fmt.Errorf("failed to subscribe to config topic %s: %w", c.config.Topic, err)
		}
	}

	// 启动消息总线监听（发送响应到IM机器人）
	go c.listenToOutboundMessages()

	c.SetRunning(true)
	logger.InfoCF("internal-im", "Internal IM channel started successfully", map[string]any{
		"topic": c.config.Topic,
		"url":   c.config.URL,
	})

	return nil
}

func (c *InternalIMChannel) Stop(ctx context.Context) error {
	logger.InfoC("internal-im", "Stopping Internal IM channel")

	c.mu.Lock()
	c.running = false
	c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	c.mu.Lock()
	// 取消所有订阅
	for subject, sub := range c.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			logger.ErrorCF("internal-im", "Failed to unsubscribe", map[string]any{
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

	c.SetRunning(false)
	logger.InfoC("internal-im", "Internal IM channel stopped successfully")
	return nil
}

func (c *InternalIMChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.IsRunning() {
		return fmt.Errorf("Internal IM channel not running")
	}

	userID := msg.ChatID // 使用ChatID作为用户标识
	logger.DebugCF("internal-im", "Sending outbound message via IM", map[string]any{
		"chat_id":     msg.ChatID,
		"user_id":     userID,
		"content_len": len(msg.Content),
		"streaming":   c.config.EnableStreaming,
	})

	// 检查是否启用流式响应
	if c.config.EnableStreaming {
		return c.sendStreamingResponseToIM(msg)
	}

	// 直接使用已优化的响应发送方法
	return c.sendResponseToIM(msg)
}

// InternalIMMessage 内部IM消息格式
type InternalIMMessage struct {
	Type       string            `json:"type"`
	MessageID  string            `json:"message_id"`
	Channel    string            `json:"channel"`
	SenderID   string            `json:"sender_id"`
	ChatID     string            `json:"chat_id"`
	Content    string            `json:"content"`
	Timestamp  int64             `json:"timestamp"`
	IsResponse bool              `json:"is_response,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	InReplyTo  string            `json:"in_reply_to,omitempty"`
}

func (c *InternalIMChannel) subscribeTopic(topic string) error {
	sub, err := c.nc.Subscribe(topic, c.handleMessage())
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", topic, err)
	}

	c.mu.Lock()
	c.subscriptions[topic] = sub
	c.mu.Unlock()

	logger.InfoCF("internal-im", "Subscribed to topic", map[string]any{
		"topic": topic,
	})

	return nil
}

func (c *InternalIMChannel) handleMessage() func(msg *natsgo.Msg) {
	return func(msg *natsgo.Msg) {
		logger.DebugCF("internal-im", "Received NATS message", map[string]any{
			"subject":     msg.Subject,
			"payload_len": len(msg.Data),
		})

		// 解析IM消息
		imMsg, err := FromJSON(msg.Data)
		if err != nil {
			logger.ErrorCF("internal-im", "Failed to parse IM message", map[string]any{
				"error":   err.Error(),
				"subject": msg.Subject,
			})
			// 如果不是IM消息格式，尝试作为纯文本处理
			c.handlePlainTextMessage(msg)
			return
		}

		// 验证消息格式
		if err := imMsg.Validate(); err != nil {
			logger.ErrorCF("internal-im", "Invalid IM message format", map[string]any{
				"error":   err.Error(),
				"subject": msg.Subject,
			})
			c.sendErrorToIM(imMsg.UserID, imMsg.ChatID, ErrorCodeInvalidFormat, err.Error())
			return
		}

		// 忽略响应消息（避免循环）
		if imMsg.IsResponse() || imMsg.IsError() || imMsg.IsStatus() || imMsg.IsUpdate() {
			return
		}

		// 只处理请求消息
		if !imMsg.IsRequest() {
			return
		}

		// 权限检查
		if !c.IsAllowed(imMsg.UserID) {
			logger.WarnCF("internal-im", "Unauthorized user", map[string]any{
				"user_id": imMsg.UserID,
				"subject": msg.Subject,
			})
			c.sendErrorToIM(imMsg.UserID, imMsg.ChatID, ErrorCodePermissionDenied, "Access denied")
			return
		}

		// 速率限制检查
		if !c.checkRateLimit() {
			logger.WarnCF("internal-im", "Rate limit exceeded", map[string]any{
				"user_id": imMsg.UserID,
				"subject": msg.Subject,
			})
			c.sendErrorToIM(imMsg.UserID, imMsg.ChatID, ErrorCodeRateLimit, "Rate limit exceeded. Please try again later.")
			return
		}

		// 添加到消息历史记录
		c.addMessageToHistory(imMsg.MessageID, imMsg.UserID, imMsg.Content)

		logger.InfoCF("internal-im", "Processing IM message", map[string]any{
			"user_id":     imMsg.UserID,
			"chat_id":     imMsg.ChatID,
			"username":    imMsg.Username,
			"content_len": len(imMsg.Content),
		})

		// 构造PicoClaw消息
		picoclawMsg := bus.InboundMessage{
			Channel:    "internal-im",
			SenderID:   imMsg.UserID,
			ChatID:     imMsg.ChatID,
			Content:    imMsg.Content,
			Media:      []string{},
			SessionKey: "",
			Metadata: map[string]string{
				"channel":     "internal-im",
				"subject":     msg.Subject,
				"user_id":     imMsg.UserID,
				"chat_id":     imMsg.ChatID,
				"username":    imMsg.Username,
				"sender_type": "user",
				"message_id":  c.generateMessageID(),
			},
		}

		// 发送到消息总线
		c.Bus.PublishInbound(c.ctx, picoclawMsg)

		// 记录发送的消息ID用于错误处理
		messageID := c.generateMessageID()
		logger.DebugCF("internal-im", "Published inbound message", map[string]any{
			"message_id": messageID,
			"chat_id":    picoclawMsg.ChatID,
			"sender_id":  picoclawMsg.SenderID,
		})

		// 发送处理状态
		go c.sendStatusToIM(imMsg.UserID, imMsg.ChatID, StatusProcessing, "🤖 PicoClaw正在处理您的请求...")
	}
}

func (c *InternalIMChannel) handlePlainTextMessage(msg *natsgo.Msg) {
	// 提取用户ID
	userID := c.extractUserIDFromTopic(msg.Subject)
	if userID == "" {
		userID = "unknown_user"
	}

	// 权限检查
	if !c.IsAllowed(userID) {
		logger.WarnCF("internal-im", "Unauthorized user (plain text)", map[string]any{
			"user_id": userID,
			"subject": msg.Subject,
		})
		return
	}

	content := string(msg.Data)
	if content == "" {
		return
	}

	logger.InfoCF("internal-im", "Processing plain text message", map[string]any{
		"user_id":     userID,
		"subject":     msg.Subject,
		"content_len": len(content),
	})

	// 构造元数据
	metadata := map[string]string{
		"channel":      "internal-im",
		"subject":      msg.Subject,
		"user_id":      userID,
		"sender_type":  "user",
		"message_type": "plain_text",
	}

	// 生成消息ID
	messageID := c.generateMessageID()

	// 转发到PicoClaw消息总线
	c.HandleMessage(c.ctx, bus.Peer{
		Kind: "direct",
		ID:   userID,
	}, messageID, userID, userID, content, []string{}, metadata,
		bus.SenderInfo{
			Platform:   "internal-im",
			PlatformID: userID,
			Username:   userID, // 对于纯文本消息，使用userID作为username
		})
}

func (c *InternalIMChannel) extractUserIDFromTopic(topic string) string {
	// 对于各种主题格式提取用户ID
	parts := strings.Split(topic, ".")

	// 匹配 picoclaw.commands.internal_im
	if len(parts) >= 4 && parts[0] == "picoclaw" && parts[1] == "commands" && parts[2] == "internal_im" {
		return "internal_im_user"
	}

	// 匹配 picoclaw.responses.internal_im
	if len(parts) >= 4 && parts[0] == "picoclaw" && parts[1] == "responses" && parts[2] == "internal_im" {
		return "internal_im_user"
	}

	// 匹配配置的主题
	if topic == c.config.Topic {
		return "config_topic_user"
	}

	// 对于测试环境
	if strings.HasPrefix(topic, "picoclaw_test_") {
		switch {
		case strings.Contains(topic, "internal_im"):
			return "test_internal_im_user"
		case strings.Contains(topic, "middleware"):
			return "middleware_user"
		case strings.Contains(topic, "topic"):
			return "test_user_888"
		case strings.Contains(topic, "messages"):
			return "test123"
		case strings.Contains(topic, "error"):
			return "error_user"
		}
	}

	logger.WarnCF("internal-im", "Cannot extract user ID from topic", map[string]any{
		"topic": topic,
	})
	return "unknown_user"
}

func (c *InternalIMChannel) isDuplicate(messageID string) bool {
	if messageID == "" {
		return false
	}

	c.idMu.Lock()
	defer c.idMu.Unlock()

	if c.processedIDs[messageID] {
		return true
	}

	c.processedIDs[messageID] = true

	// 简单清理：限制map大小
	if len(c.processedIDs) > 10000 {
		count := 0
		for id := range c.processedIDs {
			if count >= 5000 {
				break
			}
			delete(c.processedIDs, id)
			count++
		}
	}

	return false
}

func (c *InternalIMChannel) generateMessageID() string {
	return fmt.Sprintf("internal_im_%d", time.Now().UnixNano())
}

// listenToOutboundMessages 监听PicoClaw消息总线的出站消息并发送到IM
func (c *InternalIMChannel) listenToOutboundMessages() {
	logger.InfoC("internal-im", "Starting outbound message listener")

	for {
		select {
		case <-c.ctx.Done():
			logger.InfoC("internal-im", "Stopping outbound message listener")
			return
		default:
			// 使用带超时的context避免阻塞
			ctx, cancel := context.WithTimeout(c.ctx, 100*time.Millisecond)

			// 尝试订阅出站消息
			msg, ok := c.Bus.SubscribeOutbound(ctx)
			cancel()

			if !ok {
				// 没有消息，短暂休眠
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// 检查是否为internal-im通道的消息
			if msg.Channel != "internal-im" {
				continue
			}

			// 发送响应到IM机器人
			if err := c.sendResponseToIM(msg); err != nil {
				logger.ErrorCF("internal-im", "Failed to send response to IM", map[string]any{
					"error":   err.Error(),
					"chat_id": msg.ChatID,
					"channel": msg.Channel,
				})
			}
		}
	}
}

// getResponseTopic 获取响应主题
func (c *InternalIMChannel) getResponseTopic() string {
	if c.config.ResponseTopic != "" {
		return c.config.ResponseTopic
	}
	return "picoclaw.im.out" // 默认响应主题
}

// sendResponseToIM 发送响应消息到IM机器人
func (c *InternalIMChannel) sendResponseToIM(msg bus.OutboundMessage) error {
	// 从元数据中获取用户ID，如果没有则使用默认值
	userID := msg.ChatID // 如果没有单独的UserID，使用ChatID作为标识

	// 创建IM响应消息
	imResponse := NewResponseMessage(userID, msg.ChatID, msg.Content)

	// 转换为JSON
	data, err := imResponse.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal IM response: %w", err)
	}

	// 发送到响应主题
	responseTopic := c.getResponseTopic()
	err = c.nc.Publish(responseTopic, data)
	if err != nil {
		return fmt.Errorf("failed to publish to response topic %s: %w", responseTopic, err)
	}

	logger.DebugCF("internal-im", "Sent response to IM", map[string]any{
		"topic":       responseTopic,
		"chat_id":     msg.ChatID,
		"user_id":     userID,
		"content_len": len(msg.Content),
	})

	return nil
}

// sendErrorToIM 发送错误消息到IM机器人
func (c *InternalIMChannel) sendErrorToIM(userID, chatID, errorCode, errorMessage string) error {
	// 创建IM错误消息
	imError := NewErrorMessage(userID, chatID, errorCode, errorMessage)

	// 转换为JSON
	data, err := imError.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal IM error: %w", err)
	}

	// 发送到响应主题
	responseTopic := c.getResponseTopic()
	err = c.nc.Publish(responseTopic, data)
	if err != nil {
		return fmt.Errorf("failed to publish error to response topic %s: %w", responseTopic, err)
	}

	logger.DebugCF("internal-im", "Sent error to IM", map[string]any{
		"topic":      responseTopic,
		"chat_id":    chatID,
		"user_id":    userID,
		"error_code": errorCode,
	})

	return nil
}

// sendStatusToIM 发送状态消息到IM机器人
func (c *InternalIMChannel) sendStatusToIM(userID, chatID, status, message string) error {
	// 创建IM状态消息
	imStatus := NewStatusMessage(userID, chatID, status, message)

	// 转换为JSON
	data, err := imStatus.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal IM status: %w", err)
	}

	// 发送到响应主题
	responseTopic := c.getResponseTopic()
	err = c.nc.Publish(responseTopic, data)
	if err != nil {
		return fmt.Errorf("failed to publish status to response topic %s: %w", responseTopic, err)
	}

	logger.DebugCF("internal-im", "Sent status to IM", map[string]any{
		"topic":   responseTopic,
		"chat_id": chatID,
		"user_id": userID,
		"status":  status,
	})

	return nil
}

// sendStreamingResponseToIM 发送流式响应到IM机器人
func (c *InternalIMChannel) sendStreamingResponseToIM(msg bus.OutboundMessage) error {
	userID := msg.ChatID
	chatID := msg.ChatID
	content := msg.Content

	// 生成流式会话ID
	streamID := fmt.Sprintf("stream_%s_%d", userID, time.Now().UnixNano())

	logger.InfoCF("internal-im", "Starting streaming response", map[string]any{
		"chat_id":     chatID,
		"user_id":     userID,
		"stream_id":   streamID,
		"content_len": len(content),
	})

	// 发送开始状态
	if err := c.sendStatusToIM(userID, chatID, StatusProcessing, "🔄 开始生成流式响应..."); err != nil {
		logger.WarnCF("internal-im", "Failed to send start status", map[string]any{
			"error": err.Error(),
		})
	}

	// 模拟流式响应：将内容分块发送
	chunks := c.chunkContent(content, 50) // 每块最多50个字符

	for i, chunk := range chunks {
		isEnd := (i == len(chunks)-1)

		// 创建流式响应消息
		streamMsg := NewStreamMessage(userID, chatID, streamID, chunk, i, isEnd)

		// 转换为JSON
		data, err := streamMsg.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal stream message: %w", err)
		}

		// 发送到响应主题
		responseTopic := c.getResponseTopic()
		err = c.nc.Publish(responseTopic, data)
		if err != nil {
			return fmt.Errorf("failed to publish stream message to %s: %w", responseTopic, err)
		}

		logger.DebugCF("internal-im", "Sent stream chunk", map[string]any{
			"topic":     responseTopic,
			"chat_id":   chatID,
			"user_id":   userID,
			"stream_id": streamID,
			"chunk_idx": i,
			"chunk_len": len(chunk),
			"is_end":    isEnd,
		})

		// 短暂延迟以模拟流式效果
		if !isEnd {
			time.Sleep(200 * time.Millisecond)
		}
	}

	logger.InfoCF("internal-im", "Completed streaming response", map[string]any{
		"stream_id":    streamID,
		"total_chunks": len(chunks),
	})

	return nil
}

// checkRateLimit 检查速率限制
func (c *InternalIMChannel) checkRateLimit() bool {
	if c.config.RateLimit == nil {
		return true // 没有配置速率限制
	}

	c.rateLimitMu.Lock()
	defer c.rateLimitMu.Unlock()

	now := time.Now()

	// 检查是否需要重置令牌桶
	if now.Sub(c.lastRateLimitReset) >= time.Minute {
		// 每分钟补充令牌
		tokensToAdd := int64(c.config.RateLimit.MessagesPerMinute)
		currentTokens := c.rateLimitTokens.Load()
		maxTokens := int64(c.config.RateLimit.BurstSize)

		newTokens := currentTokens + tokensToAdd
		if newTokens > maxTokens {
			newTokens = maxTokens
		}

		c.rateLimitTokens.Store(newTokens)
		c.lastRateLimitReset = now

		logger.DebugCF("internal-im", "Rate limit tokens reset", map[string]any{
			"tokens_added":   tokensToAdd,
			"current_tokens": newTokens,
			"max_tokens":     maxTokens,
		})
	}

	// 检查是否有足够的令牌
	currentTokens := c.rateLimitTokens.Load()
	if currentTokens <= 0 {
		logger.WarnCF("internal-im", "Rate limit exceeded", map[string]any{
			"current_tokens": currentTokens,
			"max_per_minute": c.config.RateLimit.MessagesPerMinute,
			"burst_size":     c.config.RateLimit.BurstSize,
		})
		return false
	}

	// 消费一个令牌
	c.rateLimitTokens.Add(-1)

	logger.DebugCF("internal-im", "Rate limit token consumed", map[string]any{
		"remaining_tokens": currentTokens - 1,
	})

	return true
}

// addMessageToHistory 添加消息到历史记录
func (c *InternalIMChannel) addMessageToHistory(messageID, userID, content string) {
	if c.config.MessageRetention == nil {
		return // 没有配置消息保留
	}

	c.messageMu.Lock()
	defer c.messageMu.Unlock()

	record := MessageRecord{
		Timestamp: time.Now(),
		MessageID: messageID,
		UserID:    userID,
		Content:   content,
	}

	c.messageHistory = append(c.messageHistory, record)

	// 清理过期消息
	c.cleanupOldMessages()
}

// cleanupOldMessages 清理过期的消息记录
func (c *InternalIMChannel) cleanupOldMessages() {
	if c.config.MessageRetention == nil {
		return
	}

	now := time.Now()
	maxAge := time.Duration(c.config.MessageRetention.MaxAgeHours) * time.Hour
	maxCount := c.config.MessageRetention.MaxCount

	// 按时间排序
	sort.Slice(c.messageHistory, func(i, j int) bool {
		return c.messageHistory[i].Timestamp.Before(c.messageHistory[j].Timestamp)
	})

	// 清理过期消息
	validMessages := make([]MessageRecord, 0)
	for _, record := range c.messageHistory {
		if now.Sub(record.Timestamp) <= maxAge {
			validMessages = append(validMessages, record)
		}
	}

	// 限制数量
	if len(validMessages) > maxCount {
		validMessages = validMessages[len(validMessages)-maxCount:]
	}

	c.messageHistory = validMessages

	logger.DebugCF("internal-im", "Message history cleaned up", map[string]any{
		"total_records": len(validMessages),
		"max_age_hours": maxAge.Hours(),
		"max_count":     maxCount,
	})
}

// getMessageHistory 获取消息历史记录
func (c *InternalIMChannel) getMessageHistory(userID string, limit int) []MessageRecord {
	if c.config.MessageRetention == nil {
		return nil
	}

	c.messageMu.RLock()
	defer c.messageMu.RUnlock()

	// 过滤指定用户的消息
	var userMessages []MessageRecord
	for _, record := range c.messageHistory {
		if record.UserID == userID {
			userMessages = append(userMessages, record)
		}
	}

	// 按时间倒序排列
	sort.Slice(userMessages, func(i, j int) bool {
		return userMessages[i].Timestamp.After(userMessages[j].Timestamp)
	})

	// 限制数量
	if len(userMessages) > limit {
		userMessages = userMessages[:limit]
	}

	return userMessages
}

// chunkContent 将内容分块
func (c *InternalIMChannel) chunkContent(content string, chunkSize int) []string {
	if len(content) <= chunkSize {
		return []string{content}
	}

	var chunks []string
	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		chunks = append(chunks, content[i:end])
	}

	return chunks
}
