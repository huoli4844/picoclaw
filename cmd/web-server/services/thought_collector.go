package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/cmd/web-server/models"
	"github.com/sipeed/picoclaw/pkg/logger"
)

// ThoughtCollector 收集思考过程
type ThoughtCollector struct {
	thoughts   []models.Thought
	mu         sync.Mutex
	callback   func(models.Thought)
	sessionKey string // 用于过滤特定会话的日志
}

// NewThoughtCollector 创建思考收集器
func NewThoughtCollector(callback func(models.Thought)) *ThoughtCollector {
	return &ThoughtCollector{
		thoughts: make([]models.Thought, 0),
		callback: callback,
	}
}

// NewThoughtCollectorWithSession 创建带会话密钥的思考收集器
func NewThoughtCollectorWithSession(callback func(models.Thought), sessionKey string) *ThoughtCollector {
	return &ThoughtCollector{
		thoughts:   make([]models.Thought, 0),
		callback:   callback,
		sessionKey: sessionKey,
	}
}

// AddThought 添加思考
func (tc *ThoughtCollector) AddThought(thoughtType, content string) {
	tc.AddThoughtWithDetails(thoughtType, content, "", "", "", 0, 0)
}

// AddThoughtWithDetails 添加带详细信息的思考
func (tc *ThoughtCollector) AddThoughtWithDetails(
	thoughtType, content, toolName, args, result string,
	duration, iteration int,
) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	thought := models.Thought{
		Type:      thoughtType,
		Timestamp: time.Now(),
		Content:   content,
		ToolName:  toolName,
		Args:      args,
		Result:    result,
		Duration:  duration,
		Iteration: iteration,
	}

	tc.thoughts = append(tc.thoughts, thought)

	if tc.callback != nil {
		tc.callback(thought)
	}
}

// GetThoughts 获取思考过程
func (tc *ThoughtCollector) GetThoughts() []models.Thought {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	thoughtsCopy := make([]models.Thought, len(tc.thoughts))
	copy(tc.thoughts, thoughtsCopy)
	return thoughtsCopy
}

// Reset 重置思考过程
func (tc *ThoughtCollector) Reset() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.thoughts = make([]models.Thought, 0)
}

// LogEntryAdapter 适配日志系统以收集详细日志
type LogEntryAdapter struct {
	collector *ThoughtCollector
}

// NewLogEntryAdapter 创建日志条目适配器
func NewLogEntryAdapter(collector *ThoughtCollector) *LogEntryAdapter {
	return &LogEntryAdapter{
		collector: collector,
	}
}

