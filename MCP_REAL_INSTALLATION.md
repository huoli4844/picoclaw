# MCP 真实安装功能说明

## 概述

现在 MCP 安装功能已经升级为**真实安装**，不再是简单的配置记录。当用户点击"安装"按钮时，系统会：

1. **真实下载并安装** MCP 服务器软件包
2. **验证安装** 确保服务器可以正常工作
3. **持续监控** 已安装服务器的健康状态

## 支持的安装类型

### 1. NPM 包安装
适用于基于 Node.js 的 MCP 服务器：
- `@modelcontextprotocol/server-git`
- `@mcp-community/database-server`
- `@mcp-community/time-server`

**安装过程：**
- 自动检测 `npm` 或 `yarn` 可用性
- 使用 `npm install -g` 全局安装包
- 失败时自动尝试 `yarn global add`
- 安装后验证 `npx` 命令可用性

### 2. 独立命令安装
适用于预编译的 MCP 服务器：
- `mcp-server-filesystem`

**安装过程：**
- 检查命令是否存在于 PATH 中
- 如果不存在，提供安装指导
- 尝试运行命令验证可用性

### 3. Python 包安装（预留）
为未来的 Python MCP 服务器预留接口。

## 安装验证机制

### 实时验证
- 安装完成后立即运行验证
- 检查命令是否真正可用
- 测试基本功能（如 `--help` 或 `--version`）

### 持续健康检查
- 定期验证已安装服务器状态
- 检测依赖是否被意外删除
- 及时报告安装问题

## 新增 API 端点

### POST `/api/mcp/servers/{id}/validate`
验证已安装服务器的健康状态

**响应示例：**
```json
{
  "success": true,
  "data": {
    "status": "success",
    "message": "Server filesystem installation is valid",
    "server": { ... }
  }
}
```

## 用户界面改进

### 安装按钮状态
- **安装中**: 显示下载和安装进度
- **验证中**: 显示安装后验证状态
- **已安装**: 服务器可用并就绪

### 错误处理
- 详细的错误信息显示
- 安装失败的自动回滚
- 针对性的修复建议

## 环境要求

### 必需依赖
- **Node.js 和 npm**: 用于安装基于 npm 的 MCP 服务器
- **Git**: 用于 Git 相关的 MCP 服务器

### 可选依赖
- **Yarn**: 作为 npm 的替代包管理器
- **Python 和 pip**: 用于未来的 Python MCP 服务器

## 安装前检查

系统会自动检查：
1. 包管理器可用性
2. 网络连接状态
3. 权限是否足够进行全局安装

## 故障排除

### 常见问题

#### 1. npm 权限错误
```bash
# 解决方案：配置 npm 全局目录
npm config set prefix ~/.npm-global
export PATH=~/.npm-global/bin:$PATH
```

#### 2. 网络连接问题
```bash
# 解决方案：使用国内镜像源
npm config set registry https://registry.npmmirror.com
```

#### 3. 命令未找到
```bash
# 解决方案：检查 PATH 环境变量
echo $PATH
# 确保包含 npm 全局安装目录
```

## 测试验证

运行测试脚本：
```bash
chmod +x test_mcp_install.sh
./test_mcp_install.sh
```

该脚本会：
1. 检查环境依赖
2. 测试真实安装流程
3. 验证安装结果
4. 测试 API 功能

## 配置文件位置

- **安装目录**: `{storage_path}/servers/{server_id}/`
- **配置文件**: `{storage_path}/servers/{server_id}/server.json`
- **日志文件**: 安装过程中的实时输出

## 安全注意事项

1. **全局安装**: 需要适当的系统权限
2. **网络下载**: 确保网络安全和来源可信
3. **命令执行**: 只执行已知安全的 MCP 服务器命令

## 未来改进计划

1. **沙箱安装**: 在隔离环境中安装和测试
2. **自动更新**: 支持已安装服务器的自动更新
3. **依赖管理**: 更智能的依赖冲突解决
4. **多版本支持**: 同时安装同一服务器的多个版本

---

现在，当用户点击 MCP 安装按钮时，他们可以确信系统正在执行真实的软件安装和验证！