# macOS/Linux 构建修复指南

## 问题描述

在macOS/Linux上构建PicoClaw时遇到以下错误：
```
cmd/picoclaw/cmd_onboard.go:17:12: pattern workspace: no matching files found
make: *** [build] Error 1
```

## 原因分析

这个错误是因为：
1. **Go embed找不到workspace文件** - `//go:embed workspace` 在`cmd/picoclaw`目录下找不到workspace
2. **工作目录问题** - Go embed需要文件在正确的相对位置
3. **generate命令不完整** - 需要在构建前复制workspace文件

## ✅ 解决方案

### 方案1：使用修复后的Makefile（推荐）
```bash
# 现在可以直接使用，已自动修复
make build
```

### 方案2：手动修复
```bash
# 1. 复制workspace到正确位置
cp -r workspace cmd/picoclaw/

# 2. 构建
make build
```

### 方案3：完整重新构建
```bash
# 1. 清理
make clean

# 2. 确保workspace存在
ls workspace/

# 3. 构建（现在会自动复制workspace）
make build
```

## 🔧 修复的Makefile变更

我已经修复了Makefile的generate部分：

```makefile
## generate: Run generate
generate:
	@echo "Run generate..."
	@rm -r ./$(CMD_DIR)/workspace 2>/dev/null || true
	@cp -r workspace ./$(CMD_DIR)/ 2>/dev/null || echo "Warning: workspace directory not found"
	@$(GO) generate ./...
	@echo "Run generate complete"
```

**关键变更**：
- ✅ 在`go generate`之前复制workspace文件
- ✅ 添加错误处理和警告信息
- ✅ 确保embed能找到文件

## 🧪 验证构建

### 检查构建结果
```bash
# 查看构建文件
ls -la build/

# 测试版本
./build/picoclaw --version
```

### 检查workspace文件
```bash
# 确认workspace已复制
ls -la cmd/picoclaw/workspace/

# 检查关键文件
ls cmd/picoclaw/workspace/USER.md
```

## 🐛 常见问题

### 问题1：workspace目录不存在
```bash
# 检查workspace
ls workspace/

# 如果不存在，创建基础文件
mkdir -p workspace/memory workspace/skills
touch workspace/USER.md workspace/AGENT.md workspace/IDENTITY.md workspace/SOUL.md
```

### 问题2：权限问题
```bash
# 检查权限
ls -la workspace/

# 修复权限
chmod -R 755 workspace/
```

### 问题3：Go版本问题
```bash
# 检查Go版本
go version

# 推荐：Go 1.19+
```

### 问题4：依赖问题
```bash
# 重新下载依赖
go mod download
go mod tidy
```

## 🔄 构建工作流

### 日常构建
```bash
# 快速构建（现在应该成功）
make build
```

### 完整构建
```bash
# 清理 + 构建
make clean && make build
```

### 多平台构建
```bash
# 构建所有平台
make build-all
```

### 安装构建
```bash
# 安装到系统
make install
```

## 📁 文件结构

构建成功后的目录结构：
```
picoclaw/
├── cmd/picoclaw/
│   ├── workspace/              # ✅ 现在存在
│   │   ├── USER.md
│   │   ├── AGENT.md
│   │   └── ...
│   ├── main.go
│   └── cmd_*.go
├── build/
│   ├── picoclaw              # 软链接
│   └── picoclaw-darwin-arm64 # 实际二进制
└── workspace/                # 原始workspace
    └── ...
```

## 📊 构建输出示例

成功的构建应该显示：
```
Run generate...
Warning: workspace directory not found  # 只有当workspace不存在时
Run generate complete
Building picoclaw for darwin/arm64...
[Go构建输出...]
Build complete: build/picoclaw-darwin-arm64
```

## 🎯 测试构建

### 基本测试
```bash
# 版本测试
./build/picoclaw --version

# 帮助测试
./build/picoclaw --help
```

### 功能测试
```bash
# onboard测试
./build/picoclaw onboard

# agent测试
./build/picoclaw agent
```

## 💡 提示

1. **首次构建**：建议先运行 `make clean && make build`
2. **日常构建**：直接运行 `make build` 即可
3. **workspace修改**：修改workspace内容后需要重新构建
4. **embed问题**：如果仍有embed错误，检查 `cmd/picoclaw/workspace/` 是否存在

## 🆘 如果仍有问题

1. **清理重建**：
   ```bash
   make clean
   rm -rf cmd/picoclaw/workspace
   cp -r workspace cmd/picoclaw/
   make build
   ```

2. **检查Go环境**：
   ```bash
   go env
   which go
   ```

3. **查看详细错误**：
   ```bash
   make build VERBOSE=1
   ```

---

**现在macOS/Linux用户应该能够成功构建PicoClaw了！** 🦞

如果仍有问题，请检查workspace目录是否正确复制到 `cmd/picoclaw/` 下。