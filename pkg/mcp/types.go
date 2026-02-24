package mcp

import (
	"time"
)

// MCPServer represents a Model Context Protocol server
type MCPServer struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Author      string            `json:"author,omitempty"`
	Homepage    string            `json:"homepage,omitempty"`
	Repository  string            `json:"repository,omitempty"`
	License     string            `json:"license,omitempty"`
	Keywords    []string          `json:"keywords,omitempty"`
	Category    string            `json:"category,omitempty"`
	Transport   string            `json:"transport"` // stdio, sse, websocket
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Status      string            `json:"status"` // installed, available, error
	Config      map[string]any    `json:"config,omitempty"`
	Tools       []MCPTool         `json:"tools,omitempty"`
	Resources   []MCPResource     `json:"resources,omitempty"`
	InstalledAt *time.Time        `json:"installed_at,omitempty"`
}

// MCPTool represents a tool provided by an MCP server
type MCPTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema,omitempty"`
	ServerID    string         `json:"serverId"`
	Category    string         `json:"category,omitempty"`
}

// MCPResource represents a resource provided by an MCP server
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	ServerID    string `json:"serverId"`
}

// MCPSearchRequest represents a search request for MCP servers
type MCPSearchRequest struct {
	Query     string `json:"query"`
	Category  string `json:"category,omitempty"`
	Transport string `json:"transport,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// MCPSearchResponse represents the response from MCP server search
type MCPSearchResponse struct {
	Query   string      `json:"query"`
	Results []MCPServer `json:"results"`
	Total   int         `json:"total"`
	Offset  int         `json:"offset"`
	Limit   int         `json:"limit"`
}

// MCPInstallRequest represents a request to install an MCP server
type MCPInstallRequest struct {
	ServerID string         `json:"serverId"`
	Config   map[string]any `json:"config,omitempty"`
}

// MCPInstallResponse represents the response from MCP server installation
type MCPInstallResponse struct {
	Status  string     `json:"status"` // success, error
	Message string     `json:"message"`
	Server  *MCPServer `json:"server,omitempty"`
}

// MCPServerConfig represents the configuration for an MCP server
type MCPServerConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Transport   string            `json:"transport"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Config      map[string]any    `json:"config,omitempty"`
	Enabled     bool              `json:"enabled"`
}

// MCPRegistryConfig represents the configuration for MCP registries
type MCPRegistryConfig struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Enabled  bool   `json:"enabled"`
	APIKey   string `json:"apiKey,omitempty"`
	Priority int    `json:"priority"`
}

// MCPConfig represents the MCP configuration
type MCPConfig struct {
	Enabled     bool                `json:"enabled"`
	Servers     []MCPServerConfig   `json:"servers"`
	Registries  []MCPRegistryConfig `json:"registries"`
	StoragePath string              `json:"storage_path"`
}

