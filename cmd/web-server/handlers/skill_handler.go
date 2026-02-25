package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sipeed/picoclaw/cmd/web-server/models"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/skills"
	"github.com/sipeed/picoclaw/pkg/tools"
)

// SkillHandler 技能处理器
type SkillHandler struct {
	skillsLoader    *skills.SkillsLoader
	skillsWorkspace string
	config          *config.Config
}

// NewSkillHandler 创建技能处理器
func NewSkillHandler(skillsLoader *skills.SkillsLoader, skillsWorkspace string, cfg *config.Config) *SkillHandler {
	return &SkillHandler{
		skillsLoader:    skillsLoader,
		skillsWorkspace: skillsWorkspace,
		config:          cfg,
	}
}

// GetSkills 获取技能列表
func (h *SkillHandler) GetSkills(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.skillsLoader == nil {
		http.Error(w, "Skills loader not initialized", http.StatusInternalServerError)
		return
	}

	skillsList := h.skillsLoader.ListSkills()
	result := make([]models.SkillInfo, 0, len(skillsList))

	for _, skill := range skillsList {
		result = append(result, models.SkillInfo{
			Name:        skill.Name,
			Path:        skill.Path,
			Source:      skill.Source,
			Description: skill.Description,
		})
	}

	json.NewEncoder(w).Encode(result)
}

