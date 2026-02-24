# Windows环境下运行PicoClaw指南

## 🎯 快速开始 - 五种运行方式

### 方式1：一键Web启动（最简单）🔥
```cmd
# 下载项目后，双击运行
START.bat
```
**特点**：一键启动Web界面，自动安装依赖，自动打开浏览器

### 方式2：交互式完整菜单
```cmd
# 运行交互式菜单
run-picoclaw.bat
```
**特点**：图形化菜单，包含所有功能选择

### 方式3：一键安装运行
```cmd
# 下载项目后，直接运行
install-windows.bat
```
**特点**：自动完成构建、安装、配置所有步骤

### 方式4：专用Web启动脚本
```cmd
# 完整Web服务
start-web.bat

# 开发模式
quick-web.bat
```
**特点**：专门的Web启动脚本，支持开发和生产模式

### 方式5：手动命令行（推荐开发者）
```cmd
# 手动构建和运行
go build -o build\picoclaw.exe ./cmd/picoclaw
build\picoclaw.exe [命令]
```

---

## 📋 完整运行步骤

### 前置条件
1. **安装Go语言**
   - 下载地址：https://golang.org/dl/
   - 建议版本：Go 1.19+
   - 安装后重启命令提示符

2. **Git（可选）**
   - 下载地址：https://git-scm.com/
   - 用于克隆项目

### 第一次运行

#### 步骤1：获取项目
```cmd
# 方式A：Git克隆
git clone https://github.com/sipeed/picoclaw.git
cd picoclaw

# 方式B：下载ZIP解压
# 进入解压后的picoclaw目录
```

#### 步骤2：选择启动方式

##### 🔰 最简单方式（推荐）
```cmd
# 双击运行，一键启动Web界面
START.bat
```

##### 🛠️ 完整安装方式
```cmd
# 运行一键安装脚本
install-windows.bat
```
这个脚本会自动：
- ✅ 检查Go环境
- ✅ 下载依赖
- ✅ 复制workspace文件
- ✅ 构建可执行文件
- ✅ 安装到系统
- ✅ 创建桌面快捷方式

#### 步骤3：初始化配置（可选）
```cmd
# 初始化PicoClaw配置
picoclaw onboard
```

#### 步骤4：运行Web界面

##### 🌐 Web启动方式选择
```cmd
# 方式1：一键启动（最简单）
START.bat

# 方式2：完整Web服务
start-web.bat

# 方式3：开发模式（热重载）
quick-web.bat

# 方式4：交互式菜单选择
run-picoclaw.bat

# 方式5：使用内置Web服务
picoclaw gateway
```

### 日常运行

#### 方式A：一键启动（推荐）
```cmd
# 双击运行，自动启动Web界面
START.bat
```

#### 方式B：使用交互式菜单
```cmd
# 运行完整功能菜单
run-picoclaw.bat
```
菜单选项包括：
- 🚀 启动Web界面
- 💬 终端聊天
- 🌐 Web开发服务器
- 🔧 构建Web服务器
- ⚙️ 配置管理
- 📦 技能管理
- 🔐 认证管理

#### 方式C：使用命令行
```cmd
# 交互式AI助手
picoclaw agent

# 启动Web界面
picoclaw gateway

# 查看状态
picoclaw status

# 管理技能
picoclaw skills list
picoclaw skills install [skill-name]

# 查看帮助
picoclaw --help
```

#### 方式D：专用Web启动
```cmd
# 完整Web服务（生产模式）
start-web.bat

# 开发模式（热重载）
quick-web.bat

# 构建Web服务器
build-web-server.bat
```

#### 方式E：使用Web界面
1. 选择任意Web启动方式
2. 浏览器自动打开访问：http://localhost:8080
3. 在现代化Web界面中与AI对话

#### 方式F：使用桌面快捷方式
- 双击安装脚本创建的桌面快捷方式
- 或手动创建快捷方式指向 `START.bat`

---

## 🔧 常用命令详解

### 基础命令
```cmd
# 查看版本
picoclaw version

# 初始化配置
picoclaw onboard

# 查看状态
picoclaw status

# 直接与AI对话
picoclaw agent
```

