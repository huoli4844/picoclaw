package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sipeed/picoclaw/cmd/web-server/models"
	"github.com/sipeed/picoclaw/cmd/web-server/services"
)

// ConversationHandler 对话历史处理器
type ConversationHandler struct {
	conversationService *services.ConversationService
}

// NewConversationHandler 创建对话历史处理器
func NewConversationHandler(convSvc *services.ConversationService) *ConversationHandler {
	return &ConversationHandler{
		conversationService: convSvc,
	}
}

// GetConversations 获取对话列表
func (h *ConversationHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	convList := h.conversationService.GetConversations()
	json.NewEncoder(w).Encode(convList)
}

// CreateConversation 创建新对话
func (h *ConversationHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conv, err := h.conversationService.CreateConversation(req.Title, req.Model)
	if err != nil {
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(conv)
}

// GetConversation 获取特定对话
func (h *ConversationHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		http.Error(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}

	conv, err := h.conversationService.GetConversation(id)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(conv)
}

// UpdateConversation 更新对话
func (h *ConversationHandler) UpdateConversation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		http.Error(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conv, err := h.conversationService.UpdateConversation(id, &req)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(conv)
}

// DeleteConversation 删除对话
func (h *ConversationHandler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		http.Error(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}

	conv, err := h.conversationService.DeleteConversation(id)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// 返回成功响应
	result := map[string]interface{}{
		"success": true,
		"message": "Conversation deleted successfully",
		"deleted": conv,
	}

	json.NewEncoder(w).Encode(result)
}
