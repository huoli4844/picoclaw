# PicoClaw Web 界面安装指南

## 🚀 快速启动

### 1. 安装前端依赖

```bash
cd web
pnpm install
```

### 2. 启动前端开发服务器

```bash
pnpm dev
```

前端将在 http://localhost:3000 启动。

### 3. 启动后端服务器（在另一个终端）

```bash
# 首先确保你有 picoclaw 的配置
picoclaw onboard  # 如果还没有配置的话

# 启动 Web 服务器
go run cmd/web-server/main.go
```

后端将在 http://localhost:8080 启动。

### 4. 访问 Web 界面

打开浏览器访问：http://localhost:8080

## 📦 依赖安装问题解决

如果遇到依赖安装错误，请确保：

1. **使用 pnpm**（推荐）：
   ```bash
   npm install -g pnpm
   cd web
   pnpm install
   ```

2. **清理缓存**：
   ```bash
   rm -rf node_modules package-lock.json pnpm-lock.yaml yarn.lock
   pnpm install
   ```

3. **Node.js 版本**：确保使用 Node.js 18+ 版本

## ⚙️ 配置说明

### 环境变量

创建 `web/.env` 文件：
```env
VITE_API_URL=http://localhost:8080/api
```

### PicoClaw 配置

确保 `~/.picoclaw/config.json` 配置了模型：

```json
{
  "model_list": [
    {
      "model_name": "gpt-4",
      "model": "openai/gpt-4",
      "api_key": "your-openai-key"
    }
  ],
  "agents": {
    "defaults": {
      "model": "gpt-4"
    }
  }
}
```

## 🐛 故障排除

### 前端启动失败
```bash
# 检查 Node.js 版本
node --version  # 应该是 18+

# 重新安装依赖
rm -rf node_modules
pnpm install
```

### 后端启动失败
```bash
# 检查 Go 版本
go version

# 检查配置文件
ls -la ~/.picoclaw/config.json
```

### API 连接错误
- 检查后端是否在 8080 端口运行
- 检查 `VITE_API_URL` 环境变量配置
- 查看浏览器控制台的错误信息

## 🎨 界面功能

- ✨ 现代化聊天界面
- 🤖 多模型支持（GPT、Claude、GLM 等）
- ⚙️ 可视化模型配置
- 📱 响应式设计
- 🌙 暗色主题支持

## 📚 更多信息

- 查看 `web/README.md` 获取详细的开发文档
- 查看 `cmd/web-server/main.go` 了解后端 API 实现

🦞 享受 PicoClaw Web 体验！