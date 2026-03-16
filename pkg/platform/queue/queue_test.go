package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient("redis://localhost:6379")
	assert.NotNil(t, client)
	assert.NotNil(t, client.inner)
}

func TestNewServer(t *testing.T) {
	server := NewServer("redis://localhost:6379", 5, 10)
	assert.NotNil(t, server)
}

func TestNewServer_DefaultConcentrations(t *testing.T) {
	server := NewServer("redis://localhost:6379", 1, 1)
	assert.NotNil(t, server)
}

func TestClient_Close(t *testing.T) {
	client := NewClient("redis://localhost:6379")
	err := client.Close()
	assert.NoError(t, err)
}
