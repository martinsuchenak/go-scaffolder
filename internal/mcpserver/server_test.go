package mcpserver

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	server := NewServer("test")
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
}
