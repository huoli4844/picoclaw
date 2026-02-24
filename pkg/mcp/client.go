package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// MCPClient represents a client for communicating with MCP servers
type MCPClient struct {
	serverID    string
	command     string
	args        []string
	env         map[string]string
	transport   string
	process     *exec.Cmd
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	conn        *WebSocketClient // For websocket transport
	mu          sync.RWMutex
	connected   bool
	requestID   int64
	pendingReqs map[int64]chan *JSONRPCResponse
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolCallResult represents the result of a tool call
type ToolCallResult struct {
	Content []ToolContent `json:"content,omitempty"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent represents content returned by a tool
type ToolContent struct {
	Type string      `json:"type"`
	Text string      `json:"text,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// WebSocketClient represents a websocket client (simplified for now)
type WebSocketClient struct {
	// TODO: Implement actual websocket client
	connected bool
}

// NewMCPClient creates a new MCP client for the given server
func NewMCPClient(server *MCPServer) (interface {
	Connect(ctx context.Context) error
	CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*ToolCallResult, error)
	Close() error
}, error) {
	// 优先使用 mcp-go 客户端
	if mcpClient, err := NewMCPGoClient(server); err == nil {
		fmt.Printf("使用 mcp-go 客户端\n")
		return mcpClient, nil
	}

	// 降级到自定义客户端
	fmt.Printf("降级使用自定义客户端\n")
	client := &MCPClient{
		serverID:    server.ID,
		command:     server.Command,
		args:        server.Args,
		env:         server.Env,
		transport:   server.Transport,
		pendingReqs: make(map[int64]chan *JSONRPCResponse),
	}
	return client, nil
}

// CreateWorkingMCPClient 创建一个工作正常的MCP客户端（当stdio有问题时使用）
func CreateWorkingMCPClient(server *MCPServer) (*WorkingMCPClient, error) {
	return NewWorkingMCPClient(server)
}

// Connect establishes a connection to the MCP server
func (c *MCPClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.transport {
	case "stdio":
		return c.connectStdioUnlocked(ctx)
	case "websocket":
		return c.connectWebSocket(ctx)
	case "sse":
		return c.connectSSE(ctx)
	default:
		return fmt.Errorf("unsupported transport: %s", c.transport)
	}
}

// connectStdio establishes a stdio connection to the MCP server
func (c *MCPClient) connectStdioUnlocked(ctx context.Context) error {
	fmt.Printf("连接MCP服务器: 命令=%s, 参数=%v\n", c.command, c.args)

	if c.command == "" {
		return fmt.Errorf("no command specified for stdio transport")
	}

	// Prepare the command
	c.process = exec.CommandContext(ctx, c.command, c.args...)

	// Set environment variables
	if len(c.env) > 0 {
		env := c.process.Env
		for k, v := range c.env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		c.process.Env = env
	}

	// Setup pipes before starting the process
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	c.stdin = stdinW
	c.process.Stdin = stdinR

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		stdinR.Close()
		stdinW.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	c.stdout = stdoutR
	c.process.Stdout = stdoutW

	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		stdinR.Close()
		stdinW.Close()
		stdoutR.Close()
		stdoutW.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	c.stderr = stderrR
	c.process.Stderr = stderrW

	fmt.Printf("启动MCP服务器进程...")
	// Start the process
	if err := c.process.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server process: %w", err)
	}
	fmt.Printf("MCP服务器进程已启动，PID=%d\n", c.process.Process.Pid)

	// Start stderr reader to capture any error messages
	go func() {
		scanner := bufio.NewScanner(c.stderr)
		for scanner.Scan() {
			fmt.Printf("MCP服务器STDERR: %s\n", scanner.Text())
		}
	}()

	// Start the response reader immediately to handle any server output
	go c.readResponses()

	// Wait a bit for the reader to be ready and server to start
	time.Sleep(200 * time.Millisecond)

	fmt.Printf("开始初始化MCP连接...")
	// Initialize the connection
	if err := c.initialize(ctx); err != nil {
		c.closeUnlocked() // 使用无锁关闭方法
		return fmt.Errorf("failed to initialize MCP connection: %w", err)
	}

	fmt.Printf("MCP连接初始化成功\n")
	c.connected = true
	return nil
}

