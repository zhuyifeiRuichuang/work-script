# Copyright 2024 Apache Software Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Hadoop Operator Dockerfile
# 采用多阶段构建策略，实现最小化运行时镜像
# 支持多平台构建: linux/amd64, linux/arm64, linux/s390x, linux/ppc64le

# ============================================
# 阶段 1: 构建阶段 (Builder Stage)
# ============================================
FROM golang:1.21-alpine AS builder

# 构建参数
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG BUILD_DATE
ARG VERSION
ARG GIT_COMMIT
ARG GIT_BRANCH

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /workspace

# 先复制 Go Modules 清单，利用 Docker 缓存层
# 这样源码变更不会导致依赖重新下载
COPY go.mod go.mod
COPY go.sum go.sum

# 下载依赖（在源码复制之前，利用缓存）
RUN go mod download && go mod verify

# 复制源码
COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/

# 构建二进制文件
# - CGO_ENABLED=0: 禁用 CGO，生成静态链接二进制
# - -a: 强制重新构建所有包
# - -ldflags: 注入版本信息，优化二进制大小
# - -trimpath: 移除构建路径信息，提高可复现性
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -trimpath \
    -ldflags="-w -s \
        -X main.Version=${VERSION} \
        -X main.BuildDate=${BUILD_DATE} \
        -X main.GitCommit=${GIT_COMMIT} \
        -X main.GitBranch=${GIT_BRANCH}" \
    -o manager cmd/main.go

# 验证构建结果
RUN chmod +x /workspace/manager && \
    /workspace/manager --version 2>/dev/null || true

# ============================================
# 阶段 2: 安全扫描阶段 (Security Scan Stage)
# ============================================
FROM aquasec/trivy:latest AS security-scanner
COPY --from=builder /workspace/manager /tmp/manager
# 可选: 运行安全扫描（构建时可跳过）
# RUN trivy filesystem --exit-code 0 --no-progress /tmp/manager

# ============================================
# 阶段 3: 运行时阶段 (Runtime Stage)
# ============================================
# 使用 distroless 作为最小化基础镜像
# distroless 镜像特点:
# - 无 shell，减少攻击面
# - 无包管理器
# - 仅包含运行应用所需的最小依赖
FROM gcr.io/distroless/static:nonroot

# 从 builder 阶段复制 CA 证书
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 从 builder 阶段复制时区数据
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 设置工作目录
WORKDIR /

# 复制构建好的二进制文件
COPY --from=builder /workspace/manager .

# 使用非 root 用户运行 (65532:nonroot)
# distroless 镜像预定义了 nonroot 用户
USER 65532:65532

# 暴露健康检查端口（如果启用 metrics）
EXPOSE 8080
EXPOSE 8443

# 设置入口点
ENTRYPOINT ["/manager"]

# 健康检查（注意：distroless 无 shell，使用 exec 形式）
# 实际健康检查由 Kubernetes 通过 livenessProbe/readinessProbe 执行
# 这里仅作为文档说明
