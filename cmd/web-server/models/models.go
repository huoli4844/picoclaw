package models

import "time"

// ChatRequest 聊天请求
type ChatRequest struct {
	Message        string `json:"message"`
	Model          string `json:"model"`
	Stream         bool   `json:"stream"`
	ConversationID string `json:"conversationId,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Message   string    `json:"message"`
	Model     string    `json:"model"`
	Timestamp time.Time `json:"timestamp"`
	Thoughts  []Thought `json:"thoughts,omitempty"`
}

// Thought AI思考过程
type Thought struct {
	Type      string    `json:"type"` // "tool_call", "tool_result", "thinking"
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
	ToolName  string    `json:"tool_name,omitempty"`
	Args      string    `json:"args,omitempty"`
	Result    string    `json:"result,omitempty"`
	Duration  int       `json:"duration,omitempty"`  // in milliseconds
	Iteration int       `json:"iteration,omitempty"` // LLM iteration number
}

// ConversationMessage 对话消息
type ConversationMessage struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Role      string    `json:"role"` // "user" | "assistant"
	Timestamp time.Time `json:"timestamp"`
	Model     string    `json:"model,omitempty"`
	Thoughts  []Thought `json:"thoughts,omitempty"`
}

// Conversation 对话历史
type Conversation struct {
	ID        string                `json:"id"`
	Title     string                `json:"title"`
	Messages  []ConversationMessage `json:"messages"`
	CreatedAt time.Time             `json:"createdAt"`
	UpdatedAt time.Time             `json:"updatedAt"`
	Model     string                `json:"model"`
}

// CreateConversationRequest 创建对话请求
type CreateConversationRequest struct {
	Title string `json:"title,omitempty"`
	Model string `json:"model"`
}

// UpdateConversationRequest 更新对话请求
type UpdateConversationRequest struct {
	Title    string                `json:"title,omitempty"`
	Messages []ConversationMessage `json:"messages,omitempty"`
}

// ConfigResponse 配置响应
type ConfigResponse struct {
	ModelList []ModelConfig `json:"model_list"`
	Agents    struct {
		Defaults struct {
			Model string `json:"model"`
		} `json:"defaults"`
	} `json:"agents"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	ModelName string `json:"model_name"`
	Model     string `json:"model"`
	APIKey    string `json:"api_key,omitempty"`
	APIBase   string `json:"api_base,omitempty"`
}

// SkillInfo 技能信息
type SkillInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

// SkillDetail 技能详情
type SkillDetail struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Source      string            `json:"source"`
	Description string            `json:"description"`
	Content     string            `json:"content"`
	Metadata    map[string]string `json:"metadata"`
}

// SearchSkillsRequest 搜索技能请求
type SearchSkillsRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// InstallSkillRequest 安装技能请求
type InstallSkillRequest struct {
	Slug     string `json:"slug"`
	Registry string `json:"registry"`
	Version  string `json:"version,omitempty"`
	Force    bool   `json:"force,omitempty"`
}

// FileInfo 文件信息
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	IsDir   bool      `json:"isDir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

// FileListResponse 文件列表响应
type FileListResponse struct {
	Success bool       `json:"success"`
	Path    string     `json:"path"`
	Files   []FileInfo `json:"files"`
}

// FileContentResponse 文件内容响应
type FileContentResponse struct {
	Success     bool   `json:"success"`
	Path        string `json:"path"`
	Content     string `json:"content"`
	Size        int    `json:"size"`
	ContentType string `json:"contentType"`
}

// FileDeleteRequest 删除文件请求
type FileDeleteRequest struct {
	Path string `json:"path"`
}

// FileDeleteResponse 删除文件响应
type FileDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Path    string `json:"path"`
	IsDir   bool   `json:"isDir"`
}

// SkillInstallerRequest 技能安装器请求
type SkillInstallerRequest struct {
	Slug     string `json:"slug"`
	Registry string `json:"registry"`
	Version  string `json:"version,omitempty"`
	Force    bool   `json:"force,omitempty"`
}

// SkillInstallerResponse 技能安装器响应
type SkillInstallerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}