// connectWebSocket establishes a websocket connection (placeholder)
func (c *MCPClient) connectWebSocket(ctx context.Context) error {
	// TODO: Implement websocket connection
	return fmt.Errorf("websocket transport not yet implemented")
}

// connectSSE establishes an SSE connection (placeholder)
func (c *MCPClient) connectSSE(ctx context.Context) error {
	// TODO: Implement SSE connection
	return fmt.Errorf("SSE transport not yet implemented")
}

// initialize sends the initialize message to the MCP server
func (c *MCPClient) initialize(ctx context.Context) error {
	fmt.Printf("发送初始化消息...")

	initParams := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"clientInfo": map[string]interface{}{
			"name":    "PicoClaw MCP Client",
			"version": "1.0.0",
		},
	}

	fmt.Printf("发送initialize请求...")
	response, err := c.sendRequestUnlocked("initialize", initParams)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("initialize error: %s", response.Error.Message)
	}
	fmt.Printf("initialize请求成功")

	fmt.Printf("发送initialized通知...")
	// Send initialized notification
	if err := c.sendNotificationUnlocked("notifications/initialized", nil); err != nil {
		return fmt.Errorf("initialized notification failed: %w", err)
	}
	fmt.Printf("initialized通知发送成功")

	return nil
}

// sendRequestUnlocked sends a JSON-RPC request without acquiring lock (for initialization)
func (c *MCPClient) sendRequestUnlocked(method string, params interface{}) (*JSONRPCResponse, error) {
	fmt.Printf("发送JSON-RPC请求(无锁): method=%s, params=%v\n", method, params)

	// Generate unique request ID
	c.requestID++
	id := c.requestID

	// Create response channel
	respChan := make(chan *JSONRPCResponse, 1)
	c.pendingReqs[id] = respChan

	// Create and send request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	fmt.Printf("写入消息到服务器...")
	if err := c.writeMessageUnlocked(request); err != nil {
		delete(c.pendingReqs, id)
		close(respChan)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	fmt.Printf("消息已发送，等待响应...")

	// Wait for response with timeout
	select {
	case response := <-respChan:
		delete(c.pendingReqs, id)
		close(respChan)
		fmt.Printf("收到响应: %+v\n", response)
		return response, nil
	case <-time.After(30 * time.Second): // 增加超时时间
		delete(c.pendingReqs, id)
		close(respChan)
		return nil, fmt.Errorf("request timeout after 30 seconds")
	}
}

// sendRequest sends a JSON-RPC request and waits for response
func (c *MCPClient) sendRequest(method string, params interface{}) (*JSONRPCResponse, error) {
	fmt.Printf("发送JSON-RPC请求: method=%s, params=%v\n", method, params)

	c.mu.Lock()
	defer c.mu.Unlock()

	// 允许在连接过程中发送请求（用于初始化）
	if !c.connected && method != "initialize" {
		return nil, fmt.Errorf("not connected to MCP server")
	}

	// Generate unique request ID
	c.requestID++
	id := c.requestID

	// Create response channel
	respChan := make(chan *JSONRPCResponse, 1)
	c.pendingReqs[id] = respChan

	// Create and send request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	fmt.Printf("写入消息到服务器...")
	if err := c.writeMessage(request); err != nil {
		delete(c.pendingReqs, id)
		close(respChan)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	fmt.Printf("消息已发送，等待响应...")

	// Wait for response with timeout
	select {
	case response := <-respChan:
		delete(c.pendingReqs, id)
		close(respChan)
		fmt.Printf("收到响应: %+v\n", response)
		return response, nil
	case <-time.After(30 * time.Second): // 增加超时时间
		delete(c.pendingReqs, id)
		close(respChan)
		return nil, fmt.Errorf("request timeout after 30 seconds")
	}
}

// sendNotificationUnlocked sends a JSON-RPC notification without acquiring lock
func (c *MCPClient) sendNotificationUnlocked(method string, params interface{}) error {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	return c.writeMessageUnlocked(request)
}

// sendNotification sends a JSON-RPC notification (no response expected)
func (c *MCPClient) sendNotification(method string, params interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 允许在连接过程中发送通知（用于initialized通知）
	if !c.connected && method != "notifications/initialized" {
		return fmt.Errorf("not connected to MCP server")
	}

	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	return c.writeMessage(request)
}

// writeMessageUnlocked writes a JSON-RPC message without acquiring lock
func (c *MCPClient) writeMessageUnlocked(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	fmt.Printf("消息数据(无锁): %s\n", string(data))

	if c.transport == "stdio" {
		// Create the complete message with headers
		fullMessage := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)

		// Send the complete message in one write
		_, err = c.stdin.Write([]byte(fullMessage))
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}

		// Flush the stdin to ensure message is sent immediately
		if flusher, ok := c.stdin.(interface{ Flush() error }); ok {
			if err := flusher.Flush(); err != nil {
				fmt.Printf("Warning: failed to flush stdin: %v\n", err)
			} else {
				fmt.Printf("stdin已flush\n")
			}
		} else {
			fmt.Printf("stdin不支持flush\n")
		}

		fmt.Printf("消息已写入stdio，长度: %d\n", len(fullMessage))
		return nil
	}

	return fmt.Errorf("write not implemented for transport: %s", c.transport)
}

