# Build Stage
FROM golang:1.24-alpine AS build

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
# - CGO_ENABLED=0 for full static linking
# - ldflags "-s -w" to strip symbol table and debug info for size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gophership ./cmd/gophership

# Final Stage: Distroless Static
# Provides minimal C libraries and SSL certs without a shell or package manager
FROM gcr.io/distroless/static:nonroot

LABEL maintainer="sungp"
LABEL description="GopherShip: Biological Resilient Log Engine"

WORKDIR /

# Copy only the compiled binary
COPY --from=build /app/gophership /gophership

# Use nonroot user for security (NFR.Sec2)
USER 65532:65532

# Ingestion Port
EXPOSE 4317
# Metrics Port
EXPOSE 9091
# Control Port
EXPOSE 9092

ENTRYPOINT ["/gophership"]
