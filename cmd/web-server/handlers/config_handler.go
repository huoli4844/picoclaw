package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sipeed/picoclaw/cmd/web-server/models"
	"github.com/sipeed/picoclaw/pkg/config"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	config *config.Config
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{
		config: cfg,
	}
}

// GetConfig 获取配置
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := models.ConfigResponse{
		ModelList: make([]models.ModelConfig, 0),
	}

	if h.config != nil && h.config.ModelList != nil {
		for _, model := range h.config.ModelList {
			modelConfig := models.ModelConfig{
				ModelName: model.ModelName,
				Model:     model.Model,
				// 不返回 API Key 到前端
			}
			if model.APIBase != "" {
				modelConfig.APIBase = model.APIBase
			}
			response.ModelList = append(response.ModelList, modelConfig)
		}
	}

	if h.config != nil {
		response.Agents.Defaults.Model = h.config.Agents.Defaults.Model
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateConfig 更新配置
func (h *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ConfigResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 更新配置
	if h.config == nil {
		h.config = &config.Config{}
	}

	h.config.ModelList = make([]config.ModelConfig, 0)
	for _, model := range req.ModelList {
		modelConfig := config.ModelConfig{
			ModelName: model.ModelName,
			Model:     model.Model,
			APIKey:    model.APIKey,
		}
		if model.APIBase != "" {
			modelConfig.APIBase = model.APIBase
		}
		h.config.ModelList = append(h.config.ModelList, modelConfig)
	}

	// 更新默认模型
	if len(req.Agents.Defaults.Model) > 0 {
		h.config.Agents.Defaults.Model = req.Agents.Defaults.Model
	}

	// 保存配置
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".picoclaw", "config.json")

	configData, err := json.MarshalIndent(h.config, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal config: %v", err), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// GetModels 获取模型列表
func (h *ConfigHandler) GetModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	modelList := make([]models.ModelConfig, 0)

	if h.config != nil && h.config.ModelList != nil {
		for _, model := range h.config.ModelList {
			modelConfig := models.ModelConfig{
				ModelName: model.ModelName,
				Model:     model.Model,
			}
			if model.APIBase != "" {
				modelConfig.APIBase = model.APIBase
			}
			modelList = append(modelList, modelConfig)
		}
	}

	json.NewEncoder(w).Encode(modelList)
}
