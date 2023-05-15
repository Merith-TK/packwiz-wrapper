default:
	go build ./cmd/pw


install:
	git pull
	-go install ./cmd/pw
