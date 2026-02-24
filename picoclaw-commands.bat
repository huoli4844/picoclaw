@echo off
REM Interactive PicoClaw Commands for Windows
REM User-friendly menu interface

setlocal enabledelayedexpansion

:menu
cls
echo.
echo ==========================================
echo           PicoClaw Build Menu
echo ==========================================
echo.
echo  1. Build for Windows
echo  2. Build for All Platforms  
echo  3. Install to System
echo  4. Uninstall from System
echo  5. Complete Uninstall (remove all data)
echo  6. Clean Build Files
echo  7. Run Tests
echo  8. Check Dependencies
echo  9. Format Code
echo 10. Run Linting
echo 11. Build and Run
echo 12. Full Check (deps + fmt + vet + test)
echo.
echo  0. Exit
echo.
echo ==========================================
set /p choice="Select an option (0-12): "

if "%choice%"=="1" goto build
if "%choice%"=="2" goto build-all
if "%choice%"=="3" goto install
if "%choice%"=="4" goto uninstall
if "%choice%"=="5" goto uninstall-all
if "%choice%"=="6" goto clean
if "%choice%"=="7" goto test
if "%choice%"=="8" goto deps
if "%choice%"=="9" goto fmt
if "%choice%"=="10" goto lint
if "%choice%"=="11" goto run
if "%choice%"=="12" goto check
if "%choice%"=="0" goto exit
echo Invalid choice. Please try again.
pause
goto menu

:build
echo.
echo Building PicoClaw for Windows...
call build.bat build
echo.
pause
goto menu

:build-all
echo.
echo Building PicoClaw for All Platforms...
call build.bat build-all
echo.
pause
goto menu

:install
echo.
echo Installing PicoClaw to System...
call build.bat install
echo.
pause
goto menu

:uninstall
echo.
echo Uninstalling PicoClaw from System...
call build.bat uninstall
echo.
pause
goto menu

:uninstall-all
echo.
echo WARNING: This will remove ALL PicoClaw data including configurations!
echo.
set /p confirm="Are you sure you want to continue? (y/N): "
if /i "%confirm%"=="y" (
    call build.bat uninstall-all
) else (
    echo Operation cancelled.
)
echo.
pause
goto menu

:clean
echo.
echo Cleaning Build Files...
call build.bat clean
echo.
pause
goto menu

:test
echo.
echo Running Tests...
call build.bat test
echo.
pause
goto menu

:deps
echo.
echo Checking Dependencies...
call build.bat deps
echo.
pause
goto menu

:fmt
echo.
echo Formatting Code...
call build.bat fmt
echo.
pause
goto menu

:lint
echo.
echo Running Linting...
call build.bat lint
echo.
pause
goto menu

:run
echo.
echo Building and Running PicoClaw...
call build.bat run
echo.
pause
goto menu

:check
echo.
echo Running Full Check...
call build.bat check
echo.
pause
goto menu

:exit
echo.
echo Goodbye!
exit /b 0

endlocal