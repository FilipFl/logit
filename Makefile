test:
	go test ./...

build:
	go build -o bin/logit

run:
	go run main.go

install: build
	sudo cp bin/logit /usr/local/bin/logit