# -------------------------
# Build stage
# -------------------------
FROM golang:1.24 AS builder

WORKDIR /app
COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd

# -------------------------
# Final stage: minimal runtime
# -------------------------
FROM alpine:latest

RUN apk --no-cache add ca-certificates libwebp-tools


WORKDIR /app
COPY --from=builder /app/app .
COPY src/configs ./configs

EXPOSE 8080

CMD ["./app"]