// OnLogEntry 实现LogListener接口，处理日志条目，转换为Thought格式
func (la *LogEntryAdapter) OnLogEntry(logLevel logger.LogLevel, component string, message string, fields map[string]any) {
	// 只处理INFO级别及以上，且组件为agent或tool的日志
	if component != "agent" && component != "tool" {
		return
	}

	// 会话过滤 - 如果有sessionKey字段，确保匹配当前会话
	if la.collector.sessionKey != "" && fields != nil {
		if sessionKey, ok := fields["session_key"].(string); ok {
			// 如果session key不匹配，则忽略此日志
			if sessionKey != la.collector.sessionKey && !strings.Contains(sessionKey, strings.TrimPrefix(la.collector.sessionKey, "web:")) {
				return
			}
		}
	}

	// 根据消息内容和组件类型生成不同类型的Thought
	switch {
	case component == "agent" && strings.Contains(message, "LLM requested tool calls"):
		// 解析工具调用列表 - 只显示实际存在的工具
		var toolsInfo string
		if fields != nil {
			if tools, ok := fields["tools"].([]string); ok {
				// 过滤掉不存在的工具（如list_skills）
				var validTools []string
				for _, tool := range tools {
					if tool != "list_skills" { // list_skills不是真实工具
						validTools = append(validTools, tool)
					}
				}
				if len(validTools) > 0 {
					toolsInfo = fmt.Sprintf("请求工具: [%s]", strings.Join(validTools, ", "))
					toolsInfo += fmt.Sprintf(" (共%d个)", len(validTools))
					la.collector.AddThought("tool_request", fmt.Sprintf("🤖 AI请求工具调用: %s", toolsInfo))
				}
			}
		}

	case component == "agent" && strings.Contains(message, "Tool call:"):
		// 解析工具调用信息，从消息中提取工具名和参数
		toolName := "unknown"
		argsStr := ""

		// 尝试从fields获取信息
		if fields != nil {
			if name, ok := fields["tool"].(string); ok {
				toolName = name
			}
		}

		// 过滤掉不存在的工具
		if toolName == "list_skills" {
			return // 跳过不存在的工具
		}

		// 尝试从消息中解析参数
		if strings.Contains(message, "(") && strings.Contains(message, ")") {
			start := strings.Index(message, "(") + 1
			end := strings.LastIndex(message, ")")
			if start > 0 && end > start {
				argsStr = message[start:end]
				// 尝试格式化JSON参数
				if json.Valid([]byte(argsStr)) {
					var prettyJSON bytes.Buffer
					if json.Indent(&prettyJSON, []byte(argsStr), "", "  ") == nil {
						argsStr = prettyJSON.String()
					}
				}
			}
		}

		if argsStr != "" {
			la.collector.AddThought("tool_call", fmt.Sprintf("🔧 调用工具: %s\n📝 参数:\n```json\n%s\n```", toolName, argsStr))
		} else {
			la.collector.AddThought("tool_call", fmt.Sprintf("🔧 调用工具: %s", toolName))
		}

	case component == "tool" && strings.Contains(message, "Tool execution started"):
		// 工具执行开始，获取详细参数
		var toolName string
		var argsInfo string

		if fields != nil {
			if name, ok := fields["tool"].(string); ok {
				toolName = name
			}
			if args, ok := fields["args"].(map[string]interface{}); ok {
				argsJSON, _ := json.Marshal(args)
				argsStr := string(argsJSON)
				// 美化JSON格式
				var prettyJSON bytes.Buffer
				if json.Indent(&prettyJSON, []byte(argsStr), "", "  ") == nil {
					argsInfo = prettyJSON.String()
				} else {
					argsInfo = argsStr
				}
				// 限制长度
				if len(argsInfo) > 300 {
					argsInfo = argsInfo[:297] + "..."
				}
			}
		}

		if argsInfo != "" {
			la.collector.AddThought("tool_start", fmt.Sprintf("🚀 开始执行工具: %s\n📋 完整参数:\n```json\n%s\n```", toolName, argsInfo))
		} else {
			la.collector.AddThought("tool_start", fmt.Sprintf("🚀 开始执行工具: %s", toolName))
		}

	case component == "tool" && strings.Contains(message, "Tool execution completed"):
		// 工具执行完成，显示执行时间和详细结果
		var toolName string
		var duration int
		var resultLength int

		if fields != nil {
			if name, ok := fields["tool"].(string); ok {
				toolName = name
			}
			if dur, ok := fields["duration_ms"].(int); ok {
				duration = dur
			}
			if resLen, ok := fields["result_length"].(int); ok {
				resultLength = resLen
			}
		}

		resultInfo := fmt.Sprintf("✅ 工具执行完成: %s", toolName)
		if duration > 0 {
			resultInfo += fmt.Sprintf(" (耗时: %dms)", duration)
		}
		if resultLength > 0 {
			resultInfo += fmt.Sprintf("\n📊 返回结果长度: %d字符", resultLength)
		}

		// 为特定工具添加详细信息
		la.addToolSpecificInfo(toolName, fields, resultInfo)

		la.collector.AddThought("tool_complete", resultInfo)

	case component == "tool" && strings.Contains(message, "Tool execution failed"):
		// 工具执行失败
		var toolName string
		var errorMsg string

		if fields != nil {
			if name, ok := fields["tool"].(string); ok {
				toolName = name
			}
			if err, ok := fields["error"].(string); ok {
				errorMsg = err
			}
		}

		la.collector.AddThought("tool_error", fmt.Sprintf("❌ 工具执行失败: %s\n🔍 错误信息: %s", toolName, errorMsg))

	default:
		// 其他agent相关的日志
		if component == "agent" {
			la.collector.AddThought("agent_log", message)
		}
	}
}

