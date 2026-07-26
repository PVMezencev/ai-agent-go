package filesystem

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileSystemConfig_Create(t *testing.T) {
	// Test creating filesystem configuration
	config := FileSystemConfig{
		BasePath: "/tmp",
		Timeout:  10 * time.Second,
	}

	assert.Equal(t, "/tmp", config.BasePath)
	assert.Equal(t, 10*time.Second, config.Timeout)
}

func TestFileInfo_Create(t *testing.T) {
	// Test creating file info
	fileInfo := FileInfo{
		Name:    "test.txt",
		Size:    1024,
		Mode:    0644,
		ModTime: time.Now(),
		IsDir:   false,
	}

	assert.Equal(t, "test.txt", fileInfo.Name)
	assert.Equal(t, int64(1024), fileInfo.Size)
	assert.Equal(t, 0644, fileInfo.Mode)
	assert.False(t, fileInfo.IsDir)
}