package filesystem

import (
	"fmt"
	"time"
)

// ExampleFileSystemConfig demonstrates how to create and use filesystem configuration
func ExampleFileSystemConfig() {
	// Create filesystem configuration
	config := FileSystemConfig{
		BasePath: "/tmp",
		Timeout:  10 * time.Second,
	}

	// Print configuration details
	fmt.Printf("Base Path: %s\n", config.BasePath)
	fmt.Printf("Timeout: %v\n", config.Timeout)

	// Output:
	// Base Path: /tmp
	// Timeout: 10s
}

// ExampleFileInfo demonstrates how to work with file information
func ExampleFileInfo() {
	// Create file info
	fileInfo := FileInfo{
		Name:    "example.txt",
		Size:    1024,
		Mode:    0644,
		ModTime: time.Now(),
		IsDir:   false,
	}

	// Print file information
	fmt.Printf("File Name: %s\n", fileInfo.Name)
	fmt.Printf("File Size: %d bytes\n", fileInfo.Size)
	fmt.Printf("Is Directory: %t\n", fileInfo.IsDir)

	// Output:
	// File Name: example.txt
	// File Size: 1024 bytes
	// Is Directory: false
}