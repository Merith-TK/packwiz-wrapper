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

# Quick development build (no compression)
dev: cli-debug headless-debug
	@echo "Development builds complete."

# Test target
test:
	@echo "Running tests..."
	go test -v ./...

# Test and build
test-build: test dev

# Show version information
version:
	@echo "Current version info:"
	@git describe --tags --always --dirty 2>nul || echo "No git tags found"
	@go version

clean:
	@echo "Cleaning build artifacts..."
	-del /Q pw.exe pw-headless.exe pw-gui.exe 2>nul
	-rmdir /S /Q dist 2>nul
	-del /Q *.upx 2>nul
	@echo "Clean complete."

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

# GoReleaser targets
release-test:
	@echo "Validating GoReleaser configuration..."
	goreleaser check

release-snapshot:
	@echo "Building snapshot release..."
	goreleaser build --snapshot --clean

release-dry-run:
	@echo "Testing release process..."
	goreleaser release --snapshot --skip-publish --clean

release:
	@echo "Creating release..."
	goreleaser release --clean

# Help target
help:
	@echo "Available targets:"
	@echo "  gui              - Build GUI version with UPX compression"
	@echo "  headless         - Build headless version with UPX compression"
	@echo "  all              - Build both versions"
	@echo "  dev              - Quick development builds (no compression)"
	@echo "  test             - Run tests"
	@echo "  test-build       - Run tests then build"
	@echo "  version          - Show version information"
	@echo "  cli-debug        - Build GUI version without compression"
	@echo "  headless-debug   - Build headless version without compression"
	@echo "  clean            - Remove built binaries and artifacts"
	@echo "  install          - Install GUI version with go install"
	@echo "  release-test     - Validate GoReleaser configuration"
	@echo "  release-snapshot - Build all platforms locally"
	@echo "  release-dry-run  - Test release without publishing"
	@echo "  release          - Create actual release (requires git tag)"
	@echo "  help             - Show this help message"

.PHONY: default gui headless all dev test test-build version clean install cli-debug headless-debug check-upx release-test release-snapshot release-dry-run release help
