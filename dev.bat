@echo off
REM Quick Development Script for PicoClaw on Windows
REM This script handles common development tasks

setlocal enabledelayedexpansion

echo PicoClaw Quick Development Script
echo ==================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

REM Quick development menu
if "%1"=="test" goto quick-test
if "%1"=="build" goto quick-build
if "%1"=="install" goto quick-install
if "%1"=="run" goto quick-run
if "%1"=="check" goto quick-check
if "%1"=="clean" goto quick-clean
goto show-options

:show-options
echo Quick Development Options:
echo.
echo   dev.bat test     - Run quick test
echo   dev.bat build    - Quick build
echo   dev.bat install  - Quick install
echo   dev.bat run      - Quick build and run
echo   dev.bat check    - Quick code quality check
echo   dev.bat clean    - Quick clean
echo.
echo Or run without arguments for this help message.
echo.
goto end

:quick-test
echo Running quick tests...
call build.bat test
goto end

:quick-build
echo Quick building...
call build.bat build
goto end

:quick-install
echo Quick installing...
call build.bat install
echo.
echo Remember to add %USERPROFILE%\.local\bin to your PATH!
goto end

:quick-run
echo Quick building and running...
call build.bat run
goto end

:quick-check
echo Quick code quality check...
echo 1. Formatting code...
call build.bat fmt
echo.
echo 2. Running static analysis...
call build.bat vet
echo.
echo 3. Running tests...
call build.bat test
echo.
echo Quick check complete!
goto end

:quick-clean
echo Quick clean...
call build.bat clean
goto end

:end
echo.
pause