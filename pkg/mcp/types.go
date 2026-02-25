package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// MCPSource represents an MCP source website
type MCPSource struct {
	Name     string `json:"name"`
	Homepage string `json:"homepage"`
}

// MCPSourceRegistry represents the MCP source registry structure
type MCPSourceRegistry struct {
	Version     string      `json:"version"`
	LastUpdated time.Time   `json:"last_updated"`
	RegistryURL string      `json:"registry_url"`
	Servers     []MCPSource `json:"servers"`
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
	Query     string   `json:"query"`
	Category  string   `json:"category,omitempty"`
	Transport string   `json:"transport,omitempty"`
	Sources   []string `json:"sources,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
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

// SearchMCPServersOnline searches for MCP servers from source websites
func SearchMCPServersOnline(query string, category string, sources []string, limit int, offset int) (*MCPSearchResponse, error) {
	// Use the new source-based search
	response, err := SearchMCPServersFromSources(query, sources, limit, offset)
	if err != nil {
		return nil, err
	}

	// If category filtering is requested, apply it
	if category != "" {
		var filteredResults []MCPServer
		for _, server := range response.Results {
			if server.Category == category {
				filteredResults = append(filteredResults, server)
			}
		}
		response.Results = filteredResults
	}

	return response, nil
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

// fetchMCPSources reads MCP sources from mcp.json config file
func fetchMCPSources() ([]MCPSource, error) {
	// Always use the absolute path to ensure correct file is read
	configPath := "/Users/huoli4844/Documents/ai_project/picoclaw/mcp/mcp.json"

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mcp.json from %s: %v", configPath, err)
	}

	var config struct {
		Servers     []MCPSource `json:"servers"`
		Version     string      `json:"version"`
		LastUpdated string      `json:"last_updated"`
		RegistryURL string      `json:"registry_url"`
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

// extractSourceFromServerID extracts the source prefix from a server ID
func extractSourceFromServerID(serverID string) string {
	parts := strings.Split(serverID, "/")
	if len(parts) >= 2 {
		return parts[0]
	}
	return "unknown"
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

// GetAvailableMCPSources returns a list of available MCP sources
func GetAvailableMCPSources() ([]string, error) {
	sources, err := fetchMCPSources()
	if err != nil {
		return nil, fmt.Errorf("failed to load MCP sources from config: %v", err)
	}

	var sourceNames []string
	for _, source := range sources {
		sourceNames = append(sourceNames, source.Name)
	}

	// Sort sources for consistent display
	sort.Strings(sourceNames)
	return sourceNames, nil
}

// SearchMCPServersFromSources searches for MCP servers from specified source websites
func SearchMCPServersFromSources(query string, sources []string, limit int, offset int) (*MCPSearchResponse, error) {
	var allServers []MCPServer

	// Get available sources from config
	availableSources, err := fetchMCPSources()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch MCP sources: %v", err)
	}

	// If no specific sources requested, search all sources
	if len(sources) == 0 {
		for _, source := range availableSources {
			sources = append(sources, source.Name)
		}
	}

	// Search each source
	for _, requestedSource := range sources {
		// Find the source configuration
		var sourceConfig *MCPSource
		for _, source := range availableSources {
			if source.Name == requestedSource {
				sourceConfig = &source
				break
			}
		}

		if sourceConfig == nil {
			continue // Skip unknown sources
		}

		// Fetch servers from this source
		serversFromSource, err := fetchServersFromSource(*sourceConfig, query)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch from source %s: %v\n", sourceConfig.Name, err)
			continue
		}

		allServers = append(allServers, serversFromSource...)
	}

	// Apply pagination
	total := len(allServers)
	start := offset
	end := start + limit
	if end > total {
		end = total
	}

	if start >= total {
		return &MCPSearchResponse{
			Query:   query,
			Results: []MCPServer{},
			Total:   total,
			Offset:  offset,
			Limit:   limit,
		}, nil
	}

	return &MCPSearchResponse{
		Query:   query,
		Results: allServers[start:end],
		Total:   total,
		Offset:  offset,
		Limit:   limit,
	}, nil
}

// fetchServersFromSource fetches MCP servers from a specific source website
func fetchServersFromSource(source MCPSource, query string) ([]MCPServer, error) {
	// For now, create a mock implementation that returns sample servers
	// In a real implementation, this would make HTTP requests to the source websites

	var servers []MCPServer

	// Mock data based on source name and query
	switch source.Name {
	case "pulsemcp":
		queryLower := strings.ToLower(query)
		if query == "" || strings.Contains(queryLower, "file") || strings.Contains(queryLower, "filesystem") {
			servers = append(servers, MCPServer{
				ID:          "pulsemcp/filesystem",
				Name:        "Filesystem Server",
				Description: "File system operations from PulseMCP",
				Version:     "1.0.0",
				Author:      "PulseMCP",
				Homepage:    source.Homepage,
				Repository:  "https://github.com/pulsemcp/filesystem",
				License:     "MIT",
				Keywords:    []string{"filesystem", "files", "directory", "io"},
				Category:    "filesystem",
				Transport:   "stdio",
				Status:      "available",
			})
		}
		if query == "" || strings.Contains(queryLower, "database") || strings.Contains(queryLower, "db") {
			servers = append(servers, MCPServer{
				ID:          "pulsemcp/database",
				Name:        "Database Server",
				Description: "Database operations from PulseMCP",
				Version:     "1.0.0",
				Author:      "PulseMCP",
				Homepage:    source.Homepage,
				Repository:  "https://github.com/pulsemcp/database",
				License:     "Apache-2.0",
				Keywords:    []string{"database", "db", "sql", "storage"},
				Category:    "database",
				Transport:   "stdio",
				Status:      "available",
			})
		}
	case "mcp.so":
		queryLower := strings.ToLower(query)
		if query == "" || strings.Contains(queryLower, "search") {
			servers = append(servers, MCPServer{
				ID:          "mcp.so/web-search",
				Name:        "Web Search",
				Description: "Web search functionality from MCP.so",
				Version:     "2.0.0",
				Author:      "MCP.so",
				Homepage:    source.Homepage,
				Repository:  "https://github.com/mcpso/web-search",
				License:     "MIT",
				Keywords:    []string{"search", "web", "internet", "query"},
				Category:    "utilities",
				Transport:   "sse",
				Status:      "available",
			})
		}
		if query == "" || strings.Contains(queryLower, "file") || strings.Contains(queryLower, "filesystem") {
			servers = append(servers, MCPServer{
				ID:          "mcp.so/file-manager",
				Name:        "File Manager",
				Description: "File management and operations from MCP.so",
				Version:     "1.5.0",
				Author:      "MCP.so",
				Homepage:    source.Homepage,
				Repository:  "https://github.com/mcpso/file-manager",
				License:     "MIT",
				Keywords:    []string{"file", "manager", "filesystem", "io"},
				Category:    "filesystem",
				Transport:   "stdio",
				Status:      "available",
			})
		}
	case "mcpservers":
		queryLower := strings.ToLower(query)
		if query == "" || strings.Contains(queryLower, "tools") || strings.Contains(queryLower, "dev") {
			servers = append(servers, MCPServer{
				ID:          "mcpservers/development-tools",
				Name:        "Development Tools",
				Description: "Development tools from MCPServers.org",
				Version:     "1.5.0",
				Author:      "MCPServers",
				Homepage:    source.Homepage,
				Repository:  "https://github.com/mcpservers/development-tools",
				License:     "BSD-3-Clause",
				Keywords:    []string{"development", "tools", "dev", "programming"},
				Category:    "development",
				Transport:   "websocket",
				Status:      "available",
			})
		}
		if query == "" || strings.Contains(queryLower, "file") || strings.Contains(queryLower, "filesystem") {
			servers = append(servers, MCPServer{
				ID:          "mcpservers/file-explorer",
				Name:        "File Explorer",
				Description: "File exploration and management from MCPServers.org",
				Version:     "2.0.0",
				Author:      "MCPServers",
				Homepage:    source.Homepage,
				Repository:  "https://github.com/mcpservers/file-explorer",
				License:     "GPL-3.0",
				Keywords:    []string{"file", "explorer", "filesystem", "browser"},
				Category:    "filesystem",
				Transport:   "stdio",
				Status:      "available",
			})
		}
	}

	return servers, nil
}

// GetMCPServerFromRegistry fetches MCP server configuration from external registry
func GetMCPServerFromRegistry(serverID string) (*MCPServer, error) {
	// Search the online registry for the specific server
	response, err := SearchMCPServersOnline("", "", nil, 100, 0)
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
