# NATS 功能测试说明

## 测试文件说明

### picoclaw_nats_test.go
PicoClaw NATS Channel 集成测试，主要测试：

1. **Channel生命周期测试** (`ChannelLifecycle`)
   - 测试 NATS Channel 的启动和停止
   - 验证连接状态和资源清理

2. **消息处理测试** (`MessageProcessing`)
   - 测试基本的消息收发功能
   - 验证消息路由和格式处理

3. **Watermill中间件测试** (`WatermillMiddleware`)
   - 测试 NATS Channel 的中间件功能
   - 验证多条消息的批量处理能力

4. **错误处理测试** (`ErrorHandling`)
   - 测试权限检查和错误处理
   - 验证未授权用户的消息过滤

### nats_simple.go
简化的 NATS Channel 实现，专注于核心功能：

- **直接使用原生 NATS API**
- **统一 JSON 消息格式**
- **简化消息处理流程**
- **保持权限检查和消息路由**

## 运行测试

### 前置条件

1. **确保 NATS 服务器正在运行**
   ```bash
   # 检查 NATS 服务是否运行
   curl http://171.221.201.55:28222/varz
   ```

2. **环境变量设置**（可选）
   ```bash
   export NATS_URL="nats://171.221.201.55:24222"
   ```

### 运行 PicoClaw NATS Channel 测试

```bash
# 运行所有 NATS Channel 测试
cd cmd/web-server/test
go test -v -run TestPicoClawNATSChannel

# 运行单个测试
go test -v -run TestPicoClawNATSChannel/ChannelLifecycle
go test -v -run TestPicoClawNATSChannel/MessageProcessing
go test -v -run TestPicoClawNATSChannel/WatermillMiddleware
go test -v -run TestPicoClawNATSChannel/ErrorHandling
```

### 快速测试脚本

```bash
# 使用提供的测试脚本
./run_nats_test.sh
```

## 核心功能说明

### 📡 NATS Channel 架构

PicoClaw 使用简化的 NATS Channel 实现：

```
┌─────────────────┐    JSON消息    ┌─────────────────┐
│   测试客户端     │ ──────────────► │  SimpleNATS     │
│                │                │  Channel        │
│                │ ◄────────────── │                │
│                │    响应消息    │                │
└─────────────────┘                └─────────────────┘
                                          │
                                          ▼
                                ┌─────────────────┐
                                │  PicoClaw       │
                                │  MessageBus     │
                                └─────────────────┘
```

### 🎯 支持的功能

#### 1. **单聊功能**
- **主题格式**：`picoclaw_test_topic`（配置的主题）
- **消息格式**：JSON 格式，包含 `sender_id`, `content`, `timestamp` 等字段
- **特点**：直接发布/订阅模式，支持实时消息传递

#### 2. **群聊功能**
- **主题格式**：与单聊相同，通过 `session_id` 或 `chat_id` 区分
- **广播机制**：所有订阅同一主题的客户端都会收到消息
- **用途**：支持群组消息广播

#### 3. **权限控制**
- **配置方式**：通过 `AllowFrom` 配置允许的用户ID列表
- **检查机制**：每条消息都会验证发送者权限
- **用户提取**：支持从消息内容或主题中提取用户ID

#### 4. **消息路由**
- **入口**：NATS Channel 接收外部消息
- **处理**：权限检查 → 格式解析 → 转发到 PicoClaw MessageBus
- **出口**：通过 `Send()` 方法发送响应消息

### 📝 消息格式

#### 接收消息格式
```json
{
  "type": "request",
  "message_id": "middleware_test_0_1234567890",
  "sender_id": "middleware_user", 
  "content": "测试消息内容",
  "timestamp": 1234567890123
}
```

#### 响应消息格式
```json
{
  "type": "response",
  "message_id": "msg_1234567890",
  "content": "AI响应内容",
  "model": "deepseek",
  "session_id": "chat_id_123",
  "timestamp": 1234567890123
}
```

## 测试输出示例

