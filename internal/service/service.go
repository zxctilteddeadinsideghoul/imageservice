package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"imagesv2/internal/config"
	"imagesv2/internal/storage"
)

type Service struct {
	storage   storage.Storage
	client    *http.Client
	cfg       *config.Config
	semaphore chan struct{}
}

func New(cfg *config.Config, s storage.Storage) *Service {
	return &Service{
		storage: s,
		client: &http.Client{
			Timeout: cfg.DownloadTimeout,
		},
		cfg:       cfg,
		semaphore: make(chan struct{}, cfg.MaxConcurrentDownloads),
	}
}

func (s *Service) UploadFromURL(ctx context.Context, rawURL string) (string, error) {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
	if u.Path == "" || u.Path == "/" {
		return "", fmt.Errorf("url has no file path")
	}

	select {
	case s.semaphore <- struct{}{}:
		defer func() { <-s.semaphore }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		return "", fmt.Errorf("not an image: content-type %q", ct)
	}

	limited := io.LimitReader(resp.Body, s.cfg.MaxFileSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if int64(len(data)) > s.cfg.MaxFileSize {
		return "", fmt.Errorf("file too large: exceeds %d bytes", s.cfg.MaxFileSize)
	}

	path := u.Path
	if err := s.storage.Save(path, data); err != nil {
		return "", fmt.Errorf("save: %w", err)
	}

	slog.Info("image saved", "path", path, "size", len(data))
	return s.cfg.PublicURL + path, nil
}

func (s *Service) GetImage(path string) ([]byte, error) {
	return s.storage.Get(path)
}

func (s *Service) ImageExists(path string) bool {
	return s.storage.Exists(path)
}

func (s *Service) ListImages() ([]string, error) {
	paths, err := s.storage.List()
	if err != nil {
		return nil, err
	}
	urls := make([]string, len(paths))
	for i, p := range paths {
		urls[i] = s.cfg.PublicURL + p
	}
	return urls, nil
}
