.PHONY: build run run-user run-product run-order fmt test tidy proto-generate proto-lint proto-breaking

build: proto-generate
	go build ./...

run:
	go run ./apps/bff

run-user:
	go run ./apps/services/user

run-product:
	go run ./apps/services/product

run-order:
	go run ./apps/services/order

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

test: proto-generate
	go test ./...

tidy:
	go mod tidy

proto-generate:
	buf generate

proto-lint:
	buf lint

proto-breaking:
	buf breaking --against '.git#branch=main'
