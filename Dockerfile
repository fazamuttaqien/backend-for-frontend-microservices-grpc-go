FROM golang:1.25-alpine AS builder
ARG APP
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN test -n "$APP" && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./apps/$APP
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/healthcheck ./cmd/healthcheck

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder /out/app /app/app
COPY --from=builder /out/healthcheck /app/healthcheck
USER app
ENTRYPOINT ["/app/app"]
