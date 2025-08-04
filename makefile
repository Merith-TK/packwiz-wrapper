default: gui

# Build with GUI (full version)
gui: check-upx
	@echo "Building GUI application..."
	go build -tags gui -ldflags="-s -w" -o pw.exe ./cmd/pw
	upx --best --lzma pw.exe

# Build without GUI (headless/embedded version)
headless:
	@echo "Building headless application..."
	go build -ldflags="-s -w" -o pw-headless.exe ./cmd/pw

# Build both versions
all: gui headless

clean:
	del /Q pw.exe pw-headless.exe 2>nul || true

install:
	git pull
	-go install -tags gui ./cmd/pw

# Build without compression (for development/debugging)
cli-debug:
	go build -o pw.exe ./cmd/pw

# Build headless without compression (for development/debugging)  
headless-debug:
	go build -o pw-headless.exe ./cmd/pw

# Check if UPX is available
check-upx:
	@upx --version >nul 2>&1 || (echo "UPX not found. Install from https://upx.github.io/" && exit 1)

.PHONY: default cli gui clean install cli-debug gui-debug check-upx
