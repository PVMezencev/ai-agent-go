package core

import (
	"context"
	"time"
)

// Agent represents the main AI agent structure
type Agent struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AgentInterface defines the interface for all agent operations
type AgentInterface interface {
	// Core operations
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Execute(ctx context.Context, task string) (string, error)

	// Configuration
	GetConfig() AgentConfig
	SetConfig(config AgentConfig) error

	// Status
	GetStatus() AgentStatus
}

// AgentConfig represents the agent configuration
type AgentConfig struct {
	ID           string
	Name         string
	MaxRetries   int
	Timeout      time.Duration
	Environment  string
	Enabled      bool
}

// AgentStatus represents the agent status
type AgentStatus struct {
	IsRunning   bool
	LastActive  time.Time
	Error       string
	Stats       map[string]interface{}
}