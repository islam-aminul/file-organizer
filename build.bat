@echo off
echo Building ZenSort with GUI support...

REM Compile resource file if windres is available
where windres >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo Compiling icon resources...
    windres -i icon.rc -o icon.syso
    if %ERRORLEVEL% NEQ 0 (
        echo Warning: Failed to compile icon resources. Continuing without icon.
    ) else (
        echo Icon resources compiled successfully.
    )
) else (
    echo Warning: windres not found. Building without icon.
)

set CGO_ENABLED=1
go build -ldflags="-H windowsgui" -o zensort.exe main.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo CGO build failed. Building CLI-only version...
    set CGO_ENABLED=0
    go build -tags nocgo -o zensort-cli-only.exe main.go
    echo.
    echo CLI-only build complete! Use zensort-cli-only.exe for command-line interface.
    echo To use GUI, install a C compiler and run: set CGO_ENABLED=1 ^&^& go build -o zensort.exe main.go
) else (
    echo.
    echo Build complete! Run zensort.exe to start the application.
)
pause
