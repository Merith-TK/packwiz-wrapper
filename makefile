default: cli gui

cli:
	go build -o pw.exe ./cmd/pw

gui:
	go build -o pw-gui.exe ./cmd/pw-gui

clean:
	del /Q pw.exe pw-gui.exe 2>nul || true

install:
	git pull
	-go install ./cmd/pw
	-go install ./cmd/pw-gui

.PHONY: default cli gui clean install
