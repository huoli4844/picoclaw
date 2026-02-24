@echo off
REM PicoClaw Web 启动脚本 - Windows版本
REM 对应 start-web.sh 的Windows实现

setlocal enabledelayedexpansion

echo 🦞 PicoClaw Web 启动脚本
echo ==========================

REM 检查是否已安装 Node.js
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ 错误: 请先安装 Node.js ^(推荐 18+ 版本^)
    echo    访问: https://nodejs.org/
    pause
    exit /b 1
)

REM 检查 Go 环境
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ 错误: 请先安装 Go
    echo    访问: https://golang.org/
    pause
    exit /b 1
)

REM 检查是否在项目根目录
if not exist "web" (
    echo ❌ 错误: 请在 PicoClaw 项目根目录运行此脚本
    pause
    exit /b 1
)

echo 📦 安装前端依赖...
cd web
if not exist "node_modules" (
    echo 正在安装依赖...
    call npm install
    if %errorlevel% neq 0 (
        echo ❌ 前端依赖安装失败
        pause
        exit /b 1
    )
) else (
    echo ✅ 前端依赖已安装
)

echo 🔨 构建前端...
call npm run build
if %errorlevel% neq 0 (
    echo ❌ 前端构建失败
    pause
    exit /b 1
)

echo 🚀 启动后端服务器...
cd ..

REM 检查后端是否已存在
if exist "build\web-server.exe" (
    echo 使用已构建的后端...
    set BACKEND_EXE=build\web-server.exe
) else (
    echo 实时运行后端...
    set BACKEND_EXE=go run cmd/web-server/main.go
)

echo 正在启动后端服务器...
start "PicoClaw Backend" cmd /c "echo 后端服务器运行中... && echo 按 Ctrl+C 停止服务器 && %BACKEND_EXE% && pause"

REM 等待后端启动
echo 等待服务器启动...
timeout /t 3 /nobreak >nul

echo.
echo 🌐 服务地址:
echo    前端界面: http://localhost:8080
echo    API 接口:  http://localhost:8080/api
echo.
echo 💡 提示:
echo    - 请确保已配置 %USERPROFILE%\.picoclaw\config.json
echo    - 按 Ctrl+C 在后端窗口停止服务器
echo.

REM 尝试打开浏览器
echo 正在打开浏览器...
start http://localhost:8080

echo.
echo ✅ PicoClaw Web 服务已启动!
echo.
echo 如需完全停止服务，请关闭后端服务器窗口。
echo 按任意键关闭此窗口...
pause >nul