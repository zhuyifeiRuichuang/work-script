# Build stage
FROM golang:1.22-alpine AS builder

# Install build tools
RUN apk add --no-cache git make

# Set working directory
WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the operator
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager ./cmd/main.go

# Run stage
FROM gcr.io/distroless/static:nonroot

# Set user and group
USER nonroot:nonroot

# Copy the binary from builder
COPY --from=builder /workspace/manager /manager

# Set entrypoint
ENTRYPOINT ["/manager"]
