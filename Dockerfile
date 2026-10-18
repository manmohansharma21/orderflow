# Multi-stage build for the OrderFlow API.
FROM golang:1.26-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/orderflow-api ./cmd/api

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S orderflow \
    && adduser -S -G orderflow orderflow
WORKDIR /app
COPY --from=builder /out/orderflow-api /app/orderflow-api
USER orderflow
EXPOSE 8080
ENTRYPOINT ["/app/orderflow-api"]