### Web服务
```cmd
# 方式1：使用PicoClaw内置Web服务（推荐）
picoclaw gateway

# 方式2：使用Windows专用Web启动脚本
start-web.bat

# 方式3：使用开发模式（适合开发者）
quick-web.bat

# 方式4：使用交互式菜单
run-picoclaw.bat  # 选择选项1或3

# Web服务器默认端口8080
# 访问：http://localhost:8080
```

### 技能管理
```cmd
# 列出已安装技能
picoclaw skills list

# 列出内置技能
picoclaw skills list-builtin

# 安装技能
picoclaw skills install [skill-name]

# 安装内置技能
picoclaw skills install-builtin

# 移除技能
picoclaw skills remove [skill-name]

# 查看技能详情
picoclaw skills show [skill-name]

# 搜索技能
picoclaw skills search [keyword]
```

### 认证管理
```cmd
# 登录认证
picoclaw auth login

# 查看认证状态
picoclaw auth status

# 登出
picoclaw auth logout
```

### 定时任务
```cmd
# 管理定时任务
picoclaw cron list
picoclaw cron add [command]
picoclaw cron remove [id]
```

---

## 🌐 Web界面使用

### Web启动方式完整列表
```cmd
# 🔰 最简单方式
START.bat

# 🛠️ 完整Web服务
start-web.bat

# 🚀 开发模式（热重载）
quick-web.bat

# 🎛️ 交互式菜单选择
run-picoclaw.bat

# ⚡ 内置Web服务（无需Node.js）
picoclaw gateway

# 🔧 手动构建后运行
build-web-server.bat
build\web-server.exe
```

### 访问Web界面
1. 选择任意启动方式
2. 浏览器自动打开或手动访问
   - 生产模式：http://localhost:8080
   - 开发模式：http://localhost:5173
3. 开始与AI对话

### Web界面功能
- 💬 实时AI对话
- 🎨 现代化UI界面
- 📝 支持Markdown格式
- 🌙 主题切换（亮色/暗色）
- ⚙️ 模型配置和管理
- 📚 技能管理
- 🔄 实时思考过程显示
- 📤 JSON导出功能

### Web启动方式对比
| 启动方式 | 端口 | 特点 | 适用场景 |
|---------|------|------|----------|
| `START.bat` | 8080 | 一键启动，自动安装 | 新手用户 |
| `start-web.bat` | 8080 | 构建前端+后端 | 生产环境 |
| `quick-web.bat` | 5173 | 开发模式，热重载 | 开发者 |
| `run-picoclaw.bat` | 8080 | 交互式菜单 | 功能选择 |
| `picoclaw gateway` | 8080 | 无需Node.js | 简单使用 |

---

## 🔍 故障排除

### 问题1：命令不存在
```cmd
'picoclaw' 不是内部或外部命令...
```
**解决方案**：
1. 确保运行了 `install-windows.bat`
2. 检查PATH是否包含 `%USERPROFILE%\.local\bin`
3. 重启命令提示符
4. 或者使用 `START.bat` 无需安装

### 问题2：Web启动失败
```cmd
# 检查端口是否被占用
netstat -ano | findstr :8080

# 结束占用进程
taskkill /PID [进程号] /F

# 尝试其他启动方式
quick-web.bat          # 使用开发端口5173
picoclaw gateway       # 使用内置服务
```

### 问题3：构建失败
**解决方案**：
```cmd
# 运行修复脚本
fix-windows-build.bat

# 或重新安装
install-windows.bat

# 或查看详细构建指南
type WINDOWS_BUILD_FIX.md
```

### 问题4：Node.js相关问题
```cmd
# 检查Node.js安装
node --version

# 重新安装前端依赖
cd web
rmdir /s /q node_modules
npm install

# 或使用无需Node.js的方式
picoclaw gateway
```

### 问题5：配置文件错误
```cmd
# 重新初始化
picoclaw onboard

# 或删除配置重置
rmdir /s %USERPROFILE%\.picoclaw
picoclaw onboard
```

### 问题6：浏览器无法打开
- 手动访问：http://localhost:8080 或 http://localhost:5173
- 检查防火墙设置
- 尝试其他浏览器
- 确认Web服务器已启动

---

## 📁 重要文件位置

