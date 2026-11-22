# ===============================
# Stage 1: Builder
# ===============================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -mod=readonly -o /app/main .

# ===============================
# Stage 2: Production Runtime
# ===============================
FROM alpine:3.20

RUN apk --no-cache add ca-certificates wget \
	&& adduser -D -u 65532 -g "" appuser

WORKDIR /app

COPY --from=builder /app/main .

USER 65532

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD wget -qO- http://127.0.0.1:8000/health || exit 1

CMD ["./main"]