package channels

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// IMMessage 统一的IM消息协议
type IMMessage struct {
	Type      string    `json:"type"` // message, response, error, status, update
	UserID    string    `json:"user_id"`
	ChatID    string    `json:"chat_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	MessageID string    `json:"message_id"` // 消息唯一标识

	// 可选字段
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	Status       string `json:"status,omitempty"`
	Progress     int    `json:"progress,omitempty"`
	StreamID     string `json:"stream_id,omitempty"`     // 流式响应会话ID
	IsStreamEnd  bool   `json:"is_stream_end,omitempty"` // 是否为流式响应的结束
	ChunkIndex   int    `json:"chunk_index,omitempty"`   // 流式响应块索引
}

// MessageTypes 支持的消息类型
const (
	MessageTypeRequest  = "message"
	MessageTypeResponse = "response"
	MessageTypeError    = "error"
	MessageTypeStatus   = "status"
	MessageTypeUpdate   = "update"
)

// StatusTypes 支持的状态类型
const (
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusQueued     = "queued"
)

// ErrorCodes 支持的错误代码
const (
	ErrorCodePermissionDenied = "PERMISSION_DENIED"
	ErrorCodeInvalidFormat    = "INVALID_FORMAT"
	ErrorCodeTimeout          = "TIMEOUT"
	ErrorCodeInternalError    = "INTERNAL_ERROR"
	ErrorCodeRateLimit        = "RATE_LIMIT"
)

// NewRequestMessage 创建请求消息
func NewRequestMessage(userID, chatID, username, content string) *IMMessage {
	return &IMMessage{
		Type:      MessageTypeRequest,
		UserID:    userID,
		ChatID:    chatID,
		Username:  username,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// NewResponseMessage 创建响应消息
func NewResponseMessage(userID, chatID, content string) *IMMessage {
	return &IMMessage{
		Type:      MessageTypeResponse,
		UserID:    userID,
		ChatID:    chatID,
		Username:  "PicoClaw",
		Content:   content,
		Timestamp: time.Now(),
	}
}

// NewErrorMessage 创建错误消息
func NewErrorMessage(userID, chatID, errorCode, errorMessage string) *IMMessage {
	return &IMMessage{
		Type:         MessageTypeError,
		UserID:       userID,
		ChatID:       chatID,
		Username:     "PicoClaw",
		Content:      errorMessage,
		Timestamp:    time.Now(),
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
}

// NewStatusMessage 创建状态消息
func NewStatusMessage(userID, chatID, status, message string) *IMMessage {
	return &IMMessage{
		Type:      MessageTypeStatus,
		UserID:    userID,
		ChatID:    chatID,
		Username:  "PicoClaw",
		Content:   message,
		Timestamp: time.Now(),
		Status:    status,
	}
}

// NewUpdateMessage 创建更新消息
func NewUpdateMessage(userID, chatID, content string, progress int) *IMMessage {
	return &IMMessage{
		Type:      MessageTypeUpdate,
		UserID:    userID,
		ChatID:    chatID,
		Username:  "PicoClaw",
		Content:   content,
		Timestamp: time.Now(),
		Progress:  progress,
	}
}

// NewStreamMessage 创建流式响应消息
func NewStreamMessage(userID, chatID, streamID, content string, chunkIndex int, isEnd bool) *IMMessage {
	return &IMMessage{
		Type:        MessageTypeResponse,
		UserID:      userID,
		ChatID:      chatID,
		Username:    "PicoClaw",
		Content:     content,
		Timestamp:   time.Now(),
		StreamID:    streamID,
		ChunkIndex:  chunkIndex,
		IsStreamEnd: isEnd,
	}
}

// ToJSON 将消息转换为JSON
func (m *IMMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON 从JSON创建消息
func FromJSON(data []byte) (*IMMessage, error) {
	var msg IMMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}

	// 如果没有MessageID，自动生成一个
	if msg.MessageID == "" {
		msg.MessageID = generateMessageID(msg)
	}

	return &msg, nil
}

// generateMessageID 生成消息唯一标识
func generateMessageID(msg IMMessage) string {
	content := fmt.Sprintf("%s_%s_%s_%d_%s",
		msg.Type, msg.UserID, msg.ChatID, msg.Timestamp.UnixNano(), msg.Content)
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])[:16] // 取前16位
}

// Validate 验证消息格式
func (m *IMMessage) Validate() error {
	if m.Type == "" {
		return ErrInvalidMessageType
	}
	if m.UserID == "" {
		return ErrMissingUserID
	}
	if m.ChatID == "" {
		return ErrMissingChatID
	}
	if m.Username == "" {
		return ErrMissingUsername
	}

	// 验证消息类型
	switch m.Type {
	case MessageTypeRequest, MessageTypeResponse, MessageTypeError, MessageTypeStatus, MessageTypeUpdate:
		// 有效的类型
	default:
		return ErrInvalidMessageType
	}

	return nil
}

// IsRequest 判断是否为请求消息
func (m *IMMessage) IsRequest() bool {
	return m.Type == MessageTypeRequest
}

// IsResponse 判断是否为响应消息
func (m *IMMessage) IsResponse() bool {
	return m.Type == MessageTypeResponse
}

// IsError 判断是否为错误消息
func (m *IMMessage) IsError() bool {
	return m.Type == MessageTypeError
}

// IsStatus 判断是否为状态消息
func (m *IMMessage) IsStatus() bool {
	return m.Type == MessageTypeStatus
}

// IsUpdate 判断是否为更新消息
func (m *IMMessage) IsUpdate() bool {
	return m.Type == MessageTypeUpdate
}

// 错误定义
var (
	ErrInvalidMessageType = &ValidationError{"invalid message type"}
	ErrMissingUserID      = &ValidationError{"missing user_id"}
	ErrMissingChatID      = &ValidationError{"missing chat_id"}
	ErrMissingUsername    = &ValidationError{"missing username"}
)

// ValidationError 验证错误
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
