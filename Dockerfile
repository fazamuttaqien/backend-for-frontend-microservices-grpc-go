# Build all Go services from one multi-stage image and select the binary with APP.
FROM golang:1.24-alpine AS builder

ARG APP
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN test -n "$APP" && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./apps/$APP

FROM alpine:3.22

RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder /out/app /app/app

USER app
ENTRYPOINT ["/app/app"]
