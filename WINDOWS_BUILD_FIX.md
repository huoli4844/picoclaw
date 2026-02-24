# Windows 构建错误解决方案

## 问题：`undefined: xxx` 错误

当在Windows上直接执行以下命令时：
```cmd
go build .\main.go
```

会出现类似这样的错误：
```
.\main.go:105:3: undefined: onboard
.\main.go:107:3: undefined: agentCmd
.\main.go:109:3: undefined: gatewayCmd
...\很多错误
```

## 原因分析

这个错误是因为：
1. **Go包结构问题** - PicoClaw使用多文件包结构，每个命令在单独的文件中
2. **错误的构建方式** - 直接构建单个文件时，Go不会包含其他相关的Go文件

## 正确的解决方案

### 方案1：使用推荐的构建脚本（推荐）

```cmd
# 运行专门的Windows安装脚本
install-windows.bat
```

### 方案2：手动构建（正确方式）

```cmd
# 进入项目根目录
cd D:\GolandProjects\picoclaw

# 构建整个包，而不是单个文件
go build -o build\picoclaw.exe ./cmd/picoclaw

# 或者使用完整构建选项
go build -v -tags stdjson -ldflags "-s -w" -o build\picoclaw.exe ./cmd/picoclaw
```

### 方案3：使用修复脚本

```cmd
# 运行修复脚本
fix-windows-build.bat
```

## 常见的错误做法

❌ **错误的构建命令：**
```cmd
go build .\main.go                    # 只构建单个文件
go build .\cmd\picoclaw\main.go       # 只构建单个文件
cd cmd\picoclaw && go build main.go   # 在子目录中构建单个文件
```

✅ **正确的构建命令：**
```cmd
go build ./cmd/picoclaw              # 构建整个包
go build -o picoclaw.exe ./cmd/picoclaw  # 指定输出文件名
```

## 为什么PicoClaw使用多文件结构？

PicoClaw采用模块化设计，每个功能在单独的文件中：

| 文件 | 功能 | 主要函数 |
|------|------|---------|
| `cmd_onboard.go` | 初始化配置 | `onboard()` |
| `cmd_agent.go` | 代理交互 | `agentCmd()` |
| `cmd_gateway.go` | 网关服务 | `gatewayCmd()` |
| `cmd_status.go` | 状态检查 | `statusCmd()` |
| `cmd_auth.go` | 认证管理 | `authCmd()` |
| `cmd_cron.go` | 定时任务 | `cronCmd()` |
| `cmd_skills.go` | 技能管理 | `skillsHelp()`, `skillsListCmd()` 等 |
| `cmd_migrate.go` | 数据迁移 | `migrateCmd()` |

这种方式的好处：
- 代码组织清晰
- 便于维护
- 每个功能独立

## 在不同IDE中的正确配置

### GoLand (推荐)
1. 确保项目根目录正确设置
2. 配置Run/Debug为：`./cmd/picoclaw`
3. 不要设置为单个文件

### VS Code
1. 在项目根目录工作
2. 使用终端命令：`go build ./cmd/picoclaw`

### 命令行
1. 确保在项目根目录
2. 使用正确的包路径

## 完整的Windows开发工作流

### 1. 首次设置
```cmd
# 克隆项目
git clone https://github.com/sipeed/picoclaw.git
cd picoclaw

# 运行一键安装
install-windows.bat
```

### 2. 日常开发
```cmd
# 修改代码后，重新构建
go build -o build\picoclaw.exe ./cmd/picoclaw

# 或者使用开发脚本
dev.bat build
dev.bat run
```

### 3. 发布构建
```cmd
# 完整构建
build.bat build-all
```

## 调试构建问题

如果仍然遇到问题，按以下步骤排查：

### 1. 检查Go环境
```cmd
go version
go env GOPATH
go env GOROOT
```

### 2. 检查项目结构
```cmd
dir cmd\picoclaw\*.go
```
应该看到多个.go文件，不只是main.go

### 3. 检查模块文件
```cmd
dir go.mod
```
应该在项目根目录找到go.mod文件

### 4. 依赖检查
```cmd
go mod tidy
go mod download
```

### 5. 详细构建信息
```cmd
go build -v -x ./cmd/picoclaw
```
`-v`显示详细信息，`-x`显示执行命令

## 常见问题FAQ

**Q: 为什么不能直接构建main.go？**
A: Go需要包中的所有文件来解析依赖关系，main.go引用了其他文件中的函数。

**Q: 在IDE中如何设置？**
A: 设置工作目录为项目根目录，构建目标为`./cmd/picoclaw`包，而不是单个文件。

**Q: 构建成功但运行时报错？**
A: 检查workspace文件是否正确复制，运行`install-windows.bat`会处理这个问题。

---

**总结：在Windows上开发PicoClaw，关键是构建整个包`./cmd/picoclaw`，而不是单个文件。使用提供的脚本可以避免大部分问题。**