.PHONY: build run fmt test tidy

build:
	go build ./...

run:
	go run ./apps/bff

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

test:
	go test ./...

tidy:
	go mod tidy
