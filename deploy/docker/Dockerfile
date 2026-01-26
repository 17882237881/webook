# ============================================
# 简化版 Dockerfile - 使用本地构建的二进制文件
# ============================================
# 使用前请先运行:
# $env:GOOS='linux'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'; go build -o webook-linux ./cmd/webook

FROM alpine:3.19

# 安装 CA 证书（用于 HTTPS 请求）和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 创建非 root 用户运行应用（安全最佳实践）
RUN adduser -D -g '' appuser

WORKDIR /app

# 复制本地构建的二进制文件
COPY webook-linux ./webook

# 设置可执行权限
RUN chmod +x ./webook

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./webook"]