### 成功的测试输出
```
=== RUN   TestPicoClawNATSChannel
=== RUN   TestPicoClawNATSChannel/WatermillMiddleware
    picoclaw_nats_test.go:195: ⚙️ 测试Watermill中间件功能
2026/02/26 14:30:57 [2026-02-26T06:30:57Z] [INFO] nats: Starting simple NATS channel
2026/02/26 14:30:57 [2026-02-26T06:30:57Z] [INFO] nats: Subscribed to topic {topic=picoclaw_test_middleware_simple}
2026/02/26 14:30:57 [2026-02-26T06:30:57Z] [INFO] nats: Simple NATS channel started successfully
    picoclaw_nats_test.go:274: 📤 多条测试消息已发送
2026/02/26 14:30:57 [2026-02-26T06:30:57Z] [INFO] nats: Received message {subject=picoclaw_test_middleware_simple, payload_len=163}
2026/02/26 14:30:57 [2026-02-26T06:30:57Z] [INFO] nats: Processing message {user_id=middleware_user, message_id=middleware_test_0_123, content_len=23}
2026/02/26 14:30:57 [2026-02-26T06:30:57Z] [INFO] nats: Processing message {user_id=middleware_user, message_id=middleware_test_1_456, content_len=23}
... (处理更多消息)
    picoclaw_nats_test.go:285: ✅ 中间件成功处理了 5 条消息
--- PASS: TestPicoClawNATSChannel/WatermillMiddleware (2.05s)
```

## 配置说明

### NATS 配置
```go
type NATSConfig struct {
    Enabled        bool                `json:"enabled"`
    URL            string              `json:"url"`             // NATS 服务器地址
    WebSocket      string              `json:"websocket"`        // WebSocket 地址（可选）
    Topic          string              `json:"topic"`           // 默认主题
    EnableJetStream bool               `json:"enable_jetstream"` // 是否启用 JetStream
    AllowFrom      FlexibleStringSlice `json:"allow_from"`      // 允许的用户列表
}
```

### 默认配置
```json
{
  "nats": {
    "enabled": true,
    "url": "nats://171.221.201.55:24222",
    "topic": "picoclaw_test_topic", 
    "enable_jetstream": false,
    "allow_from": ["user_888", "test123"]
  }
}
```

## 故障排查

### 常见问题

#### 1. 连接失败
```
❌ 创建NATS Channel失败: failed to connect to NATS: dial tcp 171.221.201.55:24222: connect: connection refused
```
**解决方法**：
- 检查 NATS 服务是否运行
- 验证服务器地址和端口
- 检查防火墙设置

#### 2. 消息未被处理
```
❌ 消息处理数量不匹配: 期望=5, 实际=0
```
**解决方法**：
- 检查 `AllowFrom` 配置是否包含测试用户
- 验证消息格式是否正确
- 确认主题名称匹配

#### 3. 权限被拒绝
```
WARN: Unauthorized user attempted to send message
```
**解决方法**：
- 添加用户ID到 `AllowFrom` 配置中
- 检查用户ID提取逻辑

### 调试技巧

1. **启用调试日志**
   ```go
   logger.SetLevel("debug")
   ```

2. **检查消息接收**
   ```bash
   # 查看 NATS 消息流
   nats sub picoclaw_test_topic
   ```

3. **验证配置**
   ```go
   fmt.Printf("Config: %+v\n", cfg)
   ```

## 性能优化

### 建议配置

1. **连接池设置**
   ```go
   natsgo.MaxReconnects(10)
   natsgo.ReconnectWait(2*time.Second)
   ```

2. **消息处理优化**
   - 使用批量处理减少系统调用
   - 异步处理提高吞吐量
   - 合理设置缓冲区大小

3. **JetStream 持久化**
   ```go
   EnableJetStream: true  // 启用消息持久化
   ```

## 下一步开发计划

1. **增强功能**
   - 添加 JetStream 持久化支持
   - 实现消息历史查询
   - 支持文件传输

2. **性能优化**
   - 消息批量处理
   - 连接池管理
   - 内存使用优化

3. **监控和调试**
   - 添加指标收集
   - 实现健康检查
   - 增强调试工具

## 注意事项

1. **测试环境**
   - 使用独立的测试环境
   - 避免与生产数据冲突
   - 及时清理测试资源

2. **安全性**
   - 配置适当的用户权限
   - 使用 TLS 加密连接
   - 定期更新依赖

3. **可靠性**
   - 实现重连机制
   - 添加错误处理
   - 监控系统状态