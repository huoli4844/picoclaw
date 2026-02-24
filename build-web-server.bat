@echo off
REM 构建 PicoClaw Web 服务器

echo 🔨 构建 PicoClaw Web 服务器...
echo.

REM 检查Go环境
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ 错误: 请先安装 Go
    echo    访问: https://golang.org/
    pause
    exit /b 1
)

REM 创建构建目录
if not exist "build" mkdir "build"

REM 构建Web服务器
echo 正在构建 Web 服务器...
go build -ldflags "-s -w" -o build\web-server.exe cmd/web-server/main.go

if %errorlevel% equ 0 (
    echo ✅ Web 服务器构建成功!
    echo 输出文件: build\web-server.exe
    echo.
    echo 现在可以运行:
    echo   start-web.bat        # 启动完整Web服务
    echo   build\web-server.exe # 直接运行后端
) else (
    echo ❌ Web 服务器构建失败
)

echo.
pause