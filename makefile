install:
	git pull
	go install ./cmd/pw

test:
	-mkdir build/
	
	go run ./cmd/pw-modlist -d build -o build/modlist.txt