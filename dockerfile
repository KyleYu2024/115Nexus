# 使用基础镜像，不指定平台，让 buildx 自动处理
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
# 禁用 CGO 以确保高度兼容性
RUN CGO_ENABLED=0 go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/main .
RUN chmod +x /app/main
RUN mkdir -p config
EXPOSE 7833
CMD ["./main"]
