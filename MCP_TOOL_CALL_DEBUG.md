# MCP 工具调用调试指南

## 问题描述

当通过界面的"调试工具"按钮执行MCP功能时，没有返回结果。

## 已实施的修复

### 1. 改进了响应处理逻辑

**文件**: `pkg/mcp/client.go`

**修复内容**:
- 增加了详细的调试日志输出
- 修复了ID解析逻辑，支持更多类型（int, int64, float64, string）
- 添加了通知消息处理
- 增加了超时时间从10秒到30秒
- 改进了错误处理和状态报告

**关键改进**:
```go
// 改进的ID解析
switch v := response.ID.(type) {
case float64:
    id = int64(v)
case int: // 新增支持
    id = int64(v)
case int64:
    id = v
case string:
    fmt.Sscanf(v, "%d", &id)
default:
    fmt.Printf("未知的ID类型: %T, value: %v\n", v, v)
    return
}
```

### 2. 增强了前端调试功能

**文件**: `web/src/components/mcp/McpToolTester.tsx`

**修复内容**:
- 添加了详细的控制台日志输出
- 改进了错误信息显示
- 确保正确处理API响应数据

### 3. 改进了后端API处理

**文件**: `cmd/web-server/main.go`

**现有功能**:
- 详细的日志记录
- 完整的错误处理
- 正确的响应格式化

## 调试步骤

### 1. 启动服务器并检查日志

```bash
# 启动web服务器
./picoclaw-web-server

# 查看实时日志
# 日志会显示详细的MCP连接和调用过程
```

### 2. 检查MCP服务器安装状态

```bash
# 检查已安装的MCP服务器
curl -X GET http://localhost:8080/api/mcp/servers | jq '.'

# 验证特定服务器
curl -X POST http://localhost:8080/api/mcp/servers/filesystem/validate \
  -H "Content-Type: application/json" | jq '.'
```

### 3. 直接测试工具调用

使用提供的测试脚本：
```bash
./test_mcp_tool_call.sh
```

或手动测试：
```bash
# 测试read_file工具
curl -X POST http://localhost:8080/api/mcp/servers/filesystem/call \
  -H "Content-Type: application/json" \
  -d '{
    "toolName": "read_file",
    "arguments": {
      "path": "/Users/huoli4844/Documents/ai_project/picoclaw/README.md"
    }
  }' | jq '.'

# 测试list_directory工具
curl -X POST http://localhost:8080/api/mcp/servers/filesystem/call \
  -H "Content-Type: application/json" \
  -d '{
    "toolName": "list_directory",
    "arguments": {
      "path": "/Users/huoli4844/Documents/ai_project/picoclaw"
    }
  }' | jq '.'
```

### 4. 检查浏览器开发者工具

1. 打开浏览器开发者工具（F12）
2. 切换到"网络"标签
3. 在界面中执行MCP工具调用
4. 查看API请求的状态和响应
5. 检查控制台日志中的错误信息

## 常见问题和解决方案

### 1. MCP服务器未正确安装

**症状**: API返回"not found"或连接失败

**解决方案**:
```bash
# 重新安装filesystem服务器
curl -X POST http://localhost:8080/api/mcp/install \
  -H "Content-Type: application/json" \
  -d '{"serverId": "filesystem", "config": {}}'
```

### 2. 命令不可用

**症状**: 日志显示"command not found"

**解决方案**:
```bash
# 安装mcp-server-filesystem
npm install -g mcp-server-filesystem

# 或使用yarn
yarn global add mcp-server-filesystem

# 验证安装
mcp-server-filesystem --help
```

### 3. 权限问题

**症状**: 无法访问指定路径

**解决方案**:
- 检查文件路径是否正确
- 确认web服务器有读取权限
- 测试时使用绝对路径

### 4. 超时问题

**症状**: 请求在30秒后超时

**解决方案**:
- 检查MCP服务器进程是否正常运行
- 查看服务器日志中的错误信息
- 尝试重启web服务器

## 调试日志解读

### 正常的调用流程日志

```
发送JSON-RPC请求: method=tools/call, params=map[name:read_file arguments:map[path:...]]
写入消息到服务器...消息已发送，等待响应...
处理响应数据: {"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"..."}]}}
解析响应成功: ID=1, Error=<nil>, Result=map[content:[...]]
查找响应通道，ID: 1
响应已发送到通道
收到响应: &{JSONRPC:2.0 ID:1 Result:map[content:[...]] Error:<nil>}
工具调用成功，结果: &{Content:[{...}] IsError:false}
```

### 错误情况的日志

```
发送JSON-RPC请求: method=tools/call, params=...
写入消息到服务器...消息已发送，等待响应...
[30秒后]
request timeout after 30 seconds
工具调用失败: request timeout after 30 seconds
```

## 高级调试

### 1. 直接测试MCP服务器

```bash
# 直接启动filesystem服务器测试
mcp-server-filesystem /path/to/directory

# 在另一个终端发送JSON-RPC请求
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"read_file","arguments":{"path":"/test/file"}}}' | \
  { echo "Content-Length: $(wc -c)"; echo; cat; } | \
  mcp-server-filesystem /path/to/directory
```

### 2. 检查进程状态

```bash
# 查看MCP服务器进程
ps aux | grep mcp-server-filesystem

# 查看web服务器进程
ps aux | grep picoclaw-web-server

# 检查端口占用
lsof -i :8080
```

### 3. 网络调试

```bash
# 使用telnet测试API连接
telnet localhost 8080

# 使用wireshark或tcpdump监控HTTP请求
sudo tcpdump -i lo port 8080
```

## 性能优化建议

1. **减少调试日志**: 生产环境中可以减少`fmt.Printf`输出
2. **连接池**: 为频繁的MCP调用实现连接复用
3. **异步处理**: 对于长时间运行的工具调用，考虑异步处理
4. **缓存机制**: 对重复的工具调用结果进行缓存

## 联系支持

如果问题仍然存在，请提供以下信息：

1. 完整的错误日志
2. 浏览器开发者工具的网络请求详情
3. MCP服务器和web服务器的版本信息
4. 操作系统和环境信息
5. 重现问题的详细步骤

---

通过这些调试步骤和修复措施，MCP工具调用问题应该能够得到有效解决！