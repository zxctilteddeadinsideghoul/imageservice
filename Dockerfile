FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o imagesvc ./cmd/server
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/imagesvc /usr/local/bin/imagesvc

RUN mkdir -p /data
EXPOSE 8080
ENTRYPOINT ["imagesvc"]