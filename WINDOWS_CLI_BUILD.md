# Windows 命令行构建指南

## 🚀 快速开始

在Windows上正确构建PicoClaw的命令：

### ✅ 正确的构建方式
```cmd
# 在项目根目录执行
go build -o build\picoclaw.exe ./cmd/picoclaw
```

### ❌ 错误的构建方式
```cmd
# 这些都会导致 undefined 错误
go build .\main.go
go build .\cmd\picoclaw\main.go
cd cmd\picoclaw && go build main.go
```

## 📁 目录结构理解

PicoClaw的项目结构：
```
picoclaw/
├── cmd/
│   └── picoclaw/           # 主包目录
│       ├── main.go          # 主入口文件
│       ├── cmd_agent.go     # 代理命令
│       ├── cmd_auth.go      # 认证命令
│       ├── cmd_onboard.go   # 初始化命令
│       └── ...             # 其他命令文件
├── go.mod                  # Go模块文件
└── build.bat              # Windows构建脚本
```

## 🔧 完整的构建步骤

### 1. 环境检查
```cmd
# 检查Go安装
go version

# 检查当前目录
pwd
# 应该在 picoclaw 项目根目录
```

### 2. 依赖处理
```cmd
# 下载依赖
go mod download

# 整理依赖
go mod tidy
```

### 3. 复制workspace文件
```cmd
# Windows下复制workspace（解决embed问题）
if exist workspace (
    xcopy /E /I /Y workspace cmd\picoclaw\workspace
)
```

### 4. 构建应用
```cmd
# 创建构建目录
if not exist build mkdir build

# 构建应用
go build -v -tags stdjson -ldflags "-s -w" -o build\picoclaw.exe ./cmd/picoclaw
```

### 5. 测试构建
```cmd
# 测试构建结果
build\picoclaw.exe --version
```

## 🛠️ 不同构建选项

### 开发构建（快速）
```cmd
go build -o build\picoclaw.exe ./cmd/picoclaw
```

### 生产构建（优化）
```cmd
go build -v -tags stdjson -ldflags "-s -w" -o build\picoclaw.exe ./cmd/picoclaw
```

### 调试构建（详细）
```cmd
go build -v -x -o build\picoclaw.exe ./cmd/picoclaw
```

## 🐛 故障排除

### 错误1: `undefined: xxx`
**原因**: 构建单个文件而不是整个包
**解决**: 确保使用 `./cmd/picoclaw` 而不是单个文件路径

### 错误2: `pattern workspace: no matching files found`
**原因**: Go embed找不到workspace文件
**解决**: 复制workspace文件
```cmd
xcopy /E /I /Y workspace cmd\picoclaw\workspace
```

### 错误3: `go: cannot find main module`
**原因**: 不在正确的目录
**解决**: 确保在项目根目录，并且有go.mod文件

### 错误4: 权限错误
**原因**: 防火墙或杀毒软件阻止
**解决**: 以管理员身份运行或添加例外

## 🎯 IDE配置

### GoLand配置
1. **项目根目录**: 设置为picoclaw目录
2. **运行配置**:
   - **Files**: `cmd/picoclaw`
   - **Output directory**: `build`
   - **Program arguments**: 你的参数

### VS Code配置
1. **工作目录**: picoclaw根目录
2. **终端命令**: `go build -o build\picoclaw.exe ./cmd/picoclaw`

## 📋 一键构建脚本

创建 `quick-build.bat` 文件：
```batch
@echo off
echo Quick Build for PicoClaw
echo =========================

REM 检查Go
where go >nul 2>nul || (
    echo Error: Go not found in PATH
    exit /b 1
)

REM 复制workspace
if exist workspace (
    xcopy /E /I /Y workspace cmd\picoclaw\workspace
    echo Workspace copied
)

REM 构建
if not exist build mkdir build
go build -v -tags stdjson -ldflags "-s -w" -o build\picoclaw.exe ./cmd/picoclaw

if %errorlevel% equ 0 (
    echo Build successful!
    echo Output: build\picoclaw.exe
) else (
    echo Build failed!
)
```

## 🔄 日常开发工作流

### 开发时
```cmd
# 修改代码后快速构建
go build -o build\picoclaw.exe ./cmd/picoclaw

# 测试
build\picoclaw.exe version
```

### 提交前
```cmd
# 完整检查
go mod tidy
go build -v -tags stdjson -ldflags "-s -w" -o build\picoclaw.exe ./cmd/picoclaw
go test ./...
```

### 发布前
```cmd
# 使用完整脚本
build.bat build-all
```

## 💡 提示和技巧

### 1. 路径使用正斜杠
在Windows上Go构建命令中，路径使用正斜杠：
```cmd
# ✅ 正确
go build ./cmd/picoclaw

# ❌ 可能有问题（但在某些情况也可用）
go build .\cmd\picoclaw
```

### 2. 相对路径
始终从项目根目录使用相对路径，不要切换到子目录。

### 3. 构建标志
- `-v`: 详细输出
- `-x`: 显示执行的命令
- `-tags stdjson`: 使用标准JSON
- `-ldflags "-s -w"`: 减少二进制大小

### 4. 环境变量
可以设置构建环境变量：
```cmd
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -o build\picoclaw.exe ./cmd/picoclaw
```

---

**记住：在Windows上构建PicoClaw的关键是构建整个包 `./cmd/picoclaw`，而不是单个文件！**