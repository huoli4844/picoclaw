# PicoClaw Windows 构建指南

由于Windows系统没有make命令，我们创建了批处理文件来替代Makefile的功能。

## 文件说明

### 1. `build.bat` - 命令行版本
这是`Makefile`的直接Windows替代品，支持所有原有的命令。

### 2. `picoclaw-commands.bat` - 交互式菜单版本（推荐）
提供用户友好的菜单界面，适合不熟悉命令行的用户。

## 前置要求

1. **安装Go语言环境**
   - 下载地址：https://golang.org/dl/
   - 确保Go已添加到系统PATH中

2. **安装Git（可选，用于版本信息）**
   - 下载地址：https://git-scm.com/

## 使用方法

### 方法一：命令行模式 (build.bat)

```cmd
# 构建当前平台
build.bat build

# 构建所有平台
build.bat build-all

# 安装到系统
build.bat install

# 卸载
build.bat uninstall

# 完全卸载（包括数据）
build.bat uninstall-all

# 清理构建文件
build.bat clean

# 运行测试
build.bat test

# 下载依赖
build.bat deps

# 格式化代码
build.bat fmt

# 运行代码检查
build.bat lint

# 完整检查（依赖+格式+静态分析+测试）
build.bat check

# 构建并运行
build.bat run [参数]

# 查看帮助
build.bat help
```

### 方法二：菜单模式 (picoclaw-commands.bat) 推荐使用

直接双击运行或在命令行执行：
```cmd
picoclaw-commands.bat
```

然后选择对应的数字选项即可。

## Windows特有的注意事项

### 1. 路径差异
- **原版Makefile**: `~/.local/bin` (Linux/Mac)
- **Windows版本**: `%USERPROFILE%\.local\bin`

### 2. 文件扩展名
- **Windows**: 可执行文件使用`.exe`扩展名
- **Linux/Mac**: 无扩展名

### 3. 环境变量设置
安装后会提示将以下路径添加到系统PATH：
```
%USERPROFILE%\.local\bin
```

## 命令对照表

| Make命令 | Bat命令 | 说明 |
|---------|---------|------|
| `make build` | `build.bat build` | 构建当前平台 |
| `make build-all` | `build.bat build-all` | 构建所有平台 |
| `make install` | `build.bat install` | 安装到系统 |
| `make uninstall` | `build.bat uninstall` | 卸载可执行文件 |
| `make uninstall-all` | `build.bat uninstall-all` | 完全卸载 |
| `make clean` | `build.bat clean` | 清理构建文件 |
| `make test` | `build.bat test` | 运行测试 |
| `make fmt` | `build.bat fmt` | 格式化代码 |
| `make lint` | `build.bat lint` | 代码检查 |
| `make deps` | `build.bat deps` | 下载依赖 |
| `make check` | `build.bat check` | 完整检查 |
| `make run` | `build.bat run` | 构建并运行 |

## 构建输出

### 当前平台构建
- 输出路径：`build\picoclaw-windows-x64.exe`
- 同时创建：`build\picoclaw.exe` (便于使用)

### 多平台构建
- `build\picoclaw-linux-amd64` (Linux 64位)
- `build\picoclaw-linux-arm64` (Linux ARM64)
- `build\picoclaw-darwin-arm64` (macOS ARM64)
- `build\picoclaw-windows-amd64.exe` (Windows 64位)

## 安装目录结构

```
%USERPROFILE%\.local\
├── bin\
│   └── picoclaw.exe          # 主程序
└── share\
    └── man\
        └── man1\
```

## 配置和数据目录

Windows下的配置和数据存储在：
```
%USERPROFILE%\.picoclaw\
├── config.json               # 配置文件
├── workspace\                # 工作空间
│   └── skills\               # 技能目录
└── ...                       # 其他数据
```

## 故障排除

### 1. Go未找到
```
错误: Go is not installed or not in PATH
```
**解决方案**: 安装Go并添加到PATH环境变量

### 2. 权限问题
```
Access denied
```
**解决方案**: 以管理员身份运行命令提示符

### 3. 依赖问题
```
go: module not found
```
**解决方案**: 运行 `build.bat deps` 下载依赖

### 4. 构建错误：`pattern workspace: no matching files found`
**这是Windows特定问题，由于Go embed找不到workspace文件**

**快速修复**：
```cmd
# 运行专用修复脚本
fix-windows-build.bat
```

**手动修复**：
```cmd
# 确保workspace目录存在
if exist workspace (
    xcopy /E /I /Y workspace cmd\picoclaw\workspace
)
# 然后重新构建
build.bat build
```

### 5. 安装错误：`系统找不到指定的文件`
**这通常是因为构建失败导致可执行文件不存在**

**解决方案**：
1. 先运行 `fix-windows-build.bat` 修复构建问题
2. 确认构建成功后再运行 `build.bat install`

### 6. 完整问题解决流程
```cmd
# 1. 运行修复脚本
fix-windows-build.bat

# 2. 如果修复成功，安装程序
build.bat install

# 3. 如果仍有问题，尝试完全清理重建
build.bat clean
build.bat build
build.bat install
```

## 开发建议

1. **日常开发使用**：
   - 使用 `picoclaw-commands.bat` 的交互菜单
   
2. **CI/CD自动化**：
   - 使用 `build.bat` 的命令行版本

3. **首次设置**：
   - 运行 `build.bat install` 安装到系统
   - 将 `%USERPROFILE%\.local\bin` 添加到PATH

4. **代码质量**：
   - 提交前运行 `build.bat check` 进行完整检查

现在您可以在Windows系统上完整地使用PicoClaw的所有构建功能了！