// addToolSpecificInfo 为特定工具添加详细信息
func (la *LogEntryAdapter) addToolSpecificInfo(toolName string, fields map[string]any, resultInfo string) string {
	switch toolName {
	case "read_file":
		if fields != nil {
			if args, ok := fields["args"].(map[string]interface{}); ok {
				if path, ok := args["path"].(string); ok {
					resultInfo += fmt.Sprintf("\n📄 读取文件: %s", path)
					// 显示文件路径类型
					if strings.Contains(path, "memory") {
						resultInfo += "\n🧠 类型: 记忆文件"
					} else if strings.Contains(path, "SKILL.md") {
						resultInfo += "\n🛠️ 类型: 技能说明文件"
					} else if strings.Contains(path, ".md") {
						resultInfo += "\n📝 类型: Markdown文档"
					}
				}
			}
		}
	case "write_file":
		if fields != nil {
			if args, ok := fields["args"].(map[string]interface{}); ok {
				if path, pathOk := args["path"].(string); pathOk {
					resultInfo += fmt.Sprintf("\n📝 写入文件: %s", path)
					if content, contentOk := args["content"].(string); contentOk {
						lines := strings.Count(content, "\n") + 1
						resultInfo += fmt.Sprintf("\n📏 内容长度: %d行, %d字符", lines, len(content))
						// 分析内容类型
						if strings.Contains(content, "#!/bin/bash") {
							resultInfo += "\n🐚 类型: Shell脚本"
						} else if strings.Contains(content, "```") {
							resultInfo += "\n💻 类型: 代码文件"
						} else if strings.Contains(content, "# ") {
							resultInfo += "\n📖 类型: Markdown文档"
						}
					}
				}
			}
		}
	case "append_file":
		if fields != nil {
			if args, ok := fields["args"].(map[string]interface{}); ok {
				if path, pathOk := args["path"].(string); pathOk {
					resultInfo += fmt.Sprintf("\n📎 追加内容到文件: %s", path)
					if content, contentOk := args["content"].(string); contentOk {
						resultInfo += fmt.Sprintf("\n📏 追加长度: %d字符", len(content))
					}
				}
			}
		}
	case "list_dir":
		if fields != nil {
			if args, ok := fields["args"].(map[string]interface{}); ok {
				if path, ok := args["path"].(string); ok {
					resultInfo += fmt.Sprintf("\n📁 列出目录: %s", path)
					if strings.Contains(path, "skills") {
						resultInfo += "\n🛠️ 类型: 技能目录"
					}
				}
			}
		}
	case "edit_file":
		if fields != nil {
			if args, ok := fields["args"].(map[string]interface{}); ok {
				if path, ok := args["path"].(string); ok {
					resultInfo += fmt.Sprintf("\n✏️ 编辑文件: %s", path)
					if oldText, ok := args["old_text"].(string); ok {
						resultInfo += fmt.Sprintf("\n📝 替换文本长度: %d字符", len(oldText))
					}
				}
			}
		}
	case "exec":
		if fields != nil {
			if args, ok := fields["args"].(map[string]interface{}); ok {
				if cmd, cmdOk := args["command"].(string); cmdOk {
					resultInfo += fmt.Sprintf("\n⚡ 执行命令: %s", cmd)
					// 分析命令类型
					if strings.HasPrefix(cmd, "ls") {
						resultInfo += "\n📋 类型: 文件列表命令"
					} else if strings.HasPrefix(cmd, "mkdir") {
						resultInfo += "\n📁 类型: 创建目录命令"
					} else if strings.HasPrefix(cmd, "git") {
						resultInfo += "\n🔧 类型: Git命令"
					} else if strings.HasPrefix(cmd, "cat") {
						resultInfo += "\n📄 类型: 文件查看命令"
					}
				}
			}
		}
	case "find_skills":
		if fields != nil {
			if args, ok := fields["args"].(map[string]interface{}); ok {
				if query, queryOk := args["query"].(string); queryOk {
					resultInfo += fmt.Sprintf("\n🔍 搜索技能: %s", query)
				}
				if limit, limitOk := args["limit"].(int); limitOk {
					resultInfo += fmt.Sprintf("\n📊 限制结果: %d个", limit)
				}
			}
		}
	case "install_skill":
		if fields != nil {
			if args, ok := fields["args"].(map[string]interface{}); ok {
				if slug, slugOk := args["slug"].(string); slugOk {
					resultInfo += fmt.Sprintf("\n📦 安装技能: %s", slug)
				}
				if registry, regOk := args["registry"].(string); regOk {
					resultInfo += fmt.Sprintf("\n📚 来源仓库: %s", registry)
				}
			}
		}
	}
	return resultInfo
}
