package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Registry manages MCP servers installation and registration
type Registry struct {
	configPath   string
	installedDir string
	servers      map[string]*MCPServer
	mu           sync.RWMutex
}

// NewRegistry creates a new MCP server registry
func NewRegistry(configPath, storageDir string) (*Registry, error) {
	r := &Registry{
		configPath:   configPath,
		installedDir: storageDir,
		servers:      make(map[string]*MCPServer),
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Load existing servers
	if err := r.loadInstalledServers(); err != nil {
		return nil, fmt.Errorf("failed to load installed servers: %w", err)
	}

	return r, nil
}

// loadInstalledServers loads installed servers from disk
func (r *Registry) loadInstalledServers() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, err := os.ReadDir(r.installedDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			serverPath := filepath.Join(r.installedDir, entry.Name(), "server.json")
			if data, err := os.ReadFile(serverPath); err == nil {
				var server MCPServer
				if err := json.Unmarshal(data, &server); err == nil {
					r.servers[server.ID] = &server
				}
			}
		}
	}

	return nil
}

// SearchServers searches for available MCP servers
func (r *Registry) SearchServers(req MCPSearchRequest) (*MCPSearchResponse, error) {
	var results []MCPServer

	// Search in known servers
	for _, server := range KnownMCPServers {
		if r.matchesQuery(server, req) {
			results = append(results, server)
		}
	}

	// Apply pagination
	total := len(results)
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	if offset >= total {
		return &MCPSearchResponse{
			Query:   req.Query,
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
		Query:   req.Query,
		Results: results[offset:end],
		Total:   total,
		Offset:  offset,
		Limit:   limit,
	}, nil
}

// matchesQuery checks if a server matches the search criteria
func (r *Registry) matchesQuery(server MCPServer, req MCPSearchRequest) bool {
	// Check query
	if req.Query != "" {
		query := req.Query
		nameMatch := containsIgnoreCase(server.Name, query)
		descMatch := containsIgnoreCase(server.Description, query)
		idMatch := containsIgnoreCase(server.ID, query)
		authorMatch := containsIgnoreCase(server.Author, query)

		keywordsMatch := false
		for _, keyword := range server.Keywords {
			if containsIgnoreCase(keyword, query) {
				keywordsMatch = true
				break
			}
		}

		if !nameMatch && !descMatch && !idMatch && !authorMatch && !keywordsMatch {
			return false
		}
	}

	// Check category
	if req.Category != "" && req.Category != "all" && server.Category != req.Category {
		return false
	}

	// Check transport
	if req.Transport != "" && req.Transport != "all" && server.Transport != req.Transport {
		return false
	}

	return true
}

// GetInstalledServers returns all installed MCP servers
func (r *Registry) GetInstalledServers() ([]MCPServer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var servers []MCPServer
	for _, server := range r.servers {
		servers = append(servers, *server)
	}

	return servers, nil
}

// GetServer returns a specific installed server by ID
func (r *Registry) GetServer(serverID string) (*MCPServer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	server, exists := r.servers[serverID]
	if !exists {
		return nil, fmt.Errorf("server %s not found", serverID)
	}

	return server, nil
}

// InstallServer installs an MCP server
func (r *Registry) InstallServer(req MCPInstallRequest) (*MCPInstallResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already installed
	if _, exists := r.servers[req.ServerID]; exists {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Server %s is already installed", req.ServerID),
		}, nil
	}

	// Find server in known servers
	var serverToInstall *MCPServer
	for i := range KnownMCPServers {
		if KnownMCPServers[i].ID == req.ServerID {
			serverToInstall = &KnownMCPServers[i]
			break
		}
	}

	if serverToInstall == nil {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Server %s not found in registry", req.ServerID),
		}, nil
	}

	// Create server directory
	serverDir := filepath.Join(r.installedDir, req.ServerID)
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to create server directory: %v", err),
		}, nil
	}

	// Apply configuration
	if req.Config != nil {
		serverToInstall.Config = req.Config
	}

	// Mark as installed
	serverToInstall.Status = "installed"
	now := time.Now()
	serverToInstall.InstalledAt = &now

	// Save server configuration
	serverFile := filepath.Join(serverDir, "server.json")
	serverData, err := json.MarshalIndent(serverToInstall, "", "  ")
	if err != nil {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to marshal server config: %v", err),
		}, nil
	}

	if err := os.WriteFile(serverFile, serverData, 0644); err != nil {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to save server config: %v", err),
		}, nil
	}

	// Add to registry
	r.servers[req.ServerID] = serverToInstall

	return &MCPInstallResponse{
		Status:  "success",
		Message: fmt.Sprintf("Server %s installed successfully", req.ServerID),
		Server:  serverToInstall,
	}, nil
}

// UninstallServer uninstalls an MCP server
func (r *Registry) UninstallServer(serverID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if server exists
	_, exists := r.servers[serverID]
	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	// Remove server directory
	serverDir := filepath.Join(r.installedDir, serverID)
	if err := os.RemoveAll(serverDir); err != nil {
		return fmt.Errorf("failed to remove server directory: %w", err)
	}

	// Remove from registry
	delete(r.servers, serverID)

	return nil
}

// GetServerStatus returns the status of a specific server
func (r *Registry) GetServerStatus(serverID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	server, exists := r.servers[serverID]
	if !exists {
		return "not_installed", nil
	}

	return server.Status, nil
}

// containsIgnoreCase checks if a string contains a substring ignoring case
func containsIgnoreCase(str, substr string) bool {
	str = str + ""
	substr = substr + ""

	// Simple case-insensitive contains check
	for i := 0; i <= len(str)-len(substr); i++ {
		if equalIgnoreCase(str[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

// equalIgnoreCase checks if two strings are equal ignoring case
func equalIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca = ca + ('a' - 'A')
		}
		if cb >= 'A' && cb <= 'Z' {
			cb = cb + ('a' - 'A')
		}
		if ca != cb {
			return false
		}
	}
	return true
}