// writeMessage writes a JSON-RPC message to the server
func (c *MCPClient) writeMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	fmt.Printf("消息数据: %s\n", string(data))

	if c.transport == "stdio" {
		// Create the complete message with headers
		fullMessage := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)

		// Send the complete message in one write
		_, err = c.stdin.Write([]byte(fullMessage))
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}

		fmt.Printf("消息已写入stdio，长度: %d\n", len(fullMessage))
		return nil
	}

	return fmt.Errorf("write not implemented for transport: %s", c.transport)
}

// readResponses reads responses from the MCP server
func (c *MCPClient) readResponses() {
	fmt.Printf("开始读取MCP服务器响应...")

	if c.transport != "stdio" {
		fmt.Printf("非stdio传输，跳过响应读取")
		return
	}

	// 使用简单的缓冲读取，类似成功的命令行测试
	buf := make([]byte, 4096)
	var allData []byte

	// 读取循环，有超时控制
	timeout := time.After(10 * time.Second)

	for {
		select {
		case <-timeout:
			fmt.Printf("读取超时，总共读取了 %d 字节\n", len(allData))
			if len(allData) > 0 {
				fmt.Printf("已读取的数据: %s\n", string(allData))
				c.parseCompleteResponse(allData)
			}
			goto done
		default:
			// 非阻塞读取
			n, err := c.stdout.Read(buf)
			if err != nil {
				if err != nil {
					if err.Error() == "EOF" {
						fmt.Printf("EOF，读取完成\n")
						goto done
					} else {
						fmt.Printf("读取错误: %v\n", err)
						goto done
					}
				}
			}

			if n > 0 {
				data := make([]byte, n)
				copy(data, buf[:n])
				allData = append(allData, data...)
				fmt.Printf("读取到 %d 字节: %s\n", n, string(data))

				// 尝试解析已读取的数据
				if c.tryParseMessage(allData) {
					fmt.Printf("成功解析了一条消息\n")
					goto done
				}
			}
		}
	}

done:
	fmt.Printf("响应读取循环结束\n")
}

// tryParseMessage 尝试从数据中解析完整的MCP消息
func (c *MCPClient) tryParseMessage(data []byte) bool {
	content := string(data)
	fmt.Printf("尝试解析消息: %s\n", content)

	// 查找Content-Length
	contentLengthRegex := `Content-Length:\s*(\d+)`
	re := regexp.MustCompile(contentLengthRegex)
	matches := re.FindStringSubmatch(content)

	if len(matches) < 2 {
		fmt.Printf("未找到Content-Length头部，等待更多数据\n")
		return false
	}

	var contentLength int
	if _, err := fmt.Sscanf(matches[1], "%d", &contentLength); err != nil {
		fmt.Printf("解析Content-Length失败: %v\n", err)
		return false
	}

	fmt.Printf("找到Content-Length: %d\n", contentLength)

	// 查找空行分隔符
	headerEnd := strings.Index(content, "\r\n\r\n")
	if headerEnd == -1 {
		headerEnd = strings.Index(content, "\n\n")
	}

	if headerEnd == -1 {
		fmt.Printf("未找到头部结束标记，等待更多数据\n")
		return false
	}

	// 计算消息开始位置
	messageStart := headerEnd
	if strings.Contains(content, "\r\n\r\n") {
		messageStart += 4
	} else {
		messageStart += 2
	}

	// 检查是否有完整的消息
	if messageStart+contentLength > len(content) {
		fmt.Printf("消息不完整，需要%d字节，有%d字节\n", contentLength, len(content)-messageStart)
		return false
	}

	// 提取消息内容
	messageData := content[messageStart : messageStart+contentLength]
	fmt.Printf("提取到完整消息: %s\n", messageData)

	// 处理响应
	c.handleResponse([]byte(messageData))

	return true
}

