package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileSystem struct {
	root string
	mu   sync.RWMutex
}

func NewFileSystem(root string) (*FileSystem, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve storage root: %w", err)
	}
	return &FileSystem{root: abs}, nil
}

func (fs *FileSystem) Save(path string, data []byte) error {
	full, err := fs.safePath(path)
	if err != nil {
		return err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	dir := filepath.Dir(full)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	tmp := full + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmp, full); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

func (fs *FileSystem) Get(path string) ([]byte, error) {
	full, err := fs.safePath(path)
	if err != nil {
		return nil, err
	}

	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := os.ReadFile(full)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (fs *FileSystem) Exists(path string) bool {
	full, err := fs.safePath(path)
	if err != nil {
		return false
	}

	fs.mu.RLock()
	defer fs.mu.RUnlock()

	info, err := os.Stat(full)
	return err == nil && !info.IsDir()
}

func (fs *FileSystem) List() ([]string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var files []string
	err := filepath.Walk(fs.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasSuffix(path, ".tmp") {
			return nil
		}
		rel, err := filepath.Rel(fs.root, path)
		if err != nil {
			return err
		}
		files = append(files, "/"+filepath.ToSlash(rel))
		return nil
	})
	return files, err
}

func (fs *FileSystem) safePath(path string) (string, error) {
	cleaned := filepath.FromSlash(filepath.Clean(path))
	full := filepath.Join(fs.root, cleaned)

	abs, err := filepath.Abs(full)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	if !strings.HasPrefix(abs, fs.root) {
		return "", fmt.Errorf("path traversal denied: %s", path)
	}
	return abs, nil
}
