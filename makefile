install:
	git pull
	-go install ./cmd/pw
	-go install ./cmd/pw-modlist
	-go install ./cmd/pw-reinstall