// parseCompleteResponse 解析完整的响应数据
func (c *MCPClient) parseCompleteResponse(data []byte) {
	content := string(data)
	fmt.Printf("解析完整响应: %s\n", content)

	// 查找Content-Length - 简单直接的字符串搜索
	lines := strings.Split(content, "\n")
	var contentLength int
	var messageStart int = -1

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Content-Length:") {
			// 直接提取数字
			lengthStr := strings.TrimPrefix(line, "Content-Length:")
			lengthStr = strings.TrimSpace(lengthStr)
			if _, err := fmt.Sscanf(lengthStr, "%d", &contentLength); err == nil {
				fmt.Printf("找到Content-Length: %d\n", contentLength)
			}
		} else if line == "" && contentLength > 0 && messageStart == -1 {
			// 找到空行，接下来是消息内容
			messageStart = i + 1
			fmt.Printf("找到消息起始位置: %d\n", messageStart)
			break
		}
	}

	if messageStart == -1 || contentLength == 0 {
		fmt.Printf("未找到完整的消息头格式\n")
		// 如果没有找到标准格式，尝试直接解析JSON
		c.tryDirectJSONParse(content)
		return
	}

	// 重建消息内容
	var messageText strings.Builder
	remainingLength := contentLength

	for i := messageStart; i < len(lines) && remainingLength > 0; i++ {
		line := lines[i]
		if i > messageStart {
			messageText.WriteString("\n")
			remainingLength--
			if remainingLength == 0 {
				break
			}
		}

		writeLen := len(line)
		if writeLen > remainingLength {
			writeLen = remainingLength
		}

		messageText.WriteString(line[:writeLen])
		remainingLength -= writeLen
	}

	if remainingLength == 0 {
		message := messageText.String()
		fmt.Printf("成功提取完整消息: %s\n", message)
		c.handleResponse([]byte(message))
	} else {
		fmt.Printf("消息不完整，还需要%d字节\n", remainingLength)
	}
}

// tryDirectJSONParse 尝试直接解析JSON
func (c *MCPClient) tryDirectJSONParse(content string) {
	fmt.Printf("尝试直接JSON解析\n")

	// 查找第一个{和最后一个}
	start := strings.Index(content, "{")
	if start == -1 {
		fmt.Printf("未找到JSON开始标记\n")
		return
	}

	end := strings.LastIndex(content, "}")
	if end == -1 {
		fmt.Printf("未找到JSON结束标记\n")
		return
	}

	if start < end {
		jsonStr := content[start : end+1]
		fmt.Printf("提取JSON: %s\n", jsonStr)
		c.handleResponse([]byte(jsonStr))
	}
}

// parseRawResponse parses raw response data that may contain MCP protocol messages
func (c *MCPClient) parseRawResponse(data []byte) {
	content := string(data)
	lines := strings.Split(content, "\n")

	var contentLength int
	var messageStart int

	// 查找Content-Length头部
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Content-Length:") {
			if _, err := fmt.Sscanf(line, "Content-Length: %d", &contentLength); err == nil {
				fmt.Printf("找到Content-Length: %d\n", contentLength)
			}
		} else if line == "" && contentLength > 0 {
			// 找到空行，接下来是消息内容
			messageStart = i + 1
			break
		}
	}

	// 提取消息内容
	if contentLength > 0 && messageStart < len(lines) {
		var messageData string
		for i := messageStart; i < len(lines) && len(messageData) < contentLength; i++ {
			if i > messageStart {
				messageData += "\n"
			}
			messageData += lines[i]
		}

		fmt.Printf("提取消息: %s\n", messageData)
		if len(messageData) > 0 {
			c.handleResponse([]byte(messageData))
		}
	}
}

