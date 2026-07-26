package filesystem

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// LocalFileSystem implements the FileSystemInterface for local file system operations
type LocalFileSystem struct {
	config FileSystemConfig
}

// NewLocalFileSystem creates a new local file system instance
func NewLocalFileSystem(config FileSystemConfig) *LocalFileSystem {
	return &LocalFileSystem{
		config: config,
	}
}

// ReadFile reads a file from the file system
func (fsys *LocalFileSystem) ReadFile(ctx context.Context, path string) ([]byte, error) {
	// Resolve the full path
	fullPath := filepath.Join(fsys.config.BasePath, path)

	// Check if file exists and is within base path
	if !isPathWithinBasePath(fsys.config.BasePath, fullPath) {
		return nil, os.ErrPermission
	}

	return os.ReadFile(fullPath)
}

// WriteFile writes data to a file
func (fsys *LocalFileSystem) WriteFile(ctx context.Context, path string, data []byte, perm fs.FileMode) error {
	// Resolve the full path
	fullPath := filepath.Join(fsys.config.BasePath, path)

	// Ensure the directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Check if path is within base path
	if !isPathWithinBasePath(fsys.config.BasePath, fullPath) {
		return os.ErrPermission
	}

	return os.WriteFile(fullPath, data, perm)
}

// ListFiles lists files in a directory
func (fsys *LocalFileSystem) ListFiles(ctx context.Context, path string) ([]FileInfo, error) {
	// Resolve the full path
	fullPath := filepath.Join(fsys.config.BasePath, path)

	// Check if path is within base path
	if !isPathWithinBasePath(fsys.config.BasePath, fullPath) {
		return nil, os.ErrPermission
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
		})
	}

	return files, nil
}

// Exists checks if a file or directory exists
func (fsys *LocalFileSystem) Exists(ctx context.Context, path string) (bool, error) {
	// Resolve the full path
	fullPath := filepath.Join(fsys.config.BasePath, path)

	// Check if path is within base path
	if !isPathWithinBasePath(fsys.config.BasePath, fullPath) {
		return false, os.ErrPermission
	}

	_, err := os.Stat(fullPath)
	return err == nil, err
}

// DeleteFile deletes a file
func (fsys *LocalFileSystem) DeleteFile(ctx context.Context, path string) error {
	// Resolve the full path
	fullPath := filepath.Join(fsys.config.BasePath, path)

	// Check if path is within base path
	if !isPathWithinBasePath(fsys.config.BasePath, fullPath) {
		return os.ErrPermission
	}

	return os.Remove(fullPath)
}

// CreateDir creates a directory
func (fsys *LocalFileSystem) CreateDir(ctx context.Context, path string, perm fs.FileMode) error {
	// Resolve the full path
	fullPath := filepath.Join(fsys.config.BasePath, path)

	// Check if path is within base path
	if !isPathWithinBasePath(fsys.config.BasePath, fullPath) {
		return os.ErrPermission
	}

	return os.MkdirAll(fullPath, perm)
}

// RemoveDir removes a directory
func (fsys *LocalFileSystem) RemoveDir(ctx context.Context, path string) error {
	// Resolve the full path
	fullPath := filepath.Join(fsys.config.BasePath, path)

	// Check if path is within base path
	if !isPathWithinBasePath(fsys.config.BasePath, fullPath) {
		return os.ErrPermission
	}

	return os.RemoveAll(fullPath)
}

// GetFileInfo gets file information
func (fsys *LocalFileSystem) GetFileInfo(ctx context.Context, path string) (FileInfo, error) {
	// Resolve the full path
	fullPath := filepath.Join(fsys.config.BasePath, path)

	// Check if path is within base path
	if !isPathWithinBasePath(fsys.config.BasePath, fullPath) {
		return FileInfo{}, os.ErrPermission
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Name:    filepath.Base(fullPath),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

// isPathWithinBasePath checks if a path is within the base path
func isPathWithinBasePath(basePath, fullPath string) bool {
	rel, err := filepath.Rel(basePath, fullPath)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

// Import for strings package
import "strings"