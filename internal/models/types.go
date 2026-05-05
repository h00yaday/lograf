package models

import (
	"time"
)

// ContainerInfo holds information about a Docker container.
type ContainerInfo struct {
	ID      string
	Name    string
	Labels  map[string]string
	Service string // Extracted from docker-compose labels
}

// LogEntry represents a parsed log entry.
type LogEntry struct {
	ContainerID string
	Service     string
	Timestamp   time.Time
	Level       string
	Message     string
	Raw         string
}
