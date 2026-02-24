@echo off
REM Fix Windows Build Issues for PicoClaw
REM This script fixes common Windows build problems

echo PicoClaw Windows Build Fix Script
echo ==================================
echo.

REM Check Go installation
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

echo Step 1: Checking and creating necessary directories...
if not exist "cmd\picoclaw" (
    echo Error: cmd\picoclaw directory not found
    echo Please run this script from the picoclaw project root
    pause
    exit /b 1
)

echo Step 2: Copying workspace files (Windows compatible)...
if exist "workspace" (
    if exist "cmd\picoclaw\workspace" (
        rmdir /s /q "cmd\picoclaw\workspace"
    )
    xcopy /E /I /Y "workspace" "cmd\picoclaw\workspace" >nul 2>&1
    echo Workspace files copied successfully
) else (
    echo Warning: workspace directory not found
    echo Creating basic workspace...
    
    if not exist "workspace" mkdir "workspace"
    if not exist "workspace\memory" mkdir "workspace\memory"
    if not exist "workspace\skills" mkdir "workspace\skills"
    
    REM Create basic workspace files
    echo # User Information > workspace\USER.md
    echo This file contains your preferences and context information. >> workspace\USER.md
    echo. >> workspace\USER.md
    echo ## Preferences >> workspace\USER.md
    echo - Response style:  >> workspace\USER.md
    echo - Topics of interest: >> workspace\USER.md
    echo - Communication preferences: >> workspace\USER.md
    echo. >> workspace\USER.md
    echo ## Context >> workspace\USER.md
    echo Add any context information you want the AI to remember about you. >> workspace\USER.md
    
    echo # Agent Information > workspace\AGENT.md
    echo This file defines the AI agent's characteristics. >> workspace\AGENT.md
    echo. >> workspace\AGENT.md
    echo ## Agent Name >> workspace\AGENT.md
    echo PicoClaw >> workspace\AGENT.md
    echo. >> workspace\AGENT.md
    echo ## Personality >> workspace\AGENT.md
    echo - Helpful and efficient >> workspace\AGENT.md
    echo - Friendly and professional >> workspace\AGENT.md
    echo - Focused on accuracy >> workspace\AGENT.md
    echo. >> workspace\AGENT.md
    echo ## Capabilities >> workspace\AGENT.md
    echo - Natural language understanding >> workspace\AGENT.md
    echo - Code generation >> workspace\AGENT.md
    echo - Problem solving >> workspace\AGENT.md
    echo - Creative assistance >> workspace\AGENT.md
    
    echo ✅ Basic workspace created
    
    REM Now copy the created workspace
    xcopy /E /I /Y "workspace" "cmd\picoclaw\workspace" >nul 2>&1
    echo Workspace files copied successfully
)

echo Step 3: Downloading Go dependencies...
go mod download
go mod verify

echo Step 4: Building with verbose output...
echo.
echo Build output:
echo ============
go build -v -tags stdjson -o "build\picoclaw.exe" "./cmd/picoclaw"
if %errorlevel% equ 0 (
    echo.
    echo ✓ Build successful!
    echo Output: build\picoclaw.exe
) else (
    echo.
    echo ✗ Build failed with error code %errorlevel%
    echo.
    echo Common solutions:
    echo 1. Make sure Go is properly installed
    echo 2. Run as administrator if needed
    echo 3. Check if Antivirus is blocking the build
    echo 4. Verify all source files are present
)

echo.
echo Step 5: Testing installation...
if exist "build\picoclaw.exe" (
    echo Testing the built binary...
    "build\picoclaw.exe" --help >nul 2>&1
    if %errorlevel% equ 0 (
        echo ✓ Binary test successful
    ) else (
        echo ⚠ Binary test failed (may be normal if --help not supported)
    )
)

echo.
echo Windows build fix complete!
echo.
echo If build succeeded, you can now run:
echo   build.bat install
echo.
pause