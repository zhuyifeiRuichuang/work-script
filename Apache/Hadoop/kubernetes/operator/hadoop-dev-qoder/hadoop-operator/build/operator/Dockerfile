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

# =============================================================================
# Hadoop Operator 镜像构建文件
# =============================================================================
# 用途: 构建 Hadoop Operator 控制器镜像
# 使用对象: 运维人员部署 Operator、开发者发布新版本
# 构建产物: hadoop-operator 二进制文件的容器镜像
# =============================================================================

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
COPY go.mod go.mod
COPY go.sum go.sum

# 下载依赖（在源码复制之前，利用缓存）
RUN go mod download && go mod verify

# 复制源码
COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -trimpath \
    -ldflags="-w -s \
        -X main.Version=${VERSION} \
        -X main.BuildDate=${BUILD_DATE} \
        -X main.GitCommit=${GIT_COMMIT} \
        -X main.GitBranch=${GIT_BRANCH}" \
    -o manager cmd/main.go

# 验证构建结果
RUN chmod +x /workspace/manager

# ============================================
# 阶段 2: 运行时阶段 (Runtime Stage)
# ============================================
FROM gcr.io/distroless/static:nonroot

# 从 builder 阶段复制 CA 证书和时区数据
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 设置工作目录
WORKDIR /

# 复制构建好的二进制文件
COPY --from=builder /workspace/manager .

# 使用非 root 用户运行
USER 65532:65532

# 暴露指标端口
EXPOSE 8080
EXPOSE 8443

# 设置入口点
ENTRYPOINT ["/manager"]
