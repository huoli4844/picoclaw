@echo off
REM PicoClaw Quick Start Script for Windows
REM This script helps you run PicoClaw easily

setlocal enabledelayedexpansion

title PicoClaw AI Assistant

:main_menu
cls
echo.
echo 🦞 PicoClaw AI Assistant - Windows Quick Start
echo ================================================
echo.
echo Current Status:
for /f "tokens=2 delims=: " %%a in ('sc query picoclaw 2^>nul ^| findstr "STATE"') do set service_state=%%a
if "!service_state!"=="RUNNING" (
    echo ✅ Web Server: Running
    echo 🌐 Web Interface: http://localhost:8080
) else (
    echo ❌ Web Server: Not running
)
echo.

echo Choose what you want to do:
echo.
echo 1. 🚀 Start Web Interface (Recommended)
echo 2. 💬 Chat with AI in Terminal
echo 3. 🌐 Start Web Development Server (npm dev)
echo 4. 🔧 Build Web Server
echo 3. ⚙️  Initialize/Update Configuration
echo 4. 📊 Show Status
echo 5. 🔧 Manage Skills
echo 6. 🔐 Authentication Management
echo 7. 📦 Install PicoClaw (First time setup)
echo 8. 🌐 Open Web Browser
echo 9. 🛠️  Advanced Options
echo 0. 🚪 Exit
echo.
set /p choice="Enter your choice (0-9): "

if "%choice%"=="1" goto start_web
if "%choice%"=="2" goto chat_terminal
if "%choice%"=="3" goto start_web_dev
if "%choice%"=="4" goto build_web_server
if "%choice%"=="3" goto init_config
if "%choice%"=="4" goto show_status
if "%choice%"=="5" goto manage_skills
if "%choice%"=="6" goto auth_management
if "%choice%"=="7" goto install_picoclaw
if "%choice%"=="8" goto open_browser
if "%choice%"=="9" goto advanced_options
if "%choice%"=="0" goto exit
goto invalid_choice

:start_web
echo.
echo 🌐 Starting PicoClaw Web Server...
echo.
if not exist "%USERPROFILE%\.local\bin\picoclaw.exe" (
    echo ❌ PicoClaw not found. Installing first...
    call :install_picoclaw_silent
)
echo Starting web server on http://localhost:8080
echo Press Ctrl+C to stop the server
echo.
start http://localhost:8080
"%USERPROFILE%\.local\bin\picoclaw.exe" gateway
goto main_menu

:chat_terminal
echo.
echo 💬 Starting AI Chat in Terminal...
echo.
if not exist "%USERPROFILE%\.local\bin\picoclaw.exe" (
    echo ❌ PicoClaw not found. Installing first...
    call :install_picoclaw_silent
)
echo Type your messages below. Type 'exit' to quit.
echo.
"%USERPROFILE%\.local\bin\picoclaw.exe" agent
pause
goto main_menu

:start_web_dev
echo.
echo 🌐 Starting Web Development Server...
echo.
cd web
if not exist "node_modules" (
    echo Installing dependencies...
    call npm install
)
echo Starting development server on http://localhost:5173
echo Press Ctrl+C to stop
echo.
start http://localhost:5173
call npm run dev
cd ..
goto main_menu

:build_web_server
echo.
echo 🔧 Building Web Server...
echo.
call build-web-server.bat
goto main_menu

:init_config
echo.
echo ⚙️ Initializing PicoClaw Configuration...
echo.
if not exist "%USERPROFILE%\.local\bin\picoclaw.exe" (
    echo ❌ PicoClaw not found. Installing first...
    call :install_picoclaw_silent
)
"%USERPROFILE%\.local\bin\picoclaw.exe" onboard
pause
goto main_menu

:show_status
echo.
echo 📊 PicoClaw Status...
echo.
if exist "%USERPROFILE%\.local\bin\picoclaw.exe" (
    "%USERPROFILE%\.local\bin\picoclaw.exe" status
) else (
    echo ❌ PicoClaw not installed. Please install first.
)
pause
goto main_menu

:manage_skills
echo.
echo 📦 Skill Management Menu
echo =====================
echo 1. List installed skills
echo 2. List builtin skills  
echo 3. Install a skill
echo 4. Install builtin skills
echo 5. Remove a skill
echo 6. Search for skills
echo 0. Back to main menu
echo.
set /p skill_choice="Enter your choice (0-6): "

