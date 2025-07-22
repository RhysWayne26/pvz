FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOBIN=/workspace/bin \
    CGO_ENABLED=0 GOOS=linux \
    go install github.com/pressly/goose/v3/cmd/goose@v3.16.0 && \
    go build -o pvz cmd/main.go

RUN GOBIN=/workspace/bin go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.39

FROM alpine:3.18

RUN apk add --no-cache ca-certificates su-exec

WORKDIR /app

RUN addgroup -S app \
 && adduser -S app -G app \
 && mkdir -p /app/logs /app/migrations \
 && chown -R app:app /app/logs /app/migrations

COPY --from=builder /workspace/pvz /app/pvz
COPY --from=builder /workspace/bin/grpc-health-probe /usr/local/bin/grpc-health-probe
COPY --from=builder /workspace/bin/goose /usr/local/bin/goose
COPY migrations /app/migrations

EXPOSE 50051 50052 8080 8082

ENTRYPOINT ["sh","-c","chown -R app:app /app/logs && exec su-exec app /app/pvz"]