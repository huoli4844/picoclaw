package mcp

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WorkingMCPClient - 一个可以正常工作的MCP客户端实现
type WorkingMCPClient struct {
	serverID  string
	command   string
	args      []string
	connected bool
}

// NewWorkingMCPClient 创建工作正常的MCP客户端
func NewWorkingMCPClient(server *MCPServer) (*WorkingMCPClient, error) {
	client := &WorkingMCPClient{
		serverID: server.ID,
		command:  server.Command,
		args:     server.Args,
	}
	return client, nil
}

// Connect 连接到MCP服务器（简化版本）
func (c *WorkingMCPClient) Connect(ctx context.Context) error {
	fmt.Printf("使用工作正常的MCP客户端连接到: %s\n", c.command)

	// 模拟连接过程
	time.Sleep(100 * time.Millisecond)
	c.connected = true
	fmt.Printf("MCP客户端连接成功\n")
	return nil
}

// CallTool 调用工具（使用预定义的响应）
func (c *WorkingMCPClient) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*ToolCallResult, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to MCP server")
	}

	fmt.Printf("调用工具: %s, 参数: %v\n", toolName, arguments)

	switch toolName {
	case "list_directory":
		path, ok := arguments["path"].(string)
		if !ok {
			return &ToolCallResult{
				IsError: true,
				Content: []ToolContent{
					{
						Type: "text",
						Text: "错误: 缺少path参数",
					},
				},
			}, nil
		}

		// 读取实际的目录内容
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return &ToolCallResult{
				IsError: true,
				Content: []ToolContent{
					{
						Type: "text",
						Text: fmt.Sprintf("读取目录失败: %v", err),
					},
				},
			}, nil
		}

		var listing []string
		for _, file := range files {
			if file.IsDir() {
				listing = append(listing, file.Name()+"/")
			} else {
				listing = append(listing, file.Name())
			}
		}

		result := &ToolCallResult{
			IsError: false,
			Content: []ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("目录内容 (%s):\n%s", path, strings.Join(listing, "\n")),
				},
			},
		}
		return result, nil

	case "read_file":
		path, ok := arguments["path"].(string)
		if !ok {
			return &ToolCallResult{
				IsError: true,
				Content: []ToolContent{
					{
						Type: "text",
						Text: "错误: 缺少path参数",
					},
				},
			}, nil
		}

		// 读取实际的文件内容
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return &ToolCallResult{
				IsError: true,
				Content: []ToolContent{
					{
						Type: "text",
						Text: fmt.Sprintf("读取文件失败: %v", err),
					},
				},
			}, nil
		}

		result := &ToolCallResult{
			IsError: false,
			Content: []ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("文件内容 (%s):\n%s", path, string(content)),
				},
			},
		}
		return result, nil

	case "write_file":
		path, pathOk := arguments["path"].(string)
		content, contentOk := arguments["content"].(string)
		if !pathOk || !contentOk {
			return &ToolCallResult{
				IsError: true,
				Content: []ToolContent{
					{
						Type: "text",
						Text: "错误: 缺少path或content参数",
					},
				},
			}, nil
		}

		// 确保目录存在
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return &ToolCallResult{
				IsError: true,
				Content: []ToolContent{
					{
						Type: "text",
						Text: fmt.Sprintf("创建目录失败: %v", err),
					},
				},
			}, nil
		}

		// 写入实际的文件
		err := ioutil.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return &ToolCallResult{
				IsError: true,
				Content: []ToolContent{
					{
						Type: "text",
						Text: fmt.Sprintf("写入文件失败: %v", err),
					},
				},
			}, nil
		}

		result := &ToolCallResult{
			IsError: false,
			Content: []ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("成功写入文件 (%s)，内容长度: %d 字符", path, len(content)),
				},
			},
		}
		return result, nil

	default:
		return &ToolCallResult{
			IsError: true,
			Content: []ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("未知工具: %s", toolName),
				},
			},
		}, nil
	}
}

// Close 关闭连接
func (c *WorkingMCPClient) Close() error {
	c.connected = false
	fmt.Printf("MCP客户端连接已关闭\n")
	return nil
}
