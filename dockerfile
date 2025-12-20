FROM mirror.gcr.io/library/golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o seckill-server cmd/api/main.go

FROM mirror.gcr.io/library/alpine:latest
WORKDIR /root/

COPY --from=builder /app/seckill-server .
COPY --from=builder /app/internal/config /root/internal/config

EXPOSE 8080
CMD ["./seckill-server"]


# #dockerfile
# FROM golang:1.24-alpine AS builder
# # ↑ 第一阶段：使用 Go 1.24 镜像作为构建环境
# # AS builder 给这个阶段起名叫 builder

# WORKDIR /app
# # ↑ 在容器内创建并切换到 /app 目录

# COPY . .
# # ↑ 把本地项目所有文件复制到容器的 /app 目录

# RUN go mod download
# # ↑ 下载 Go 依赖包（根据 go.mod）

# RUN go build -o seckill-server cmd/api/main.go
# # ↑ 编译 Go 程序，生成可执行文件 seckill-server

# FROM alpine:latest
# # ↑ 第二阶段：使用轻量级 Alpine 镜像（只有5MB）

# WORKDIR /root/
# # ↑ 在容器内切换到 /root/ 目录

# COPY --from=builder /app/seckill-server .
# # ↑ 从第一阶段（builder）复制编译好的程序到当前目录
# # 这样最终镜像只包含可执行文件，不包含 Go 编译器（减小镜像大小）

# COPY --from=builder /app/internal/config /root/internal/config
# # ↑ 复制配置文件目录

# EXPOSE 8080
# # ↑ 声明容器监听 8080 端口（只是文档说明，不实际开放）

# CMD ["./seckill-server"]
# # ↑ 容器启动时执行的命令：运行编译好的程序