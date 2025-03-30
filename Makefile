test:
	go test .

build:
	go build -o bin/logit main.go

run:
	go run main.go