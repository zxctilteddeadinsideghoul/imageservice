FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/imagesvc ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/imagesvc /bin/imagesvc

RUN mkdir -p /data
ENV STORAGE_DIR=/data
EXPOSE 8080

ENTRYPOINT ["/bin/imagesvc"]
