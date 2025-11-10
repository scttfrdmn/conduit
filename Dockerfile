# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /build/conduit \
    ./cmd/conduit

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/conduit /usr/local/bin/conduit

# Copy license and readme
COPY LICENSE README.md ./

# Create non-root user
RUN addgroup -g 1000 conduit && \
    adduser -D -u 1000 -G conduit conduit

USER conduit

ENTRYPOINT ["conduit"]
CMD ["--help"]