// Known MCP servers from official registries
var KnownMCPServers = []MCPServer{
	{
		ID:          "filesystem",
		Name:        "Filesystem MCP Server",
		Description: "提供文件系统操作工具，包括文件读写、目录管理、文件搜索等功能",
		Version:     "1.0.0",
		Author:      "Model Context Protocol",
		Homepage:    "https://github.com/modelcontextprotocol/servers",
		Repository:  "https://github.com/modelcontextprotocol/servers",
		License:     "MIT",
		Keywords:    []string{"filesystem", "files", "directory", "io"},
		Category:    "filesystem",
		Transport:   "stdio",
		Command:     "mcp-server-filesystem",
		Args:        []string{"/Users/huoli4844/Documents/ai_project/picoclaw"},
		Env:         make(map[string]string),
		Status:      "available",
		Tools: []MCPTool{
			{
				Name:        "read_file",
				Description: "读取文件内容",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "文件路径",
						},
					},
					"required": []string{"path"},
				},
				ServerID: "filesystem",
			},
			{
				Name:        "write_file",
				Description: "写入文件内容",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "文件路径",
						},
						"content": map[string]any{
							"type":        "string",
							"description": "文件内容",
						},
					},
					"required": []string{"path", "content"},
				},
				ServerID: "filesystem",
			},
			{
				Name:        "list_directory",
				Description: "列出目录内容",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "目录路径",
						},
					},
					"required": []string{"path"},
				},
				ServerID: "filesystem",
			},
		},
	},
	{
		ID:          "git",
		Name:        "Git MCP Server",
		Description: "提供Git版本控制工具，支持仓库操作、提交管理、分支管理等功能",
		Version:     "1.2.0",
		Author:      "Model Context Protocol",
		Homepage:    "https://github.com/modelcontextprotocol/servers",
		Repository:  "https://github.com/modelcontextprotocol/servers",
		License:     "MIT",
		Keywords:    []string{"git", "version control", "repository", "commit"},
		Category:    "development",
		Transport:   "stdio",
		Command:     "npx",
		Args:        []string{"@modelcontextprotocol/server-git"},
		Env:         make(map[string]string),
		Status:      "available",
		Tools: []MCPTool{
			{
				Name:        "git_status",
				Description: "查看Git仓库状态",
				ServerID:    "git",
			},
			{
				Name:        "git_add",
				Description: "添加文件到暂存区",
				ServerID:    "git",
			},
			{
				Name:        "git_commit",
				Description: "提交更改",
				ServerID:    "git",
			},
			{
				Name:        "git_push",
				Description: "推送到远程仓库",
				ServerID:    "git",
			},
		},
	},
	{
		ID:          "database",
		Name:        "Database MCP Server",
		Description: "提供数据库操作工具，支持多种数据库连接和SQL查询执行",
		Version:     "0.9.0",
		Author:      "MCP Community",
		Homepage:    "https://github.com/mcp-community/database-server",
		Repository:  "https://github.com/mcp-community/database-server",
		License:     "Apache-2.0",
		Keywords:    []string{"database", "sql", "mysql", "postgresql", "sqlite"},
		Category:    "database",
		Transport:   "stdio",
		Command:     "npx",
		Args:        []string{"@mcp-community/database-server"},
		Env:         make(map[string]string),
		Status:      "available",
		Tools: []MCPTool{
			{
				Name:        "execute_query",
				Description: "执行SQL查询",
				ServerID:    "database",
			},
			{
				Name:        "list_tables",
				Description: "列出数据库表",
				ServerID:    "database",
			},
		},
	},
	{
		ID:          "web-search",
		Name:        "Web Search MCP Server",
		Description: "提供网络搜索功能，支持多种搜索引擎和搜索结果解析",
		Version:     "2.1.0",
		Author:      "SearchTools",
		Homepage:    "https://github.com/searchtools/web-search-server",
		Repository:  "https://github.com/searchtools/web-search-server",
		License:     "MIT",
		Keywords:    []string{"web", "search", "google", "bing", "internet"},
		Category:    "communication",
		Transport:   "sse",
		Command:     "",
		Args:        []string{},
		Env:         make(map[string]string),
		Status:      "available",
		Tools: []MCPTool{
			{
				Name:        "search_web",
				Description: "执行网络搜索",
				ServerID:    "web-search",
			},
		},
	},
	{
		ID:          "time",
		Name:        "Time MCP Server",
		Description: "提供时间相关工具，包括时区转换、时间格式化等功能",
		Version:     "1.0.1",
		Author:      "MCP Community",
		Homepage:    "https://github.com/mcp-community/time-server",
		Repository:  "https://github.com/mcp-community/time-server",
		License:     "MIT",
		Keywords:    []string{"time", "date", "timezone", "format"},
		Category:    "productivity",
		Transport:   "stdio",
		Command:     "npx",
		Args:        []string{"@mcp-community/time-server"},
		Env:         make(map[string]string),
		Status:      "available",
		Tools: []MCPTool{
			{
				Name:        "get_current_time",
				Description: "获取当前时间",
				ServerID:    "time",
			},
			{
				Name:        "convert_timezone",
				Description: "时区转换",
				ServerID:    "time",
			},
		},
	},
}
