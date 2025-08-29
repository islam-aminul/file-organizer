@echo off
echo Building ZenSort with CGO support for GUI...
set CGO_ENABLED=1
go build -o zensort.exe main.go

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
    echo Build complete! zensort.exe supports both GUI and CLI modes.
)
