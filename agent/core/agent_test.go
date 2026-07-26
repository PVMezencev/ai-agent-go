package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgent_Create(t *testing.T) {
	// Test creating a new agent
	agent := Agent{
		ID:        "test-agent",
		Name:      "Test Agent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "test-agent", agent.ID)
	assert.Equal(t, "Test Agent", agent.Name)
	assert.False(t, agent.CreatedAt.IsZero())
	assert.False(t, agent.UpdatedAt.IsZero())
}

func TestAgent_GetStatus(t *testing.T) {
	// Test getting agent status
	agent := Agent{
		ID:        "test-agent",
		Name:      "Test Agent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	status := AgentStatus{
		IsRunning:   true,
		LastActive:  time.Now(),
		Error:       "",
		Stats:       make(map[string]interface{}),
	}

	assert.NotNil(t, status)
	assert.True(t, status.IsRunning)
}

func TestAgent_Config(t *testing.T) {
	// Test agent configuration
	config := AgentConfig{
		ID:          "config-test",
		Name:        "Config Test Agent",
		MaxRetries:  3,
		Timeout:     30 * time.Second,
		Environment: "test",
		Enabled:     true,
	}

	assert.Equal(t, "config-test", config.ID)
	assert.Equal(t, "Config Test Agent", config.Name)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, "test", config.Environment)
	assert.True(t, config.Enabled)
}