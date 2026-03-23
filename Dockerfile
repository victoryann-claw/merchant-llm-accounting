FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装编译依赖
RUN apk add --no-cache git

# 复制源码
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o merchant-api ./cmd/server

# 最终镜像
FROM alpine:3.19

WORKDIR /app

# 安装CA证书（用于HTTPS调用）
RUN apk --no-cache add ca-certificates tzdata

# 复制可执行文件
COPY --from=builder /app/merchant-api .
COPY --from=builder /app/.env .env 2>/dev/null || true

# 创建日志目录
RUN mkdir -p /app/logs

# 时区
ENV TZ=Asia/Shanghai

EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动
CMD ["./merchant-api"]
