# PicoClaw Windows Web启动指南

## 🌐 Web启动方式对比

| 启动方式 | 脚本文件 | 端口 | 适合人群 | 特点 |
|---------|----------|------|----------|------|
| **内置Web服务** | `picoclaw gateway` | 8080 | 普通用户 | 集成度高，配置简单 |
| **完整Web服务** | `start-web.bat` | 8080 | 生产使用 | 构建前端+后端 |
| **开发模式** | `quick-web.bat` | 5173 | 开发者 | 热重载，调试方便 |
| **交互式菜单** | `run-picoclaw.bat` | 8080 | 新手用户 | 图形化选择 |

---

## 🚀 推荐启动方式

### 新手用户（最简单）
```cmd
# 双击运行
START.bat
```
- 一键启动
- 自动安装依赖
- 自动打开浏览器

### 日常使用
```cmd
# 方式1：使用内置Web服务
picoclaw gateway

# 方式2：使用完整脚本
start-web.bat
```

### 开发者使用
```cmd
# 开发模式（热重载）
quick-web.bat

# 或交互式菜单
run-picoclaw.bat
```

---

## 📋 各启动方式详解

### 1. START.bat - 一键启动
**适用场景**：第一次使用，最简单
```cmd
# 双击运行即可
START.bat
```

**特点**：
- ✅ 自动检查和安装
- ✅ 自动打开浏览器
- ✅ 无需配置

### 2. start-web.bat - 完整Web服务
**适用场景**：生产环境，完整功能
```cmd
start-web.bat
```

**特点**：
- ✅ 构建前端
- ✅ 启动后端
- ✅ 自动打开浏览器
- ✅ 错误检查

### 3. quick-web.bat - 开发模式
**适用场景**：开发调试
```cmd
quick-web.bat
```

**特点**：
- ✅ 热重载
- ✅ 开发服务器
- ✅ 端口5173
- ✅ 快速启动

### 4. picoclaw gateway - 内置服务
**适用场景**：简单使用
```cmd
picoclaw gateway
```

**特点**：
- ✅ 无需Node.js
- ✅ 集成度高
- ✅ 配置简单

### 5. run-picoclaw.bat - 交互式菜单
**适用场景**：不确定选择哪种方式
```cmd
run-picoclaw.bat
```

**特点**：
- ✅ 图形化选择
- ✅ 多种选项
- ✅ 状态显示

---

## 🔧 故障排除

### 问题1：端口被占用
```cmd
# 检查端口占用
netstat -ano | findstr :8080

# 结束占用进程
taskkill /PID [进程号] /F
```

### 问题2：Node.js依赖问题
```cmd
# 重新安装依赖
cd web
rmdir /s /q node_modules
npm install
```

### 问题3：Go构建问题
```cmd
# 重新构建Web服务器
build-web-server.bat

# 或使用修复脚本
fix-windows-build.bat
```

### 问题4：浏览器无法打开
- 手动访问：http://localhost:8080
- 检查防火墙设置
- 尝试其他浏览器

---

## 📁 相关文件说明

| 文件名 | 作用 | 适用场景 |
|--------|------|----------|
| `START.bat` | 一键启动Web界面 | 新手推荐 |
| `start-web.bat` | 完整Web启动脚本 | 生产使用 |
| `quick-web.bat` | 开发模式启动 | 开发者使用 |
| `run-picoclaw.bat` | 交互式菜单 | 功能选择 |
| `build-web-server.bat` | 构建后端服务器 | 手动构建 |

---

## 💡 使用建议

### 首次使用
1. 使用 `START.bat` 一键启动
2. 熟悉界面后，可尝试其他方式

### 日常使用
- **简单对话**：`START.bat`
- **管理配置**：`run-picoclaw.bat`
- **开发调试**：`quick-web.bat`

### 团队协作
- **开发环境**：`quick-web.bat`
- **演示环境**：`start-web.bat`
- **用户环境**：`START.bat`

---

## 🎯 快速选择指南

**我想要...**
- 🔰 **最简单的开始** → `START.bat`
- 🛠️ **开发调试** → `quick-web.bat`
- 🏭 **生产部署** → `start-web.bat`
- 🎛️ **功能齐全** → `run-picoclaw.bat`
- ⚡ **无需Node.js** → `picoclaw gateway`

---

现在您可以根据自己的需求选择最适合的Windows Web启动方式了！