if "%skill_choice%"=="1" (
    echo.
    echo Installed skills:
    "%USERPROFILE%\.local\bin\picoclaw.exe" skills list
)
if "%skill_choice%"=="2" (
    echo.
    echo Builtin skills:
    "%USERPROFILE%\.local\bin\picoclaw.exe" skills list-builtin
)
if "%skill_choice%"=="3" (
    echo.
    set /p skill_name="Enter skill name to install: "
    "%USERPROFILE%\.local\bin\picoclaw.exe" skills install !skill_name!
)
if "%skill_choice%"=="4" (
    echo.
    echo Installing builtin skills...
    "%USERPROFILE%\.local\bin\picoclaw.exe" skills install-builtin
)
if "%skill_choice%"=="5" (
    echo.
    set /p skill_name="Enter skill name to remove: "
    "%USERPROFILE%\.local\bin\picoclaw.exe" skills remove !skill_name!
)
if "%skill_choice%"=="6" (
    echo.
    set /p keyword="Enter keyword to search: "
    "%USERPROFILE%\.local\bin\picoclaw.exe" skills search !keyword!
)
pause
goto main_menu

:auth_management
echo.
echo 🔐 Authentication Menu
echo ====================
echo 1. Login
echo 2. Logout  
echo 3. Check status
echo 0. Back to main menu
echo.
set /p auth_choice="Enter your choice (0-3): "

if "%auth_choice%"=="1" (
    echo.
    echo Logging in...
    "%USERPROFILE%\.local\bin\picoclaw.exe" auth login
)
if "%auth_choice%"=="2" (
    echo.
    echo Logging out...
    "%USERPROFILE%\.local\bin\picoclaw.exe" auth logout
)
if "%auth_choice%"=="3" (
    echo.
    echo Authentication status:
    "%USERPROFILE%\.local\bin\picoclaw.exe" auth status
)
pause
goto main_menu

:install_picoclaw
echo.
echo 📦 Installing PicoClaw...
echo.
call install-windows.bat
pause
goto main_menu

:open_browser
echo.
echo 🌐 Opening web browser...
echo.
start http://localhost:8080
goto main_menu

:advanced_options
echo.
echo 🛠️ Advanced Options
echo ==================
echo 1. Rebuild PicoClaw
echo 2. Update dependencies
echo 3. Run tests
echo 4. Clean build files
echo 5. View logs
echo 6. Configuration file location
echo 0. Back to main menu
echo.
set /p adv_choice="Enter your choice (0-6): "

if "%adv_choice%"=="1" (
    echo.
    echo Rebuilding PicoClaw...
    call build.bat clean
    call build.bat build
    call build.bat install
)
if "%adv_choice%"=="2" (
    echo.
    echo Updating dependencies...
    call build.bat update-deps
)
if "%adv_choice%"=="3" (
    echo.
    echo Running tests...
    call build.bat test
)
if "%adv_choice%"=="4" (
    echo.
    echo Cleaning build files...
    call build.bat clean
)
if "%adv_choice%"=="5" (
    echo.
    echo Configuration directory: %USERPROFILE%\.picoclaw\
    echo Log files location: %USERPROFILE%\.picoclaw\logs\
    echo Config file: %USERPROFILE%\.picoclaw\config.json
    echo.
    echo Would you like to open the config directory? (y/n)
    set /p open_config="> "
    if /i "!open_config!"=="y" explorer "%USERPROFILE%\.picoclaw"
)
if "%adv_choice%"=="6" (
    echo.
    echo Configuration File Location:
    echo %USERPROFILE%\.picoclaw\config.json
    echo.
    echo Workspace Directory:
    echo %USERPROFILE%\.picoclaw\workspace\
    echo.
    echo Installation Directory:
    echo %USERPROFILE%\.local\bin\
)
pause
goto main_menu

:install_picoclaw_silent
echo Installing PicoClaw silently...
if exist "install-windows.bat" (
    call install-windows.bat >nul 2>&1
) else (
    echo Error: install-windows.bat not found
    exit /b 1
)
goto :eof

:invalid_choice
echo.
echo ❌ Invalid choice. Please try again.
pause
goto main_menu

:exit
echo.
echo 👋 Thank you for using PicoClaw!
echo.
timeout /t 2 >nul
exit /b 0

endlocal