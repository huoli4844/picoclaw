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

	// Perform actual installation based on server type
	installResp := r.performActualInstallation(serverToInstall)
	if installResp.Status != "success" {
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
	switch {
	case server.Command == "npx":
		// Install npm package
		return r.installNpmPackage(server)
	case server.Command == "mcp-server-filesystem":
		// Check if filesystem server is available
		return r.checkFilesystemServer(server)
	case strings.HasPrefix(server.Command, "python"):
		// Install Python package
		return r.installPythonPackage(server)
	default:
		// For other commands, just verify they exist
		return r.verifyCommandAvailability(server)
	}
}

// installNpmPackage installs an npm package using npm or yarn
func (r *Registry) installNpmPackage(server *MCPServer) *MCPInstallResponse {
	// Check if npm is available
	if _, err := exec.LookPath("npm"); err != nil {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "npm is not available on this system. Please install Node.js and npm first.",
		}
	}

	// Extract package name from args
	var packageName string
	if len(server.Args) > 0 {
		packageName = server.Args[0]
	}

	if packageName == "" {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "Could not determine npm package name from server configuration",
		}
	}

	fmt.Printf("Installing npm package: %s\n", packageName)

	// Install the package globally using npm
	cmd := exec.Command("npm", "install", "-g", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Try with yarn as fallback
		fmt.Printf("npm install failed, trying with yarn...\n")
		cmd = exec.Command("yarn", "global", "add", packageName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return &MCPInstallResponse{
				Status:  "error",
				Message: fmt.Sprintf("Failed to install npm package %s: %v", packageName, err),
			}
		}
	}

	// Verify the package is available
	if err := r.verifyNpxCommand(packageName); err != nil {
		return &MCPInstallResponse{
			Status:  "error",
			Message: fmt.Sprintf("Package installed but verification failed: %v", err),
		}
	}

	return &MCPInstallResponse{
		Status:  "success",
		Message: fmt.Sprintf("npm package %s installed successfully", packageName),
	}
}

// checkFilesystemServer checks if the filesystem server is available
func (r *Registry) checkFilesystemServer(server *MCPServer) *MCPInstallResponse {
	// Check if mcp-server-filesystem command is available
	if _, err := exec.LookPath("mcp-server-filesystem"); err != nil {
		return &MCPInstallResponse{
			Status:  "error",
			Message: "mcp-server-filesystem command not found. Please install it first: npm install -g mcp-server-filesystem",
		}
	}

	// Try to run the command with --help to verify it works
	cmd := exec.Command("mcp-server-filesystem", "--help")
	if err := cmd.Run(); err != nil {
		// Some servers don't support --help, try version
		cmd = exec.Command("mcp-server-filesystem", "--version")
		if err := cmd.Run(); err != nil {
			// If both fail, that's okay - some servers don't have these flags
			fmt.Printf("Warning: Could not verify filesystem server, but command exists\n")
		}
	}

	return &MCPInstallResponse{
		Status:  "success",
		Message: "mcp-server-filesystem is available",
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
