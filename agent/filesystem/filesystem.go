package filesystem

import (
	"context"
	"io/fs"
	"time"
)

// FileSystemInterface defines the interface for file system operations
type FileSystemInterface interface {
	// File operations
	ReadFile(ctx context.Context, path string) ([]byte, error)
	WriteFile(ctx context.Context, path string, data []byte, perm fs.FileMode) error
	ListFiles(ctx context.Context, path string) ([]FileInfo, error)
	Exists(ctx context.Context, path string) (bool, error)
	DeleteFile(ctx context.Context, path string) error

	// Directory operations
	CreateDir(ctx context.Context, path string, perm fs.FileMode) error
	RemoveDir(ctx context.Context, path string) error

	// File info
	GetFileInfo(ctx context.Context, path string) (FileInfo, error)
}

// FileInfo represents file information
type FileInfo struct {
	Name    string
	Size    int64
	Mode    fs.FileMode
	ModTime time.Time
	IsDir   bool
}

// FileSystemConfig represents the configuration for file system operations
type FileSystemConfig struct {
	BasePath string
	Timeout  time.Duration
}