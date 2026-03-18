# ============================
# 第一阶段：构建 (Builder)
# ============================
FROM --platform=$BUILDPLATFORM golang:alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装 git
RUN apk add --no-cache git

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod tidy

# 复制源码
COPY . .

# 接收 Docker buildx 传入的平台参数
ARG TARGETOS
ARG TARGETARCH

# 编译阶段：显式指定目标操作系统和架构（默认 linux/amd64）
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o main .

# ============================
# 第二阶段：运行 (Runner)
# ============================
FROM --platform=linux/amd64 alpine:latest

# 安装基础证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 创建配置目录
RUN mkdir -p config

EXPOSE 7833

# 启动
CMD ["./main"]
