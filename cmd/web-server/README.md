# PicoClaw Web Server - 重构后的代码结构

## 概述

本项目已经完成了代码重构，将原来2000+行的单个`main.go`文件按照功能模块分离到不同的包中，提高了代码的可维护性和可扩展性。

## 目录结构

```
cmd/web-server/
├── main.go                    # 主入口文件，负责应用启动和路由配置
├── models/                    # 数据模型定义
│   └── models.go            # 所有数据结构定义
├── handlers/                  # HTTP处理器
│   ├── chat_handler.go      # 聊天相关API处理
│   ├── config_handler.go    # 配置相关API处理
│   ├── conversation_handler.go  # 对话历史API处理
│   ├── file_handler.go      # 文件浏览器API处理
│   ├── skill_handler.go     # 技能管理API处理
│   └── mcp_handler.go      # MCP服务器API处理
├── services/                  # 业务逻辑服务
│   ├── conversation_service.go   # 对话历史业务逻辑
│   └── thought_collector.go    # 思考过程收集服务
└── utils/                     # 工具函数
    └── sse.go             # Server-Sent Events工具函数
```

## 模块说明

### 1. Models (`models/`)
负责定义所有数据结构，包括：
- 聊天请求/响应模型
- 对话历史模型
- 配置模型
- 技能相关模型
- 文件操作模型
- MCP相关模型

### 2. Handlers (`handlers/`)
处理HTTP请求的各个模块：
- **ChatHandler**: 处理聊天API，支持流式和非流式响应
- **ConfigHandler**: 处理配置管理API
- **ConversationHandler**: 处理对话历史的CRUD操作
- **FileHandler**: 处理文件浏览、读取、删除等操作
- **SkillHandler**: 处理技能的安装、卸载、搜索等功能
- **MCPHandler**: 处理MCP服务器的管理

### 3. Services (`services/`)
包含业务逻辑：
- **ConversationService**: 对话历史的完整生命周期管理
- **ThoughtCollector**: AI思考过程的收集和处理

### 4. Utils (`utils/`)
通用工具函数：
- **SSE**: Server-Sent Events相关的工具函数

### 5. Main (`main.go`)
应用启动入口：
- 配置加载
- 服务初始化
- 路由设置
- 中间件配置
- 服务器启动

## 优势

1. **模块化**: 每个功能模块独立，职责单一
2. **可维护性**: 代码结构清晰，易于理解和修改
3. **可扩展性**: 新增功能只需添加对应的handler和service
4. **测试友好**: 各模块可以独立测试
5. **代码复用**: 业务逻辑与HTTP处理分离

## API端点

重构后支持所有原有的API端点：

### 配置相关
- `GET /api/config` - 获取配置
- `PUT /api/config` - 更新配置
- `GET /api/models` - 获取模型列表

### 聊天相关
- `POST /api/chat` - 聊天接口

### 对话历史
- `GET /api/conversations` - 获取对话列表
- `POST /api/conversations` - 创建新对话
- `GET /api/conversations/{id}` - 获取特定对话
- `PUT /api/conversations/{id}` - 更新对话
- `DELETE /api/conversations/{id}` - 删除对话

### 文件浏览器
- `GET /api/files` - 列出文件
- `GET /api/file-content` - 获取文件内容
- `DELETE /api/file-delete` - 删除文件/目录

### 技能管理
- `GET /api/skills` - 获取技能列表
- `GET /api/skills/{name}` - 获取技能详情
- `DELETE /api/skills/{name}` - 卸载技能
- `POST /api/skills/search` - 搜索技能
- `POST /api/skills/install` - 安装技能

### MCP管理
- `GET /api/mcp/servers` - 获取MCP服务器列表
- `GET /api/mcp/servers/{id}` - 获取服务器详情
- `POST /api/mcp/servers/{id}/validate` - 验证服务器
- `DELETE /api/mcp/servers/{id}` - 卸载服务器
- `POST /api/mcp/servers/{id}/call` - 调用工具
- `POST /api/mcp/search` - 搜索MCP服务器
- `GET /api/mcp/sources` - 获取MCP来源
- `POST /api/mcp/install` - 安装MCP服务器

## 构建和运行

```bash
# 构建
cd cmd/web-server
go build .

# 运行
./web-server
```

## 开发指南

### 添加新的API端点

1. 在 `models/models.go` 中定义相关的数据结构
2. 在 `services/` 中实现业务逻辑
3. 在 `handlers/` 中实现HTTP处理器
4. 在 `main.go` 的 `setupRoutes()` 函数中注册路由

### 修改现有功能

1. 找到对应的模块（handler/service/model）
2. 进行相应的修改
3. 确保API接口保持一致

这种模块化的架构使得代码更加清晰、可维护，并且便于团队协作开发。