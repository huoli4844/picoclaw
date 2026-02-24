# PicoClaw Windows 支持

## 🎉 完整的Windows支持

PicoClaw现在完全支持Windows开发！我们提供了批处理文件来替代Unix系统中的Makefile。

## 📁 新增的Windows文件

| 文件 | 说明 | 用途 |
|------|------|------|
| `build.bat` | 命令行构建脚本 | Makefile的完整Windows替代品 |
| `picoclaw-commands.bat` | 交互式菜单 | 用户友好的图形界面选择 |
| `dev.bat` | 快速开发脚本 | 常用开发任务的快捷方式 |
| `WINDOWS_BUILD_GUIDE.md` | 详细构建指南 | 完整的Windows使用文档 |

## 🚀 快速开始

### 1. 安装Go
从 https://golang.org/dl/ 下载并安装Go for Windows

### 2. 克隆项目
```cmd
git clone https://github.com/sipeed/picoclaw.git
cd picoclaw
```

### 3. 选择使用方式

#### 方式A：交互式菜单（推荐新手）
```cmd
picoclaw-commands.bat
```
选择数字即可执行对应操作。

#### 方式B：命令行（推荐开发者）
```cmd
# 构建
build.bat build

# 安装
build.bat install

# 运行
build.bat run
```

#### 方式C：快速开发
```cmd
# 快速检查代码质量
dev.bat check

# 快速测试
dev.bat test

# 快速构建并运行
dev.bat run
```

## 🎯 常用命令对照

| 原命令 | Windows命令 |
|--------|------------|
| `make build` | `build.bat build` |
| `make install` | `build.bat install` |
| `make test` | `build.bat test` |
| `make clean` | `build.bat clean` |
| `make run` | `build.bat run` |

## 📂 安装位置

- **可执行文件**: `%USERPROFILE%\.local\bin\picoclaw.exe`
- **配置文件**: `%USERPROFILE%\.picoclaw\config.json`
- **工作空间**: `%USERPROFILE%\.picoclaw\workspace`

## ⚠️ 重要提醒

1. **PATH设置**: 安装后需要将 `%USERPROFILE%\.local\bin` 添加到系统PATH
2. **管理员权限**: 某些操作可能需要管理员权限
3. **防火墙**: 首次运行可能需要允许通过防火墙

## 🐛 故障排除

### 问题：命令找不到
**解决方案**:
1. 确认Go已正确安装
2. 将 `%USERPROFILE%\.local\bin` 添加到PATH
3. 重启命令提示符

### 问题：权限被拒绝
**解决方案**:
1. 以管理员身份运行命令提示符
2. 检查文件是否被其他程序占用

### 问题：依赖下载失败
**解决方案**:
```cmd
# 设置代理（如果在企业网络）
set GOPROXY=https://goproxy.cn,direct
# 然后运行
build.bat deps
```

## 🔄 自动化脚本

### 每日开发流程
```cmd
# 一键完整检查
dev.bat check

# 如果检查通过，构建并测试
dev.bat build
dev.bat test

# 最后运行
dev.bat run
```

### 提交前检查
```cmd
# 完整代码质量检查
build.bat check

# 格式化代码
build.bat fmt

# 运行所有测试
build.bat test
```

## 📚 更多信息

- 📖 [详细构建指南](WINDOWS_BUILD_GUIDE.md)
- 🛠️ [开发工具说明](#dev.bat-使用说明)
- ❓ [常见问题解答](#故障排除)

---

**现在您可以完全在Windows上开发和运行PicoClaw了！** 🎊

如有问题，请查看 `WINDOWS_BUILD_GUIDE.md` 获取详细信息。