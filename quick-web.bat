@echo off
REM PicoClaw 快速Web启动 - 最简单的方式

echo 🦞 启动 PicoClaw Web...

REM 进入Web目录并启动开发服务器
cd web

REM 检查依赖
if not exist "node_modules" (
    echo 安装依赖...
    call npm install
)

REM 启动开发服务器
echo 启动开发服务器...
echo 访问: http://localhost:5173
echo 按 Ctrl+C 停止
echo.
call npm run dev

pause