// handleResponse handles an incoming JSON-RPC response
func (c *MCPClient) handleResponse(data []byte) {
	fmt.Printf("处理响应数据: %s\n", string(data))

	var response JSONRPCResponse
	if err := json.Unmarshal(data, &response); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		return
	}

	fmt.Printf("解析响应成功: ID=%v, Error=%v, Result=%v\n", response.ID, response.Error, response.Result)

	// Handle notifications (no ID)
	if response.ID == nil {
		// Handle notifications like logging messages
		fmt.Printf("收到通知消息: %+v\n", response)
		c.handleNotification(&response)
		return
	}

	// Handle request responses
	var id int64
	switch v := response.ID.(type) {
	case float64:
		id = int64(v)
	case int:
		id = int64(v)
	case int64:
		id = v
	case string:
		// Convert string ID if needed
		fmt.Sscanf(v, "%d", &id)
	default:
		fmt.Printf("未知的ID类型: %T, value: %v\n", v, v)
		return
	}

	fmt.Printf("查找响应通道，ID: %d\n", id)
	c.mu.RLock()
	respChan, exists := c.pendingReqs[id]
	c.mu.RUnlock()

	if exists {
		select {
		case respChan <- &response:
			fmt.Printf("响应已发送到通道\n")
		default:
			fmt.Printf("响应通道已满或已关闭\n")
		}
	} else {
		fmt.Printf("未找到对应的响应通道，ID: %d\n", id)
	}
}

// handleNotification handles incoming notifications
func (c *MCPClient) handleNotification(response *JSONRPCResponse) {
	// For now, just log notifications
	fmt.Printf("收到通知: %+v\n", response.Result)
}

// CallTool calls a tool on the MCP server
func (c *MCPClient) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*ToolCallResult, error) {
	params := map[string]interface{}{
		"name":      toolName,
		"arguments": arguments,
	}

	response, err := c.sendRequest("tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("tool call failed: %w", err)
	}

	if response.Error != nil {
		return &ToolCallResult{
			IsError: true,
			Content: []ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Tool call error: %s", response.Error.Message),
				},
			},
		}, nil
	}

	// Parse the result
	resultData, ok := response.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid result format")
	}

	result := &ToolCallResult{}

	// Check for error
	if isError, ok := resultData["isError"].(bool); ok {
		result.IsError = isError
	}

	// Parse content
	if contentData, ok := resultData["content"].([]interface{}); ok {
		for _, item := range contentData {
			if contentMap, ok := item.(map[string]interface{}); ok {
				content := ToolContent{}
				if contentType, ok := contentMap["type"].(string); ok {
					content.Type = contentType
				}
				if contentText, ok := contentMap["text"].(string); ok {
					content.Text = contentText
				}
				if contentData, ok := contentMap["data"]; ok {
					content.Data = contentData
				}
				result.Content = append(result.Content, content)
			}
		}
	}

	return result, nil
}

// closeUnlocked closes the connection without acquiring lock (assumes lock already held)
func (c *MCPClient) closeUnlocked() error {
	if !c.connected {
		return nil
	}

	c.connected = false

	// Close all pending request channels
	for _, ch := range c.pendingReqs {
		close(ch)
	}
	c.pendingReqs = make(map[int64]chan *JSONRPCResponse)

	// Close stdio connections
	var errors []error
	if c.stdin != nil {
		if err := c.stdin.Close(); err != nil {
			errors = append(errors, fmt.Errorf("stdin close error: %w", err))
		}
	}

	if c.process != nil && c.process.Process != nil {
		if err := c.process.Process.Kill(); err != nil {
			errors = append(errors, fmt.Errorf("process kill error: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple errors during close: %v", errors)
	}

	return nil
}

// Close closes the connection to the MCP server
func (c *MCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closeUnlocked()
}
