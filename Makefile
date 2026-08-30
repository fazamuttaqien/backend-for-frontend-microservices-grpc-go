.PHONY: build run fmt test tidy proto-generate proto-lint proto-breaking

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

proto-generate:
	buf generate

proto-lint:
	buf lint

proto-breaking:
	buf breaking --against '.git#branch=main'
