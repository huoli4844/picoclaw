@echo off
setlocal enabledelayedexpansion
REM PicoClaw Build Script for Windows
REM This replaces the Makefile for Windows systems

REM Configuration
set BINARY_NAME=picoclaw
set BUILD_DIR=build
set CMD_DIR=cmd\%BINARY_NAME%
set MAIN_GO=%CMD_DIR%\main.go

REM Default installation prefix for Windows
set INSTALL_PREFIX=%USERPROFILE%\.local
set INSTALL_BIN_DIR=%INSTALL_PREFIX%\bin
set INSTALL_MAN_DIR=%INSTALL_PREFIX%\share\man\man1

REM PicoClaw home directory for Windows
set PICOCLAW_HOME=%USERPROFILE%\.picoclaw
set WORKSPACE_DIR=%PICOCLAW_HOME%\workspace
set WORKSPACE_SKILLS_DIR=%WORKSPACE_DIR%\skills
set BUILTIN_SKILLS_DIR=%CD%\skills

REM Detect Windows version
set PLATFORM=windows
set ARCH=x64

REM Check for Go installation
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    exit /b 1
)

REM Parse command line arguments
if "%1"=="" goto help
if "%1"=="all" goto build
if "%1"=="generate" goto generate
if "%1"=="build" goto build
if "%1"=="build-all" goto build-all
if "%1"=="install" goto install
if "%1"=="uninstall" goto uninstall
if "%1"=="uninstall-all" goto uninstall-all
if "%1"=="clean" goto clean
if "%1"=="vet" goto vet
if "%1"=="test" goto test
if "%1"=="fmt" goto fmt
if "%1"=="lint" goto lint
if "%1"=="deps" goto deps
if "%1"=="update-deps" goto update-deps
if "%1"=="check" goto check
if "%1"=="run" goto run
if "%1"=="help" goto help
goto unknown

:generate
echo Running generate...
if exist "%CMD_DIR%\workspace" (
    rmdir /s /q "%CMD_DIR%\workspace"
)
REM Windows compatible copy command for go:generate
echo Copying workspace files...
if exist "workspace" (
    xcopy /E /I /Y "workspace" "%CMD_DIR%\workspace" >nul 2>&1
) else (
    echo Warning: workspace directory not found in project root
)
go generate ./...
echo Generate complete
goto end

:build
call :generate
echo Building %BINARY_NAME% for %PLATFORM%/%ARCH%...
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"
set BUILD_TIME=%date:~0,4%-%date:~5,2%-%date:~8,2%T%time:~0,2%:%time:~3,2%:%time:~6,2%
set LDFLAGS=-ldflags "-X main.version=dev -X main.gitCommit=dev -X main.buildTime=%BUILD_TIME% -X main.goVersion=dev -s -w"
go build -v -tags stdjson %LDFLAGS% -o "%BUILD_DIR%\%BINARY_NAME%-windows-%ARCH%.exe" "./%CMD_DIR%"
echo Build complete: %BUILD_DIR%\%BINARY_NAME%-windows-%ARCH%.exe
copy /Y "%BUILD_DIR%\%BINARY_NAME%-windows-%ARCH%.exe" "%BUILD_DIR%\%BINARY_NAME%.exe" >nul
goto end

:build-all
echo Building for multiple platforms...
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"
set LDFLAGS=-ldflags "-X main.version=dev -X main.gitCommit=dev -X main.buildTime=dev -X main.goVersion=dev -s -w"

REM Linux AMD64
set GOOS=linux
set GOARCH=amd64
go build %LDFLAGS% -o "%BUILD_DIR%\%BINARY_NAME%-linux-amd64" "./%CMD_DIR%"

REM Linux ARM64
set GOOS=linux
set GOARCH=arm64
go build %LDFLAGS% -o "%BUILD_DIR%\%BINARY_NAME%-linux-arm64" "./%CMD_DIR%"

REM Darwin ARM64
set GOOS=darwin
set GOARCH=arm64
go build %LDFLAGS% -o "%BUILD_DIR%\%BINARY_NAME%-darwin-arm64" "./%CMD_DIR%"

REM Windows AMD64
set GOOS=windows
set GOARCH=amd64
go build %LDFLAGS% -o "%BUILD_DIR%\%BINARY_NAME%-windows-amd64.exe" "./%CMD_DIR%"

REM Reset environment variables
set GOOS=
set GOARCH=

echo All builds complete
goto end

