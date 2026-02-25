package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/cmd/web-server/models"
)

// ConversationService 对话服务
type ConversationService struct {
	conversations    map[string]*models.Conversation
	conversationsMu  sync.RWMutex
	conversationsDir string
}

// NewConversationService 创建对话服务
func NewConversationService() (*ConversationService, error) {
	service := &ConversationService{
		conversations: make(map[string]*models.Conversation),
	}

	// 获取用户主目录
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	// 设置跨平台的对话存储路径
	service.conversationsDir = filepath.Join(home, ".picoclaw", "workspace", "chat")
	log.Printf("Chat storage directory: %s", service.conversationsDir)

	// 创建对话存储目录
	if err := os.MkdirAll(service.conversationsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create conversations directory: %v", err)
	}
	log.Printf("Chat storage directory created/verified successfully")

	// 加载现有对话
	if err := service.loadConversations(); err != nil {
		return nil, fmt.Errorf("failed to load conversations: %v", err)
	}

	return service, nil
}

// loadConversations 加载对话历史
func (s *ConversationService) loadConversations() error {
	files, err := filepath.Glob(filepath.Join(s.conversationsDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to read conversation files: %v", err)
	}

	s.conversationsMu.Lock()
	defer s.conversationsMu.Unlock()

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Warning: Failed to read conversation file %s: %v", file, err)
			continue
		}

		var conv models.Conversation
		if err := json.Unmarshal(data, &conv); err != nil {
			log.Printf("Warning: Failed to parse conversation file %s: %v", file, err)
			continue
		}

		s.conversations[conv.ID] = &conv
	}

	log.Printf("Loaded %d conversations", len(s.conversations))
	return nil
}

// saveConversation 保存对话
func (s *ConversationService) saveConversation(conv *models.Conversation) error {
	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %v", err)
	}

	filename := filepath.Join(s.conversationsDir, conv.ID+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to save conversation file: %v", err)
	}

	return nil
}

// SaveUserMessage 保存用户消息到对话历史
func (s *ConversationService) SaveUserMessage(conversationID, message, model string) {
	s.conversationsMu.Lock()
	defer s.conversationsMu.Unlock()

	conv, exists := s.conversations[conversationID]
	if !exists {
		log.Printf("Conversation %s not found, skipping user message save", conversationID)
		return
	}

	// 创建用户消息
	userMessage := models.ConversationMessage{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Content:   message,
		Role:      "user",
		Timestamp: time.Now(),
		Model:     model,
	}

	// 添加到对话消息列表
	conv.Messages = append(conv.Messages, userMessage)
	conv.UpdatedAt = time.Now()

	// 保存到文件
	if err := s.saveConversation(conv); err != nil {
		log.Printf("Failed to save user message to conversation %s: %v", conversationID, err)
	} else {
		log.Printf("Saved user message to conversation %s", conversationID)
	}
}

// SaveAssistantMessage 保存助手消息到对话历史
func (s *ConversationService) SaveAssistantMessage(conversationID, message, model string, thoughts []models.Thought) {
	s.conversationsMu.Lock()
	defer s.conversationsMu.Unlock()

	conv, exists := s.conversations[conversationID]
	if !exists {
		log.Printf("Conversation %s not found, skipping assistant message save", conversationID)
		return
	}

	// 创建助手消息
	assistantMessage := models.ConversationMessage{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Content:   message,
		Role:      "assistant",
		Timestamp: time.Now(),
		Model:     model,
		Thoughts:  thoughts,
	}

	// 添加到对话消息列表
	conv.Messages = append(conv.Messages, assistantMessage)
	conv.UpdatedAt = time.Now()

	// 保存到文件
	if err := s.saveConversation(conv); err != nil {
		log.Printf("Failed to save assistant message to conversation %s: %v", conversationID, err)
	} else {
		log.Printf("Saved assistant message to conversation %s", conversationID)
	}
}

// GetConversations 获取对话列表
func (s *ConversationService) GetConversations() []*models.Conversation {
	s.conversationsMu.RLock()
	defer s.conversationsMu.RUnlock()

	// 转换为切片并按更新时间排序
	convList := make([]*models.Conversation, 0, len(s.conversations))
	for _, conv := range s.conversations {
		convList = append(convList, conv)
	}

	// 按更新时间降序排序
	for i := 0; i < len(convList); i++ {
		for j := i + 1; j < len(convList); j++ {
			if convList[i].UpdatedAt.Before(convList[j].UpdatedAt) {
				convList[i], convList[j] = convList[j], convList[i]
			}
		}
	}

	return convList
}

// GetConversation 获取特定对话
func (s *ConversationService) GetConversation(id string) (*models.Conversation, error) {
	s.conversationsMu.RLock()
	conv, exists := s.conversations[id]
	s.conversationsMu.RUnlock()

	if !exists {
		// 如果内存中没有，尝试从文件加载
		filename := filepath.Join(s.conversationsDir, id+".json")
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("conversation not found")
		}

		var fileConv models.Conversation
		if err := json.Unmarshal(data, &fileConv); err != nil {
			return nil, fmt.Errorf("failed to parse conversation file")
		}

		// 加载到内存
		s.conversationsMu.Lock()
		s.conversations[id] = &fileConv
		s.conversationsMu.Unlock()

		conv = &fileConv
	}

	return conv, nil
}

// CreateConversation 创建新对话
func (s *ConversationService) CreateConversation(title, model string) (*models.Conversation, error) {
	// 生成唯一ID
	id := fmt.Sprintf("conv_%d", time.Now().UnixNano())

	// 如果没有提供标题，使用默认标题
	if title == "" {
		title = "新对话 " + time.Now().Format("2006-01-02 15:04:05")
	}

	conv := &models.Conversation{
		ID:        id,
		Title:     title,
		Messages:  make([]models.ConversationMessage, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Model:     model,
	}

	// 保存到内存和文件
	s.conversationsMu.Lock()
	s.conversations[id] = conv
	s.conversationsMu.Unlock()

	if err := s.saveConversation(conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation")
	}

	return conv, nil
}

// UpdateConversation 更新对话
func (s *ConversationService) UpdateConversation(id string, req *models.UpdateConversationRequest) (*models.Conversation, error) {
	s.conversationsMu.Lock()
	defer s.conversationsMu.Unlock()

	conv, exists := s.conversations[id]
	if !exists {
		return nil, fmt.Errorf("conversation not found")
	}

	// 更新字段
	if req.Title != "" {
		conv.Title = req.Title
	}
	if req.Messages != nil {
		conv.Messages = req.Messages
	}
	conv.UpdatedAt = time.Now()

	if err := s.saveConversation(conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation")
	}

	return conv, nil
}

// DeleteConversation 删除对话
func (s *ConversationService) DeleteConversation(id string) (*models.Conversation, error) {
	s.conversationsMu.Lock()
	defer s.conversationsMu.Unlock()

	conv, exists := s.conversations[id]
	if !exists {
		return nil, fmt.Errorf("conversation not found")
	}

	// 删除文件
	filename := filepath.Join(s.conversationsDir, id+".json")
	if err := os.Remove(filename); err != nil {
		log.Printf("Warning: Failed to delete conversation file %s: %v", filename, err)
	}

	// 从内存中删除
	delete(s.conversations, id)

	return conv, nil
}
