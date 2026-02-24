package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// MCPServerInfo represents the raw data from official MCP registry
type MCPServerInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Remote      string `json:"remote,omitempty"`
	Package     string `json:"package,omitempty"`
	Transport   string `json:"transport,omitempty"`
}

// SearchMCPServersOnline searches for MCP servers from mcp.json config file
func SearchMCPServersOnline(query string, category string, limit int, offset int) (*MCPSearchResponse, error) {
	// Fetch from mcp.json config file
	servers, err := fetchFromConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load MCP servers from config: %v", err)
	}

	// Filter servers based on query and category
	var filteredServers []MCPServer
	for _, server := range servers {
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
			// Ensure status is set to available for search results
			if server.Status == "" {
				server.Status = "available"
			}
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

// fetchFromOfficialRegistry fetches MCP servers from mcp.json config file
func fetchFromOfficialRegistry() ([]MCPServerInfo, error) {
	// Try to read from mcp.json in the project directory first
	configPath := "mcp/mcp.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Fallback to user's home directory
		configPath = filepath.Join(os.Getenv("HOME"), ".picoclaw", "mcp", "mcp.json")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mcp.json: %v", err)
	}

	var config struct {
		Servers []MCPServer `json:"servers"`
		Version string      `json:"version"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse mcp.json: %v", err)
	}

	// Convert MCPServer to MCPServerInfo
	var registryData []MCPServerInfo
	for _, server := range config.Servers {
		info := MCPServerInfo{
			ID:          server.ID,
			Name:        server.Name,
			Version:     server.Version,
			Description: server.Description,
			Package:     "",
			Remote:      "",
			Transport:   server.Transport,
		}

		// Extract package or remote from args
		if len(server.Args) > 0 {
			if server.Command == "npx" {
				info.Package = server.Args[0]
			} else if server.Command == "remote" {
				info.Remote = server.Args[0]
			} else if server.Command == "docker" && len(server.Args) >= 2 && server.Args[0] == "run" {
				info.Package = server.Args[1]
			}
		}

		registryData = append(registryData, info)
	}

	return registryData, nil
}

// fetchFromConfigFile reads MCP servers from mcp.json config file
func fetchFromConfigFile() ([]MCPServer, error) {
	// Try to read from mcp.json in the project directory first
	configPath := "mcp/mcp.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Fallback to user's home directory
		configPath = filepath.Join(os.Getenv("HOME"), ".picoclaw", "mcp", "mcp.json")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mcp.json: %v", err)
	}

	var config struct {
		Servers     []MCPServer `json:"servers"`
		Version     string      `json:"version"`
		LastUpdated string      `json:"last_updated"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse mcp.json: %v", err)
	}

	return config.Servers, nil
}

// createDefaultConfigFile creates a default config file with recommended servers
func createDefaultConfigFile(configPath string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	defaultServers := RecommendedMCPServers

	config := map[string]interface{}{
		"servers":      defaultServers,
		"version":      "1.0.0",
		"last_updated": time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// Helper functions for processing registry data
func generateServerID(name, id string) string {
	if id != "" {
		return strings.ReplaceAll(id, "/", "-")
	}
	return strings.ToLower(strings.ReplaceAll(strings.Join(strings.Fields(name), "-"), "_", "-"))
}

func extractKeywords(name, description string) []string {
	keywords := []string{}

	// Extract from name and description
	text := strings.ToLower(name + " " + description)

	// Common MCP keywords
	commonKeywords := []string{"api", "data", "management", "tools", "automation", "integration", "mcp", "server", "remote", "web", "cloud"}
	for _, keyword := range commonKeywords {
		if strings.Contains(text, keyword) {
			keywords = append(keywords, keyword)
		}
	}

	return keywords
}

func categorizeServer(name, description string) string {
	text := strings.ToLower(name + " " + description)

	// Define categories based on keywords
	if strings.Contains(text, "database") || strings.Contains(text, "sql") || strings.Contains(text, "storage") {
		return "database"
	}
	if strings.Contains(text, "development") || strings.Contains(text, "code") || strings.Contains(text, "api") || strings.Contains(text, "tools") {
		return "development"
	}
	if strings.Contains(text, "communication") || strings.Contains(text, "messaging") || strings.Contains(text, "chat") || strings.Contains(text, "slack") {
		return "communication"
	}
	if strings.Contains(text, "trading") || strings.Contains(text, "finance") || strings.Contains(text, "portfolio") {
		return "finance"
	}
	if strings.Contains(text, "docs") || strings.Contains(text, "document") || strings.Contains(text, "markdown") {
		return "productivity"
	}
	if strings.Contains(text, "search") || strings.Contains(text, "crawling") || strings.Contains(text, "web") {
		return "utilities"
	}

	return "other"
}

func determineTransport(pkg, remote string) string {
	if remote != "" {
		if strings.Contains(remote, "sse") {
			return "sse"
		}
		return "http"
	}
	if pkg != "" {
		if strings.Contains(pkg, "docker.io") {
			return "docker"
		}
		return "stdio"
	}
	return "stdio"
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
