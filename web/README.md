# PicoClaw Web 界面

基于 React + TypeScript 的 PicoClaw Web 前端界面，提供现代化的聊天体验。

## 功能特性

- 🎨 现代化聊天界面设计
- 🤖 多模型支持（GPT、Claude、GLM 等）
- ⚙️ 可视化模型配置管理
- 📱 响应式设计，支持移动端
- 🌙 支持暗色主题
- 🔄 实时消息流

## 技术栈

- **前端框架**: React 18 + TypeScript
- **构建工具**: Vite
- **UI 组件**: Radix UI + Tailwind CSS
- **图标**: Lucide React
- **状态管理**: React Hooks

## 快速开始

### 1. 安装依赖

```bash
cd web
make install
# 或者
npm install
```

### 2. 启动开发服务器

```bash
make dev
# 或者
npm run dev
```

前端将在 http://localhost:3000 启动。

### 3. 构建生产版本

```bash
make build
# 或者
npm run build
```

构建文件将输出到 `dist/` 目录。

## 后端服务器

在另一个终端启动后端服务器：

```bash
go run cmd/web-server/main.go
```

后端将在 http://localhost:8080 启动，API 端点为 `/api`。

## 配置

### 环境变量

创建 `.env` 文件：

```env
VITE_API_URL=http://localhost:8080/api
```

### 模型配置

1. 点击界面右上角的"设置"按钮
2. 在弹出的对话框中添加模型配置
3. 支持的模型格式：
   - `openai/gpt-4` - OpenAI GPT-4
   - `anthropic/claude-3-5-sonnet` - Anthropic Claude
   - `zhipu/glm-4` - 智谱 GLM-4
   - `deepseek/deepseek-chat` - DeepSeek
   - `ollama/llama3` - 本地 Ollama 模型

## 项目结构

```
web/
├── src/
│   ├── components/          # React 组件
│   │   ├── ui/             # 基础 UI 组件
│   │   ├── settings/       # 设置相关组件
│   │   ├── ChatMessage.tsx
│   │   ├── ChatInput.tsx
│   │   └── ...
│   ├── hooks/              # React Hooks
│   ├── lib/                # 工具函数
│   ├── types/              # TypeScript 类型定义
│   ├── App.tsx             # 主应用组件
│   ├── main.tsx            # 应用入口
│   └── index.css           # 全局样式
├── public/                 # 静态资源
├── package.json
├── vite.config.ts
├── tailwind.config.js
└── README.md
```

## 开发指南

### 添加新组件

1. 在 `src/components/` 下创建组件文件
2. 使用 TypeScript 定义 Props 类型
3. 遵循现有的命名约定

### 样式规范

- 使用 Tailwind CSS 类名
- 组件级样式优先使用 CSS 变量
- 响应式设计优先考虑移动端

### API 调用

使用 `useApi` Hook 进行 API 调用：

```typescript
import { useApi } from '@/hooks/useApi'

const { sendChatMessage } = useApi()

const response = await sendChatMessage({
  message: "Hello!",
  model: selectedModel
})
```

## 部署

### 使用 Docker

构建包含前端和后端的镜像：

```bash
# 构建前端
cd web && npm run build

# 回到根目录
cd ..

# 构建并运行
docker build -t picoclaw-web .
docker run -p 8080:8080 picoclaw-web
```

### 手动部署

1. 构建前端：`cd web && npm run build`
2. 启动后端：`go run cmd/web-server/main.go`
3. 使用 Nginx 反向代理（可选）

## 故障排除

### 常见问题

1. **连接后端失败**
   - 检查后端服务器是否在 8080 端口运行
   - 检查 `VITE_API_URL` 环境变量配置

2. **模型配置错误**
   - 确认 API Key 正确
   - 检查模型格式是否正确
   - 验证网络连接

3. **构建失败**
   - 清理 node_modules：`rm -rf node_modules package-lock.json && npm install`
   - 检查 Node.js 版本（推荐 18+）

## 贡献

欢迎提交 Pull Request！请确保：

- 代码通过 TypeScript 检查
- 遵循现有的代码风格
- 添加必要的测试用例

## 许可证

MIT License