default: check-upx
	@echo "Building CLI application..."
	go build -ldflags="-s -w" -o pw.exe ./cmd/pw
	upx --best --lzma pw.exe

clean:
	del /Q pw.exe 2>nul || true

install:
	git pull
	-go install ./cmd/pw

# Build without compression (for development/debugging)
cli-debug:
	go build -o pw.exe ./cmd/pw

# Check if UPX is available
check-upx:
	@upx --version >nul 2>&1 || (echo "UPX not found. Install from https://upx.github.io/" && exit 1)

.PHONY: default cli gui clean install cli-debug gui-debug check-upx
