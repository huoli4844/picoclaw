package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sipeed/picoclaw/cmd/web-server/models"
)

// FileHandler 文件处理器
type FileHandler struct{}

// NewFileHandler 创建文件处理器
func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

// ListFiles 列出文件
func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 获取查询参数中的路径，默认为用户主目录下的 .picoclaw
	path := r.URL.Query().Get("path")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			http.Error(w, "Failed to get user home directory", http.StatusInternalServerError)
			return
		}
		path = filepath.Join(home, ".picoclaw")
	}

	// 安全检查：确保路径在 .picoclaw 目录内
	home, _ := os.UserHomeDir()
	picoclawRoot := filepath.Join(home, ".picoclaw")
	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(absPath, picoclawRoot) {
		http.Error(w, "Access denied: path outside .picoclaw directory", http.StatusForbidden)
		return
	}

	// 读取目录内容
	files, err := os.ReadDir(absPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read directory: %v", err), http.StatusInternalServerError)
		return
	}

	// 构建文件信息列表
	var fileInfo []models.FileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		fileMap := models.FileInfo{
			Name:    file.Name(),
			Path:    filepath.Join(absPath, file.Name()),
			IsDir:   file.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		fileInfo = append(fileInfo, fileMap)
	}

	// 构建响应
	result := models.FileListResponse{
		Success: true,
		Path:    absPath,
		Files:   fileInfo,
	}

	json.NewEncoder(w).Encode(result)
}

// GetFileContent 获取文件内容
func (h *FileHandler) GetFileContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 获取查询参数中的文件路径
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter is required", http.StatusBadRequest)
		return
	}

	// 安全检查：确保路径在 .picoclaw 目录内
	home, _ := os.UserHomeDir()
	picoclawRoot := filepath.Join(home, ".picoclaw")
	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(absPath, picoclawRoot) {
		http.Error(w, "Access denied: path outside .picoclaw directory", http.StatusForbidden)
		return
	}

	// 检查是否为目录
	info, err := os.Stat(absPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file info: %v", err), http.StatusNotFound)
		return
	}

	if info.IsDir() {
		http.Error(w, "Path is a directory, not a file", http.StatusBadRequest)
		return
	}

	// 读取文件内容
	content, err := os.ReadFile(absPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		return
	}

	// 检查文件大小，避免读取过大的文件
	if len(content) > 10*1024*1024 { // 10MB 限制
		http.Error(w, "File too large (max 10MB)", http.StatusBadRequest)
		return
	}

	// 检查文件类型，判断是否为文本文件
	contentType := http.DetectContentType(content)
	if !h.isTextFile(contentType, path) {
		http.Error(w, "Binary files are not supported", http.StatusBadRequest)
		return
	}

	// 构建响应
	result := models.FileContentResponse{
		Success:     true,
		Path:        absPath,
		Content:     string(content),
		Size:        len(content),
		ContentType: contentType,
	}

	json.NewEncoder(w).Encode(result)
}

// DeleteFile 删除文件或目录
func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 从请求体中获取文件路径
	var request models.FileDeleteRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "Path parameter is required", http.StatusBadRequest)
		return
	}

	// 安全检查：确保路径在 .picoclaw 目录内
	home, _ := os.UserHomeDir()
	picoclawRoot := filepath.Join(home, ".picoclaw")
	absPath, err := filepath.Abs(request.Path)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(absPath, picoclawRoot) {
		http.Error(w, "Access denied: path outside .picoclaw directory", http.StatusForbidden)
		return
	}

	// 特殊保护：禁止删除某些重要的文件和目录
	protectedFiles := []string{
		filepath.Join(picoclawRoot, "config.json"),
		filepath.Join(picoclawRoot, "workspace"),
		filepath.Join(picoclawRoot, "skills"),
		filepath.Join(picoclawRoot, "AGENT.md"),
		filepath.Join(picoclawRoot, "HEARTBEAT.md"),
		filepath.Join(picoclawRoot, "IDENTITY.md"),
		filepath.Join(picoclawRoot, "SOUL.md"),
		filepath.Join(picoclawRoot, "USER.md"),
		filepath.Join(picoclawRoot, "MEMORY.md"),
	}

	for _, protected := range protectedFiles {
		if absPath == protected {
			http.Error(w, "Cannot delete protected file or directory", http.StatusForbidden)
			return
		}
	}

	// 获取文件/目录信息
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File or directory does not exist", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get file info: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// 删除文件或目录
	var deleteErr error
	if info.IsDir() {
		// 删除目录及其内容
		deleteErr = os.RemoveAll(absPath)
	} else {
		// 删除文件
		deleteErr = os.Remove(absPath)
	}

	if deleteErr != nil {
		http.Error(w, fmt.Sprintf("Failed to delete: %v", deleteErr), http.StatusInternalServerError)
		return
	}

	// 构建响应
	result := models.FileDeleteResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully deleted %s", filepath.Base(absPath)),
		Path:    absPath,
		IsDir:   info.IsDir(),
	}

	json.NewEncoder(w).Encode(result)
}

// isTextFile 检查是否为文本文件
func (h *FileHandler) isTextFile(contentType, path string) bool {
	if strings.HasPrefix(contentType, "text/") {
		return true
	}

	lowerPath := strings.ToLower(path)
	textExtensions := []string{
		".md", ".txt", ".json", ".log", ".yaml", ".yml", ".toml",
		".go", ".js", ".ts", ".tsx", ".py", ".sh", ".bat",
		".html", ".css", ".xml", ".csv",
	}

	for _, ext := range textExtensions {
		if strings.HasSuffix(lowerPath, ext) {
			return true
		}
	}

	return false
}
