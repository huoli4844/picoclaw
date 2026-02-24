# Windows 批处理命令说明

由于Windows系统没有make命令，我为您创建了两个批处理文件来替代Makefile的功能。

## 文件说明

### 1. `build.bat`
这是`Makefile`的直接替代品，使用相同的命令语法：

```cmd
# 安装依赖
build.bat install

# 启动开发服务器
build.bat dev

# 构建生产版本
build.bat build

# 清理构建文件
build.bat clean

# 显示帮助信息
build.bat help
```

### 2. `web-commands.bat` (推荐)
这是一个交互式菜单版本，更用户友好：

```cmd
# 直接运行，会显示菜单选项
web-commands.bat
```

运行后会显示菜单：
1. Install dependencies
2. Start development server  
3. Build for production
4. Clean build files
5. Exit

## 使用方法

### 方法一：命令行模式 (build.bat)
```cmd
# 安装依赖
build.bat install

# 开发模式
build.bat dev

# 构建项目
build.bat build
```

### 方法二：菜单模式 (web-commands.bat) 
```cmd
# 双击运行或在命令行执行
web-commands.bat
# 然后选择对应的数字选项
```

## 注意事项

1. 确保已安装Node.js和npm
2. 在`web`目录下运行这些批处理文件
3. 首次使用前建议先运行`build.bat install`安装依赖
4. 开发时使用`build.bat dev`启动热重载服务器
5. 部署前使用`build.bat build`构建生产版本

## 与Makefile的对应关系

| Make命令 | Bat命令 | 说明 |
|---------|---------|------|
| `make install` | `build.bat install` | 安装依赖 |
| `make dev` | `build.bat dev` | 开发模式 |
| `make build` | `build.bat build` | 构建生产 |
| `make clean` | `build.bat clean` | 清理文件 |
| `make help` | `build.bat help` | 帮助信息 |

现在您可以在Windows系统上正常使用这些命令进行开发了！