package mcp

import (
	"fmt"
	"strings"
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

// SearchMCPServersOnline searches for MCP servers from online registries
func SearchMCPServersOnline(query string, category string, limit int, offset int) (*MCPSearchResponse, error) {
	// TODO: Replace with actual MCP registry API call
	// This could integrate with:
	// 1. Official MCP registry API (when available)
	// 2. GitHub search for MCP servers
	// 3. npm registry search for @modelcontextprotocol packages
	// 4. mcp-go discovery API

	// For now, provide a mock implementation with common MCP servers
	allServers := []MCPServer{
		{
			ID:          "filesystem",
			Name:        "Filesystem MCP Server",
			Description: "Filesystem operations and file management tools",
			Version:     "1.0.0",
			Author:      "Model Context Protocol",
			Homepage:    "https://modelcontextprotocol.io",
			Repository:  "https://github.com/modelcontextprotocol/servers",
			License:     "MIT",
			Keywords:    []string{"filesystem", "files", "directory", "io"},
			Category:    "filesystem",
			Transport:   "stdio",
			Command:     "npx",
			Args:        []string{"@modelcontextprotocol/server-filesystem"},
			Status:      "available",
		},
		{
			ID:          "time",
			Name:        "Time MCP Server",
			Description: "Time and timezone conversion tools",
			Version:     "1.0.1",
			Author:      "Model Context Protocol",
			Homepage:    "https://modelcontextprotocol.io",
			Repository:  "https://github.com/modelcontextprotocol/servers",
			License:     "MIT",
			Keywords:    []string{"time", "date", "timezone", "format"},
			Category:    "productivity",
			Transport:   "stdio",
			Command:     "npx",
			Args:        []string{"@modelcontextprotocol/server-time"},
			Status:      "available",
		},
		{
			ID:          "git",
			Name:        "Git MCP Server",
			Description: "Git version control and repository management",
			Version:     "1.2.0",
			Author:      "Model Context Protocol",
			Homepage:    "https://modelcontextprotocol.io",
			Repository:  "https://github.com/modelcontextprotocol/servers",
			License:     "MIT",
			Keywords:    []string{"git", "version control", "repository", "commit"},
			Category:    "development",
			Transport:   "stdio",
			Command:     "npx",
			Args:        []string{"@modelcontextprotocol/server-git"},
			Status:      "available",
		},
		{
			ID:          "database",
			Name:        "Database MCP Server",
			Description: "Database connections and SQL query execution",
			Version:     "0.9.0",
			Author:      "Model Context Protocol",
			Homepage:    "https://modelcontextprotocol.io",
			Repository:  "https://github.com/modelcontextprotocol/servers",
			License:     "Apache-2.0",
			Keywords:    []string{"database", "sql", "mysql", "postgresql", "sqlite"},
			Category:    "database",
			Transport:   "stdio",
			Command:     "npx",
			Args:        []string{"@modelcontextprotocol/server-postgres"},
			Status:      "available",
		},
		{
			ID:          "web-search",
			Name:        "Web Search MCP Server",
			Description: "Web search functionality with multiple search engines",
			Version:     "2.1.0",
			Author:      "Model Context Protocol",
			Homepage:    "https://modelcontextprotocol.io",
			Repository:  "https://github.com/modelcontextprotocol/servers",
			License:     "MIT",
			Keywords:    []string{"web", "search", "google", "bing", "internet"},
			Category:    "communication",
			Transport:   "stdio",
			Command:     "npx",
			Args:        []string{"@modelcontextprotocol/server-brave-search"},
			Status:      "available",
		},
		{
			ID:          "memory",
			Name:        "Memory MCP Server",
			Description: "Persistent memory storage and retrieval",
			Version:     "1.0.0",
			Author:      "Model Context Protocol",
			Homepage:    "https://modelcontextprotocol.io",
			Repository:  "https://github.com/modelcontextprotocol/servers",
			License:     "MIT",
			Keywords:    []string{"memory", "storage", "persistence", "cache"},
			Category:    "productivity",
			Transport:   "stdio",
			Command:     "npx",
			Args:        []string{"@modelcontextprotocol/server-memory"},
			Status:      "available",
		},
		{
			ID:          "puppeteer",
			Name:        "Puppeteer MCP Server",
			Description: "Web automation and browser control tools",
			Version:     "1.0.0",
			Author:      "Model Context Protocol",
			Homepage:    "https://modelcontextprotocol.io",
			Repository:  "https://github.com/modelcontextprotocol/servers",
			License:     "MIT",
			Keywords:    []string{"browser", "automation", "web", "scraping"},
			Category:    "development",
			Transport:   "stdio",
			Command:     "npx",
			Args:        []string{"@modelcontextprotocol/server-puppeteer"},
			Status:      "available",
		},
		{
			ID:          "slack",
			Name:        "Slack MCP Server",
			Description: "Slack integration and messaging tools",
			Version:     "1.0.0",
			Author:      "Model Context Protocol",
			Homepage:    "https://modelcontextprotocol.io",
			Repository:  "https://github.com/modelcontextprotocol/servers",
			License:     "MIT",
			Keywords:    []string{"slack", "messaging", "communication", "chat"},
			Category:    "communication",
			Transport:   "stdio",
			Command:     "npx",
			Args:        []string{"@modelcontextprotocol/server-slack"},
			Status:      "available",
		},
	}

	// Filter servers based on query and category
	var filteredServers []MCPServer
	for _, server := range allServers {
		matches := true

		// Filter by query text
		if query != "" {
			queryLower := strings.ToLower(query)
			matches = strings.Contains(strings.ToLower(server.Name), queryLower) ||
				strings.Contains(strings.ToLower(server.Description), queryLower) ||
				strings.Contains(strings.ToLower(server.ID), queryLower)
		}

		// Filter by category
		if matches && category != "" {
			matches = server.Category == category
		}

		if matches {
			// Check if already installed
			// TODO: Check against installed servers and mark status accordingly
			filteredServers = append(filteredServers, server)
		}
	}

	// Apply pagination
	total := len(filteredServers)
	if offset >= total {
		return &MCPSearchResponse{
			Query:   query,
			Results: []MCPServer{},
			Total:   total,
			Offset:  offset,
			Limit:   limit,
		}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return &MCPSearchResponse{
		Query:   query,
		Results: filteredServers[offset:end],
		Total:   total,
		Offset:  offset,
		Limit:   limit,
	}, nil
}

// GetMCPServerFromRegistry fetches MCP server configuration from external registry
func GetMCPServerFromRegistry(serverID string) (*MCPServer, error) {
	// Search the online registry for the specific server
	response, err := SearchMCPServersOnline("", "", 100, 0)
	if err != nil {
		return nil, err
	}

	for _, server := range response.Results {
		if server.ID == serverID {
			return &server, nil
		}
	}

	return nil, fmt.Errorf("server %s not found in online registry", serverID)
}

// RecommendedMCPServers provides a minimal fallback list of well-known servers
// This should only be used as fallback when external registry is unavailable
var RecommendedMCPServers = []MCPServer{
	{
		ID:          "filesystem",
		Name:        "Filesystem MCP Server",
		Description: "Filesystem operations and file management tools",
		Version:     "1.0.0",
		Author:      "Model Context Protocol",
		Homepage:    "https://modelcontextprotocol.io",
		Repository:  "https://github.com/modelcontextprotocol/servers",
		License:     "MIT",
		Keywords:    []string{"filesystem", "files", "directory", "io"},
		Category:    "filesystem",
		Transport:   "stdio",
		Command:     "npx",
		Args:        []string{"@modelcontextprotocol/server-filesystem"},
		Status:      "available",
	},
	{
		ID:          "time",
		Name:        "Time MCP Server",
		Description: "Time and timezone conversion tools",
		Version:     "1.0.1",
		Author:      "Model Context Protocol",
		Homepage:    "https://modelcontextprotocol.io",
		Repository:  "https://github.com/modelcontextprotocol/servers",
		License:     "MIT",
		Keywords:    []string{"time", "date", "timezone", "format"},
		Category:    "productivity",
		Transport:   "stdio",
		Command:     "npx",
		Args:        []string{"@modelcontextprotocol/server-time"},
		Status:      "available",
	},
}
