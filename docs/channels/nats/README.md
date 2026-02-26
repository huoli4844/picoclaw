# NATS Channel

PicoClaw 通过 NATS 消息队列提供轻量级的点对点聊天机器人支持，特别适合企业内部IM工具集成。

## 配置

```json
{
  "channels": {
    "nats": {
      "enabled": true,
      "url": "nats://localhost:4222",
      "websocket": "wss://localhost:28444",
      "topic": "chatbot.user.888",
      "enable_jetstream": true,
      "allow_from": []
    }
  }
}
```

## 配置字段说明

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| enabled | bool | 是 | 是否启用 NATS Channel |
| url | string | 是 | NATS 服务器地址 |
| topic_pattern | string | 是 | 主题模式，`%s`会被替换为用户ID |
| user_ids | array | 否 | 指定用户ID列表（为空时使用模式订阅） |
| enable_jetstream | bool | 否 | 是否启用 JetStream 持久化（默认false） |
| allow_from | array | 否 | 用户ID白名单，空表示允许所有用户 |

## 工作原理

### 消息流程

```
用户IM → 自聊天机器人 → NATS主题 → PicoClaw → AI处理 → 响应 → NATS主题 → IM用户
```

### 主题设计

- **输入主题**: `chatbot.user.{userID}`
- **响应主题**: `chatbot.user.{userID}`（双向通信）

### 消息格式

#### 请求消息
```json
{
  "type": "request",
  "message_id": "unique-message-id",
  "sender_id": "user123",
  "content": "帮我分析这个代码",
  "model": "deepseek",
  "stream": true,
  "session_id": "optional-session-id",
  "metadata": {},
  "timestamp": 1704067200000
}
```

#### 响应消息
```json
{
  "type": "response",
  "message_id": "unique-message-id",
  "content": "代码分析结果...",
  "model": "deepseek",
  "timestamp": 1704067201000,
  "session_id": "optional-session-id",
  "is_complete": true,
  "thought": {
    "type": "thinking",
    "content": "正在分析代码结构..."
  }
}
```

## 使用场景

### 1. 企业内部IM集成
每个员工在IM中创建自己的AI助手：
- IM工具自动创建自聊天机器人
- 机器人订阅 `chatbot.user.{employeeID}` 主题
- 员工通过与自聊天机器人对话，获得专属AI助手

### 2. 多租户AI服务
为多个用户提供独立的AI聊天服务：
- 每个用户有独立的NATS主题
- 消息完全隔离，保证隐私
- 支持动态用户加入和退出

### 3. 微服务架构
作为微服务架构中的AI处理组件：
- 其他服务通过NATS调用AI能力
- 支持异步消息处理
- 可结合JetStream实现消息持久化

## 部署建议

### 1. 基础部署
```bash
# 启动NATS服务器
nats-server -p 4222

# 配置PicoClaw
vim config/config.json

# 启动PicoClaw
./picoclaw
```

### 2. 高可用部署
```bash
# 启动NATS集群
nats-server -cluster nats://localhost:6222 -p 4222
nats-server -cluster nats://localhost:6223 -p 4223
nats-server -cluster nats://localhost:6224 -p 4224

# 配置多个NATS URL
"url": "nats://server1:4222,nats://server2:4222,nats://server3:4222"
```

### 3. 持久化部署
```json
{
  "channels": {
    "nats": {
      "enabled": true,
      "url": "nats://localhost:4222",
      "enable_jetstream": true,
      "stream_config": {
        "name": "PICOCLAW_CHAT",
        "subjects": ["chatbot.user.*"],
        "retention": "workqueue",
        "max_age": "24h",
        "storage": "file"
      }
    }
  }
}
```

## 性能优化

### 1. 连接优化
- 使用连接池减少连接开销
- 启用重连机制确保服务可用性
- 配置合适的缓冲区大小

### 2. 消息优化
- 启用消息压缩减少网络传输
- 使用JetStream进行消息持久化
- 配置合适的流保留策略

### 3. 监控指标
- 监控消息延迟和吞吐量
- 跟踪连接状态和重连次数
- 统计处理成功率和错误率

## 故障排除

### 1. 连接问题
```bash
# 检查NATS服务器状态
nats server check

# 测试连接
telnet localhost 4222
```

### 2. 消息丢失
- 检查网络连接稳定性
- 启用JetStream持久化
- 配置适当的重试机制

### 3. 性能问题
- 监控NATS服务器资源使用
- 优化消息大小和频率
- 考虑分区和负载均衡

## 安全建议

1. **认证授权**: 配置NATS认证机制
2. **TLS加密**: 启用TLS传输加密
3. **网络隔离**: 使用私有网络部署
4. **访问控制**: 配置IP白名单和用户权限

## 示例代码

### 客户端发送消息（Go）
```go
nc, _ := nats.Connect("nats://localhost:4222")
defer nc.Close()

msg := NATSMessage{
    Type:      "request",
    MessageID: generateUUID(),
    SenderID:  "user123",
    Content:   "你好PicoClaw",
    Model:     "deepseek",
    Stream:    false,
    Timestamp: time.Now().UnixMilli(),
}

data, _ := json.Marshal(msg)
nc.Publish("chatbot.user.user123", data)
```

### 客户端接收响应（Go）
```go
nc.Subscribe("chatbot.user.user123", func(msg *nats.Msg) {
    var response NATSResponse
    json.Unmarshal(msg.Data, &response)
    
    if response.Type == "response" {
        fmt.Printf("AI回复: %s\n", response.Content)
    }
})
```

这种设计使得每个用户都能获得私密的AI助手体验，同时保持了系统的可扩展性和维护性。