# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Fetch dependencies first (layer-cached as long as go.mod/go.sum don't change)
COPY go.mod go.sum ./
RUN go mod download

# Build the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o simorq_backend .

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/simorq_backend .
COPY casbin/ ./casbin/

ENV TZ=Asia/Tehran

EXPOSE 8080

ENTRYPOINT ["/app/simorq_backend", "--config", "/app/config.yaml"]
CMD ["http", "start"]
