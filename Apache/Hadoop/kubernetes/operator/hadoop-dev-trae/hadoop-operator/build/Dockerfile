# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

# Build stage
FROM golang:1.20 as builder

WORKDIR /app

# Copy go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -installsuffix cgo \
    -ldflags "-w -s" \
    -o hadoopoperator ./cmd/manager

# Final stage
FROM alpine:3.17

# Add ca-certificates
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN adduser -D -u 1001 hadoop

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/hadoopoperator .

# Set permissions
RUN chown hadoop:hadoop /app/hadoopoperator
RUN chmod +x /app/hadoopoperator

# Expose ports
EXPOSE 8081 9443

# Set user
USER hadoop

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8081/healthz || exit 1

# Entrypoint
ENTRYPOINT ["./hadoopoperator"]

# Labels
LABEL org.opencontainers.image.title="Hadoop Operator"
LABEL org.opencontainers.image.description="Kubernetes Operator for Hadoop Cluster"
LABEL org.opencontainers.image.version="v1.0.0"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.source="https://github.com/apache/hadoop-operator"