// GetSkillDetail 获取技能详情
func (h *SkillHandler) GetSkillDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	skillName := vars["name"]

	if h.skillsLoader == nil {
		http.Error(w, "Skills loader not initialized", http.StatusInternalServerError)
		return
	}

	skillContent, exists := h.skillsLoader.LoadSkill(skillName)
	if !exists {
		http.Error(w, "Skill not found", http.StatusNotFound)
		return
	}

	// 获取技能信息
	skillsList := h.skillsLoader.ListSkills()
	var skillInfo *skills.SkillInfo
	for _, skill := range skillsList {
		if skill.Name == skillName {
			skillInfo = &skill
			break
		}
	}

	if skillInfo == nil {
		http.Error(w, "Skill not found", http.StatusNotFound)
		return
	}

	// 解析元数据
	metadata := make(map[string]string)
	if skillContent != "" {
		lines := strings.Split(skillContent, "\n")
		inFrontmatter := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "---" {
				if !inFrontmatter {
					inFrontmatter = true
					continue
				} else {
					break
				}
			}
			if inFrontmatter && strings.Contains(trimmed, ":") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					metadata[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	detail := models.SkillDetail{
		Name:        skillInfo.Name,
		Path:        skillInfo.Path,
		Source:      skillInfo.Source,
		Description: skillInfo.Description,
		Content:     skillContent,
		Metadata:    metadata,
	}

	json.NewEncoder(w).Encode(detail)
}

// UninstallSkill 卸载技能
func (h *SkillHandler) UninstallSkill(w http.ResponseWriter, r *http.Request) {
	log.Printf("Uninstall skill request received: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	skillName := vars["name"]
	log.Printf("Skill name from URL: %s", skillName)

	if h.skillsLoader == nil {
		log.Printf("Skills loader not initialized")
		http.Error(w, "Skills loader not initialized", http.StatusInternalServerError)
		return
	}

	// 检查技能是否存在
	skillsList := h.skillsLoader.ListSkills()
	log.Printf("Available skills: %v", skillsList)
	var skillInfo *skills.SkillInfo
	for _, skill := range skillsList {
		if skill.Name == skillName {
			skillInfo = &skill
			break
		}
	}

	if skillInfo == nil {
		log.Printf("Skill not found: %s", skillName)
		http.Error(w, "Skill not found", http.StatusNotFound)
		return
	}

	// 不能删除内置技能
	if skillInfo.Source == "builtin" {
		log.Printf("Attempt to delete builtin skill: %s", skillName)
		http.Error(w, "Cannot delete builtin skills", http.StatusForbidden)
		return
	}

	// 使用技能安装器删除技能
	skillInstaller := skills.NewSkillInstaller(h.skillsWorkspace)
	if err := skillInstaller.Uninstall(skillName); err != nil {
		log.Printf("Failed to uninstall skill %s: %v", skillName, err)
		http.Error(w, fmt.Sprintf("Failed to uninstall skill: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully uninstalled skill: %s", skillName)

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Skill %s uninstalled successfully", skillName),
	}

	json.NewEncoder(w).Encode(response)
}

// SearchSkills 搜索技能
func (h *SkillHandler) SearchSkills(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.SearchSkillsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if h.config == nil || h.config.Tools.Skills.Registries.ClawHub.Enabled == false {
		http.Error(w, "Skills registry not configured", http.StatusNotFound)
		return
	}

	// 使用工具搜索技能
	clawHubConfig := h.config.Tools.Skills.Registries.ClawHub
	registryConfig := skills.RegistryConfig{
		ClawHub: skills.ClawHubConfig{
			Enabled:         clawHubConfig.Enabled,
			BaseURL:         clawHubConfig.BaseURL,
			AuthToken:       clawHubConfig.AuthToken,
			SearchPath:      clawHubConfig.SearchPath,
			SkillsPath:      clawHubConfig.SkillsPath,
			DownloadPath:    clawHubConfig.DownloadPath,
			Timeout:         clawHubConfig.Timeout,
			MaxZipSize:      clawHubConfig.MaxZipSize,
			MaxResponseSize: clawHubConfig.MaxResponseSize,
		},
		MaxConcurrentSearches: h.config.Tools.Skills.MaxConcurrentSearches,
	}
	registryMgr := skills.NewRegistryManagerFromConfig(registryConfig)

	if req.Limit == 0 {
		req.Limit = 10
	}

	// 使用缓存机制，与tool保持一致
	cache := skills.NewSearchCache(h.config.Tools.Skills.SearchCache.MaxSize, time.Duration(h.config.Tools.Skills.SearchCache.TTLSeconds)*time.Second)
	findSkillTool := tools.NewFindSkillsTool(registryMgr, cache)

	ctx := context.Background()
	result := findSkillTool.Execute(ctx, map[string]interface{}{
		"query": req.Query,
		"limit": req.Limit,
	})

	// 检查缓存或直接搜索
	var results []skills.SearchResult
	var err error

	if !result.IsError {
		// 先检查缓存
		if cache != nil {
			if cached, hit := cache.Get(req.Query); hit {
				results = cached
			} else {
				// 缓存未命中，直接搜索
				results, err = registryMgr.SearchAll(ctx, req.Query, req.Limit)
				if err == nil && len(results) > 0 {
					cache.Put(req.Query, results)
				}
			}
		} else {
			// 没有缓存，直接搜索
			results, err = registryMgr.SearchAll(ctx, req.Query, req.Limit)
		}
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 转换为前端期望的格式
	searchResults := make([]interface{}, len(results))
	for i, r := range results {
		searchResults[i] = map[string]interface{}{
			"slug":          r.Slug,
			"display_name":  r.DisplayName,
			"summary":       r.Summary,
			"version":       r.Version,
			"registry_name": r.RegistryName,
			"score":         r.Score,
		}
	}

	response := map[string]interface{}{
		"query":   req.Query,
		"results": searchResults,
	}

	json.NewEncoder(w).Encode(response)
}

// InstallSkill 安装技能
func (h *SkillHandler) InstallSkill(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.InstallSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid install request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Installing skill: slug=%s, registry=%s, version=%s", req.Slug, req.Registry, req.Version)

	// 检查配置
	if h.config == nil {
		log.Printf("Configuration is nil")
		http.Error(w, "Configuration not loaded", http.StatusInternalServerError)
		return
	}

	if !h.config.Tools.Skills.Registries.ClawHub.Enabled {
		log.Printf("ClawHub registry is not enabled")
		http.Error(w, "ClawHub registry not enabled", http.StatusServiceUnavailable)
		return
	}

	// 使用工具安装技能
	clawHubConfig := h.config.Tools.Skills.Registries.ClawHub
	log.Printf("ClawHub config: BaseURL=%s, Enabled=%v", clawHubConfig.BaseURL, clawHubConfig.Enabled)
	registryConfig := skills.RegistryConfig{
		ClawHub: skills.ClawHubConfig{
			Enabled:         clawHubConfig.Enabled,
			BaseURL:         clawHubConfig.BaseURL,
			AuthToken:       clawHubConfig.AuthToken,
			SearchPath:      clawHubConfig.SearchPath,
			SkillsPath:      clawHubConfig.SkillsPath,
			DownloadPath:    clawHubConfig.DownloadPath,
			Timeout:         clawHubConfig.Timeout,
			MaxZipSize:      clawHubConfig.MaxZipSize,
			MaxResponseSize: clawHubConfig.MaxResponseSize,
		},
		MaxConcurrentSearches: h.config.Tools.Skills.MaxConcurrentSearches,
	}
	registryMgr := skills.NewRegistryManagerFromConfig(registryConfig)
	installSkillTool := tools.NewInstallSkillTool(registryMgr, h.skillsWorkspace)

	installParams := map[string]interface{}{
		"slug":     req.Slug,
		"registry": req.Registry,
	}

	if req.Version != "" {
		installParams["version"] = req.Version
	}
	if req.Force {
		installParams["force"] = req.Force
	}

	ctx := context.Background()
	result := installSkillTool.Execute(ctx, installParams)

	var status string
	var message string
	if !result.IsError {
		status = "success"
		message = "Skill installed successfully"
		log.Printf("Skill installation successful: %s", req.Slug)
	} else {
		status = "error"
		message = result.ForLLM
		log.Printf("Skill installation failed: %s, error: %s", req.Slug, result.ForLLM)
	}

	response := models.SkillInstallerResponse{
		Status:  status,
		Message: message,
		Result:  result.ForLLM,
	}

	log.Printf("Sending install response: %v", response)
	json.NewEncoder(w).Encode(response)
}
