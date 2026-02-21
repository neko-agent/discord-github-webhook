FROM golang:1.25.1 AS builder

WORKDIR /workspace

# 複製 local dependency
COPY go-packages/logger/ ./go-packages/logger/
COPY apps/go-github-discord-bridge/ ./apps/go-github-discord-bridge/

WORKDIR /workspace/apps/go-github-discord-bridge

# 下載依賴
RUN go mod download

# 編譯執行檔
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# 使用更小的 base image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /workspace/apps/go-github-discord-bridge/main .

CMD ["./main"]