:install
call :build
echo Installing %BINARY_NAME%...
if not exist "%INSTALL_BIN_DIR%" mkdir "%INSTALL_BIN_DIR%"
copy /Y "%BUILD_DIR%\%BINARY_NAME%.exe" "%INSTALL_BIN_DIR%\%BINARY_NAME%.exe"
if %errorlevel% neq 0 (
    echo Error: Failed to copy binary to %INSTALL_BIN_DIR%
    echo Make sure you have write permissions to %INSTALL_PREFIX%
    pause
    exit /b 1
)
echo Installed binary to %INSTALL_BIN_DIR%\%BINARY_NAME%.exe
echo.
echo Don't forget to add %INSTALL_BIN_DIR% to your PATH environment variable!
echo Installation complete!
goto end

:uninstall
echo Uninstalling %BINARY_NAME%...
if exist "%INSTALL_BIN_DIR%\%BINARY_NAME%.exe" (
    del /f "%INSTALL_BIN_DIR%\%BINARY_NAME%.exe"
    echo Removed binary from %INSTALL_BIN_DIR%\%BINARY_NAME%.exe
) else (
    echo Binary not found in %INSTALL_BIN_DIR%
)
echo.
echo Note: Only the executable file has been deleted.
echo If you need to delete all configurations ^(config.json, workspace, etc.^), run 'build.bat uninstall-all'
goto end

:uninstall-all
echo Removing workspace and skills...
if exist "%PICOCLAW_HOME%" (
    rmdir /s /q "%PICOCLAW_HOME%"
    echo Removed workspace: %PICOCLAW_HOME%
) else (
    echo Workspace directory not found
)
echo Complete uninstallation done!
goto end

:clean
echo Cleaning build artifacts...
if exist "%BUILD_DIR%" (
    rmdir /s /q "%BUILD_DIR%"
    echo Build directory cleaned
) else (
    echo No build directory to clean
)
echo Clean complete
goto end

:vet
echo Running go vet...
go vet ./...
goto end

:test
echo Running tests...
go test ./...
goto end

:fmt
echo Formatting code...
golangci-lint fmt
goto end

:lint
echo Running linters...
golangci-lint run
goto end

:deps
echo Downloading dependencies...
go mod download
go mod verify
goto end

:update-deps
echo Updating dependencies...
go get -u ./...
go mod tidy
goto end

:check
echo Running full check...
call :deps
call :fmt
call :vet
call :test
echo Check complete
goto end

:run
call :build
echo Running %BINARY_NAME%...
if exist "%BUILD_DIR%\%BINARY_NAME%.exe" (
    "%BUILD_DIR%\%BINARY_NAME%.exe" %2 %3 %4 %5 %6 %7 %8 %9
) else (
    echo Binary not found. Please run build first.
)
goto end

:help
echo.
echo PicoClaw Build Script for Windows
echo.
echo Usage:
echo   build.bat [command]
echo.
echo Commands:
echo   all           - Build the picoclaw binary ^(default^)
echo   generate      - Run go generate
echo   build         - Build picoclaw for Windows
echo   build-all     - Build picoclaw for all platforms
echo   install       - Install picoclaw to system
echo   uninstall     - Remove picoclaw executable from system
echo   uninstall-all - Remove picoclaw and all data
echo   clean         - Remove build artifacts
echo   vet           - Run go vet for static analysis
echo   test          - Test Go code
echo   fmt           - Format Go code
echo   lint          - Run linters
echo   deps          - Download dependencies
echo   update-deps   - Update dependencies
echo   check         - Run vet, fmt, and verify dependencies
echo   run           - Build and run picoclaw
echo   help          - Show this help message
echo.
echo Examples:
echo   build.bat build              # Build for current platform
echo   build.bat install            # Install to %%USERPROFILE%%\.local\bin
echo   build.bat uninstall          # Remove executable
echo   build.bat clean              # Clean build files
echo.
echo Environment Variables:
echo   INSTALL_PREFIX          # Installation prefix ^(default: %%USERPROFILE%%\.local^)
echo   WORKSPACE_DIR           # Workspace directory ^(default: %%USERPROFILE%%\.picoclaw\workspace^)
echo.
echo Current Configuration:
echo   Platform: %PLATFORM%/%ARCH%
echo   Binary: %BUILD_DIR%\%BINARY_NAME%-%PLATFORM%-%ARCH%.exe
echo   Install Prefix: %INSTALL_PREFIX%
echo   Workspace: %WORKSPACE_DIR%
echo.
goto end

:unknown
echo Unknown command: %1
echo Use 'build.bat help' to see available commands.
goto end

:end
endlocal