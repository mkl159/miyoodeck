# ── Stage 1: Build Go server for ARM Cortex-A7 ──────────────────────────────
FROM golang:1.21-alpine AS go-builder

WORKDIR /build/server
COPY server/go.mod server/go.sum* ./
RUN go mod download

COPY server/ .
RUN GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o webdeck .

# ── Stage 2: Export binary ───────────────────────────────────────────────────
FROM scratch AS export
COPY --from=go-builder /build/server/webdeck /webdeck
