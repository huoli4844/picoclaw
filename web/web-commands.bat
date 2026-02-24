@echo off
REM Quick Web Commands for Windows Development

echo Web Frontend Commands
echo ====================

:menu
echo.
echo 1. Install dependencies
echo 2. Start development server
echo 3. Build for production
echo 4. Clean build files
echo 5. Exit
echo.
set /p choice="Select an option (1-5): "

if "%choice%"=="1" goto install
if "%choice%"=="2" goto dev
if "%choice%"=="3" goto build
if "%choice%"=="4" goto clean
if "%choice%"=="5" goto exit
echo Invalid choice. Please try again.
goto menu

:install
echo.
echo Installing dependencies...
call npm install
echo.
pause
goto menu

:dev
echo.
echo Starting development server...
call npm run dev
echo.
pause
goto menu

:build
echo.
echo Building for production...
call npm run build
echo.
pause
goto menu

:clean
echo.
echo Cleaning build files...
if exist "dist" (
    rmdir /s /q "dist"
    echo Build files cleaned successfully.
) else (
    echo No dist directory found.
)
echo.
pause
goto menu

:exit
echo Goodbye!
exit /b 0