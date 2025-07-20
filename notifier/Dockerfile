FROM golang:1.23-alpine AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -o notifier cmd/main.go

FROM alpine:3.18

RUN apk add --no-cache ca-certificates su-exec

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app \
 && mkdir -p /app/logs && chown -R app:app /app/logs

COPY --from=builder /workspace/notifier .

ENTRYPOINT ["sh", "-c", "chown -R app:app /app/logs && exec su-exec app ./notifier"]
