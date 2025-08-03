default: cli gui

cli: check-upx
	@echo "Building CLI application..."
	go build -ldflags="-s -w" -o pw.exe ./cmd/pw
	upx --best --lzma pw.exe

gui: check-upx
	@echo "Building GUI application..."
	@echo "Warning: First time building GUI may take longer due to dependencies."
	go build -ldflags="-s -w" -o pw-gui.exe ./cmd/pw-gui
	upx --best --lzma pw-gui.exe

clean:
	del /Q pw.exe pw-gui.exe 2>nul || true

install:
	git pull
	-go install ./cmd/pw
	-go install ./cmd/pw-gui

# Build without compression (for development/debugging)
cli-debug:
	go build -o pw.exe ./cmd/pw

gui-debug:
	go build -o pw-gui.exe ./cmd/pw-gui

# Check if UPX is available
check-upx:
	@upx --version >nul 2>&1 || (echo "UPX not found. Install from https://upx.github.io/" && exit 1)

.PHONY: default cli gui clean install cli-debug gui-debug check-upx