### Windows下的文件路径
```
%USERPROFILE%\.picoclaw\
├── config.json              # 主配置文件
├── workspace\               # 工作空间
│   ├── skills\             # 用户技能
│   └── ...
└── ...

%USERPROFILE%\.local\bin\
└── picoclaw.exe           # 可执行文件
```

### 项目目录
```
picoclaw\
├── cmd\picoclaw\           # 主程序源码
├── web\                   # Web界面
├── skills\                # 内置技能
├── build\                 # 构建输出
├── install-windows.bat     # 一键安装脚本
├── picoclaw-commands.bat   # 交互式菜单
├── run-picoclaw.bat      # 完整交互菜单
├── START.bat              # 一键Web启动
├── start-web.bat          # 完整Web服务
├── quick-web.bat         # 开发模式启动
├── build-web-server.bat   # 构建Web服务器
├── fix-windows-build.bat  # 构建修复脚本
└── ...
```

### 新增Windows专用脚本说明
| 脚本文件 | 作用 | 推荐用户 |
|---------|------|----------|
| `START.bat` | 一键启动Web界面 | 🔰 新手用户 |
| `run-picoclaw.bat` | 完整交互式菜单 | 🎛️ 功能探索者 |
| `start-web.bat` | 完整Web服务 | 🏭 生产环境 |
| `quick-web.bat` | 开发模式启动 | 🛠️ 开发者 |
| `build-web-server.bat` | 构建后端服务器 | 🔧 手动构建者 |
| `fix-windows-build.bat` | 修复构建问题 | 🐛 问题排查 |

---

## 💡 使用建议

### 🔰 新手用户（推荐）
1. **双击 `START.bat`** - 一键启动Web界面
2. **使用Web界面** - 现代化UI，无需命令行
3. **遇到问题** - 查看故障排除部分

### 🛠️ 开发者
1. **开发模式** - 使用 `quick-web.bat` 进行热重载开发
2. **命令行工具** - 熟悉 `picoclaw` 命令
3. **源码定制** - 查看Go源码进行二次开发

### 🏭 生产环境
1. **完整构建** - 使用 `start-web.bat` 构建生产版本
2. **自动化部署** - 使用 `build-web-server.bat` 构建可执行文件
3. **监控管理** - 使用 `run-picoclaw.bat` 进行状态管理

### 🎛️ 日常使用
1. **快速启动** - 桌面快捷方式指向 `START.bat`
2. **功能管理** - 使用 `run-picoclaw.bat` 的完整菜单
3. **技能扩展** - 通过Web界面或命令行管理技能

---

## 🆘 获取帮助

### 命令行帮助
```cmd
picoclaw --help              # 主帮助
picoclaw skills --help       # 技能帮助
picoclaw auth --help         # 认证帮助
```

### 快速启动帮助
```cmd
# 查看所有启动选项
run-picoclaw.bat

# 查看Web启动指南
type WEB_STARTUP_GUIDE.md

# 查看构建问题解决方案
type WINDOWS_BUILD_FIX.md
```

### 文档资源
- 📖 **[Windows运行指南](RUN_ON_WINDOWS.md)** (当前文档)
- 🌐 **[Web启动对比指南](WEB_STARTUP_GUIDE.md)** 
- 🛠️ **[Windows构建修复](WINDOWS_BUILD_FIX.md)**
- 📝 **[Windows命令行构建](WINDOWS_CLI_BUILD.md)**
- 📚 **[主README](README.md)**

### 社区支持
- GitHub Issues: https://github.com/sipeed/picoclaw/issues
- 官方文档: 查看 docs/ 目录
- 技能库: 查看 skills/ 目录

---

## 🎉 总结

现在您在Windows上有 **5种不同的启动方式**：

| 用户类型 | 推荐启动方式 | 特点 |
|---------|-------------|------|
| 🔰 **新手用户** | `START.bat` | 一键启动，无需配置 |
| 🛠️ **开发者** | `quick-web.bat` | 热重载，调试友好 |
| 🏭 **生产环境** | `start-web.bat` | 完整构建，稳定可靠 |
| 🎛️ **探索者** | `run-picoclaw.bat` | 功能齐全，交互式菜单 |
| ⚡ **简单使用** | `picoclaw gateway` | 无需Node.js，启动快速 |

**选择最适合您的启动方式，开始使用PicoClaw AI助手吧！** 🦞