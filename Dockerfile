# 1. ビルド用ステージ
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 依存関係のキャッシュを利用するために先にコピー
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピーしてビルド
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o dc-bot main.go

# 2. 実行用ステージ (Alpine Linux を使って軽量化)
FROM alpine:latest

# DiscordBotのHTTPS通信に必要なルート証明書と、タイムゾーン設定、ffmpegをインストール
RUN apk --no-cache add ca-certificates tzdata ffmpeg && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    echo "Asia/Tokyo" > /etc/localtime

WORKDIR /app

# ビルドしたバイナリをコピー
COPY --from=builder /app/dc-bot .

# コンテナ起動時に実行するコマンド
CMD ["./dc-bot"]
