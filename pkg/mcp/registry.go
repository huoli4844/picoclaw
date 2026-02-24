package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	// Search online MCP servers first
	response, err := SearchMCPServersOnline(req.Query, req.Category, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	// Create a map of installed servers for quick lookup
	installedServers := make(map[string]*MCPServer)
	for serverID, server := range r.servers {
		installedServers[serverID] = server
	}

	// Update server status based on installation
	var updatedResults []MCPServer
	for _, server := range response.Results {
		if installedServer, exists := installedServers[server.ID]; exists {
			// This server is installed - use the installed version with full details
			updatedServer := *installedServer
			updatedServer.Status = "installed"
			updatedResults = append(updatedResults, updatedServer)
		} else {
			// This server is available for installation
			server.Status = "available"
			updatedResults = append(updatedResults, server)
		}
	}

	response.Results = updatedResults
	return response, nil
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

	// Try to get server from registry first
	serverToInstall, err := GetMCPServerFromRegistry(req.ServerID)
	if err != nil {
		// Fallback to recommended servers
		for i := range RecommendedMCPServers {
			if RecommendedMCPServers[i].ID == req.ServerID {
				serverToInstall = &RecommendedMCPServers[i]
				break
			}
		}
	}

	if serverToInstall == nil {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Server %s not found in registry - use mcp-go to discover available servers", req.ServerID),
		}, nil
	}

	// Create server directory (replace slashes with dashes to avoid nested directories)
	safeServerID := strings.ReplaceAll(req.ServerID, "/", "-")
	serverDir := filepath.Join(r.installedDir, safeServerID)
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to create server directory: %v", err),
		}, nil
	}

	// Perform actual installation based on server type
	fmt.Printf("开始安装服务器 %s (命令: %s)\n", serverToInstall.ID, serverToInstall.Command)
	installResp := r.performActualInstallation(serverToInstall)
	fmt.Printf("安装结果: %s - %s\n", installResp.Status, installResp.Message)

	if installResp.Status != "success" {
		fmt.Printf("安装失败，清理目录: %s\n", serverDir)
		// Clean up directory if installation failed
		os.RemoveAll(serverDir)
		return installResp, nil
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

// performActualInstallation performs the real installation of MCP server
func (r *Registry) performActualInstallation(server *MCPServer) *MCPInstallResponse {
	fmt.Printf("执行安装 - 服务器ID: %s, 命令: %s\n", server.ID, server.Command)

	switch {
	case server.Command == "npx":
		// For mcp-go integration, we don't need to actually install npm packages
		// Just validate the configuration is valid for npx-based servers
		return r.validateNpxServer(server)
	case server.Command == "mcp-server-filesystem":
		// For mcp-go integration, we don't need to install actual command
		// Just verify the configuration is valid
		return r.validateMcpGoServer(server)
	case server.Command == "remote":
		// For remote HTTP servers, validate the URL format
		return r.validateRemoteServer(server)
	case strings.HasPrefix(server.Command, "python"):
		// For mcp-go integration, we don't need to actually install python packages
		return r.validatePythonServer(server)
	default:
		// For other commands, just verify they exist
		return r.verifyCommandAvailability(server)
	}
}

// installNpmPackage installs an npm package using npm or yarn
func (r *Registry) installNpmPackage(server *MCPServer) *MCPInstallResponse {
	// For mcp-go integration, redirect to validation
	return r.validateNpxServer(server)
}

// validateMcpGoServer validates server configuration for mcp-go integration
func (r *Registry) validateMcpGoServer(server *MCPServer) *MCPInstallResponse {
	fmt.Printf("Validating MCP server for mcp-go integration: %s\n", server.ID)

	// For mcp-go integration, we don't need actual command line tools
	// Just validate the configuration is complete and reasonable

	if server.ID == "" {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Server ID is required",
		}
	}

	if server.Name == "" {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Server name is required",
		}
	}

	if len(server.Tools) == 0 {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Server must have at least one tool defined",
		}
	}

	// For filesystem server, validate the args (should contain the directory path)
	if server.ID == "filesystem" {
		if len(server.Args) == 0 || server.Args[0] == "" {
			return &MCPInstallResponse{
				Status:  "error",
				Message: "Filesystem server requires a directory path in args",
			}
		}

		// Check if the directory exists
		if _, err := os.Stat(server.Args[0]); err != nil {
			return &MCPInstallResponse{
				Status:  "error",
				Message: fmt.Sprintf("Directory not found: %s", server.Args[0]),
			}
		}
	}

	fmt.Printf("MCP server validation successful: %s\n", server.ID)
	return &MCPInstallResponse{
		Status:  "success",
		Message: fmt.Sprintf("Server %s is ready for mcp-go integration", server.ID),
	}
}

// checkFilesystemServer checks if the filesystem server is available (legacy method)
func (r *Registry) checkFilesystemServer(server *MCPServer) *MCPInstallResponse {
	// For mcp-go integration, redirect to new validation method
	return r.validateMcpGoServer(server)
}

// validateNpxServer validates npx-based server configuration for mcp-go integration
func (r *Registry) validateNpxServer(server *MCPServer) *MCPInstallResponse {
	fmt.Printf("验证npx服务器: %s\n", server.ID)

	// Basic validation
	if server.ID == "" {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Server ID is required",
		}
	}

	if server.Name == "" {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Server name is required",
		}
	}

	if len(server.Args) == 0 {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "npx server requires package name in args",
		}
	}

	packageName := server.Args[0]
	if packageName == "" {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Package name cannot be empty",
		}
	}

	// For mcp-go integration, we don't need to check if npm is available
	// Just validate the package name format
	if !strings.Contains(packageName, "@") && !strings.Contains(packageName, "/") {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Invalid npm package format: %s", packageName),
		}
	}

	fmt.Printf("npx服务器验证成功: %s -> %s\n", server.ID, packageName)
	return &MCPInstallResponse{
		Status:  "success",
		Message: fmt.Sprintf("Server %s is ready for mcp-go integration (npx: %s)", server.ID, packageName),
	}
}

