# Image Proxy Microservice

A minimalist Go microservice that downloads images by URL, stores them on the local filesystem mirroring the original URL path, and serves them back under your own domain.

## Configuration

All settings are read from environment variables:

| Variable | Default | Description |
|---|---|---|
| `LISTEN_ADDR` | `:8080` | Address to listen on |
| `STORAGE_DIR` | `./data` | Root directory for stored images |
| `PUBLIC_URL` | `http://localhost:8080` | Base URL returned in API responses |
| `MAX_FILE_SIZE` | `10485760` | Maximum image size in bytes (10 MB) |
| `DOWNLOAD_TIMEOUT` | `10s` | Timeout for downloading a remote image |
| `MAX_CONCURRENT_DOWNLOADS` | `20` | Maximum simultaneous outbound downloads |

## Run Locally

```bash
go run ./cmd/server
```

With custom settings:

```bash
PUBLIC_URL=https://images.example.com STORAGE_DIR=./images go run ./cmd/server
```

## Run with Docker

```bash
docker build -t imagesvc .
docker run -p 8080:8080 -e PUBLIC_URL=https://images.example.com -v imgdata:/data imagesvc
```

## API

### Health Check

```bash
curl http://localhost:8080/health
```

### Upload an Image by URL

```bash
curl -X POST http://localhost:8080/images \
  -H "Content-Type: application/json" \
  -d '{"url": "https://pcdn.goldapple.ru/p/p/19000009488/web/696d67416464315064708ddc4892b97e97b.jpg"}'
```

Response:

```json
{
  "url": "http://localhost:8080/p/p/19000009488/web/696d67416464315064708ddc4892b97e97b.jpg"
}
```

### Retrieve an Image

```bash
curl http://localhost:8080/p/p/19000009488/web/696d67416464315064708ddc4892b97e97b.jpg --output image.jpg
```

### Check If an Image Exists

```bash
curl http://localhost:8080/exists/p/p/19000009488/web/696d67416464315064708ddc4892b97e97b.jpg
```

Response:

```json
{
  "exists": true
}
```

### List All Stored Images

```bash
curl http://localhost:8080/images
```

Response:

```json
{
  "images": [
    {
      "url": "http://localhost:8080/p/p/19000009488/web/696d67416464315064708ddc4892b97e97b.jpg"
    }
  ]
}
```