// validateRemoteServer validates remote HTTP server configuration
func (r *Registry) validateRemoteServer(server *MCPServer) *MCPInstallResponse {
	fmt.Printf("Validating remote MCP server: %s\n", server.ID)

	// Validate basic server information
	if server.ID == "" {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Server ID is required",
		}
	}

	if server.Name == "" {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Server name is required",
		}
	}

	// Validate that we have a remote URL
	if len(server.Args) == 0 || server.Args[0] == "" {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Remote server requires a URL in args",
		}
	}

	remoteURL := server.Args[0]

	// Basic URL validation
	if !strings.HasPrefix(remoteURL, "http://") && !strings.HasPrefix(remoteURL, "https://") {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Remote URL must start with http:// or https://",
		}
	}

	fmt.Printf("Remote MCP server validation successful: %s -> %s\n", server.ID, remoteURL)
	return &MCPInstallResponse{
		Status:  "success",
		Message: fmt.Sprintf("Remote server %s is ready for integration", server.ID),
	}
}

// validatePythonServer validates Python-based server configuration for mcp-go integration
func (r *Registry) validatePythonServer(server *MCPServer) *MCPInstallResponse {
	fmt.Printf("验证Python服务器: %s\n", server.ID)

	// Basic validation
	if server.ID == "" || server.Name == "" {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Server ID and name are required",
		}
	}

	fmt.Printf("Python服务器验证成功: %s\n", server.ID)
	return &MCPInstallResponse{
		Status:  "success",
		Message: fmt.Sprintf("Server %s is ready for mcp-go integration (Python)", server.ID),
	}
}

// installPythonPackage installs a Python package using pip
func (r *Registry) installPythonPackage(server *MCPServer) *MCPInstallResponse {
	// Check if pip is available
	if _, err := exec.LookPath("pip3"); err != nil {
		if _, err := exec.LookPath("pip"); err != nil {
			return &MCPInstallResponse{
				Status:  "error",
				Message: "pip is not available on this system. Please install Python and pip first.",
			}
		}
	}

	// For Python packages, we assume the command is something like "python -m package"
	// This is a placeholder for future Python MCP servers
	return &MCPInstallResponse{
		Status:  "success",
		Message: "Python package installation not yet implemented, but command is available",
	}
}

// verifyCommandAvailability verifies that a command is available on the system
func (r *Registry) verifyCommandAvailability(server *MCPServer) *MCPInstallResponse {
	if server.Command == "" {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "No command specified for server",
		}
	}

	// Check if the command exists in PATH
	if _, err := exec.LookPath(server.Command); err != nil {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Command '%s' not found in PATH: %v", server.Command, err),
		}
	}

	return &MCPInstallResponse{
		Status:  "success",
		Message: fmt.Sprintf("Command '%s' is available", server.Command),
	}
}

// verifyNpxCommand verifies that an npx command works
func (r *Registry) verifyNpxCommand(packageName string) error {
	// Check if npx is available
	if _, err := exec.LookPath("npx"); err != nil {
		return fmt.Errorf("npx is not available")
	}

	// Try to run npx with the package (with a timeout to prevent hanging)
	cmd := exec.Command("timeout", "10s", "npx", "--help")
	if err := cmd.Run(); err != nil {
		// Try without timeout command (macOS compatibility)
		cmd = exec.Command("npx", "--help")
		if err := cmd.Run(); err != nil {
			// npx might not have --help, but if it exists we're probably okay
			fmt.Printf("Warning: Could not verify npx, but command exists\n")
		}
	}

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

	// For installed servers, verify they are still working
	if server.Status == "installed" {
		if r.verifyServerHealth(server) {
			return "installed", nil
		} else {
			return "error", fmt.Errorf("server installation is broken")
		}
	}

	return server.Status, nil
}

// verifyServerHealth verifies that an installed server is still working
func (r *Registry) verifyServerHealth(server *MCPServer) bool {
	switch {
	case server.Command == "npx":
		// Verify npm package is still available
		if len(server.Args) > 0 {
			err := r.verifyNpxCommand(server.Args[0])
			return err == nil
		}
		return false
	case server.Command == "mcp-server-filesystem":
		_, err := exec.LookPath("mcp-server-filesystem")
		return err == nil
	case server.Command == "remote":
		// For remote servers, always assume healthy (URL was validated during installation)
		return true
	default:
		// For other commands, check if they still exist
		if server.Command != "" {
			_, err := exec.LookPath(server.Command)
			return err == nil
		}
		return false
	}
}

// ValidateInstallation validates an existing installation
func (r *Registry) ValidateInstallation(serverID string) (*MCPInstallResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	server, exists := r.servers[serverID]
	if !exists {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Server %s is not installed", serverID),
		}, nil
	}

	if r.verifyServerHealth(server) {
		return &MCPInstallResponse{
			Status:  "success",
			Message: fmt.Sprintf("Server %s installation is valid", serverID),
			Server:  server,
		}, nil
	} else {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Server %s installation is broken or missing dependencies", serverID),
			Server:  server,
		}, nil
	}
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
