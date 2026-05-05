package streamer

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"lograf/internal/models"

	"github.com/moby/moby/client"
)

// LogStreamer handles streaming logs from containers.
type LogStreamer struct {
	cli        *client.Client
	filter     *regexp.Regexp
	bufferPool *sync.Pool
}

// NewLogStreamer creates a new LogStreamer with a regex filter.
func NewLogStreamer(cli *client.Client, filterPattern string) (*LogStreamer, error) {
	filter, err := regexp.Compile(filterPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	bufferPool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, 4096) // 4KB buffer
		},
	}

	return &LogStreamer{
		cli:        cli,
		filter:     filter,
		bufferPool: bufferPool,
	}, nil
}

// StreamLogs starts streaming logs from the given containers.
// It returns a channel of LogEntry and a function to stop streaming.
func (s *LogStreamer) StreamLogs(ctx context.Context, containers []models.ContainerInfo) (<-chan models.LogEntry, func()) {
	logCh := make(chan models.LogEntry, 100) // Buffered channel to prevent blocking
	ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup

	for _, container := range containers {
		wg.Add(1)
		go func(cont models.ContainerInfo) {
			defer wg.Done()
			s.streamContainerLogs(ctx, cont, logCh)
		}(container)
	}

	// Goroutine to close channel when all streamers are done
	go func() {
		wg.Wait()
		close(logCh)
	}()

	return logCh, cancel
}

// streamContainerLogs streams logs from a single container.
func (s *LogStreamer) streamContainerLogs(ctx context.Context, container models.ContainerInfo, logCh chan<- models.LogEntry) {
	options := client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "0", // Start from the end
	}

	reader, err := s.cli.ContainerLogs(ctx, container.ID, options)
	if err != nil {
		log.Printf("Failed to get logs for container %s: %v", container.ID, err)
		return
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	// Use pooled buffer for scanner
	buf := s.bufferPool.Get().([]byte)
	defer s.bufferPool.Put(buf)
	scanner.Buffer(buf, len(buf))

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if s.filter.MatchString(line) {
			entry := s.parseLogLine(container, line)
			select {
			case logCh <- entry:
			case <-ctx.Done():
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading logs from container %s: %v", container.ID, err)
	}
}

// parseLogLine parses a raw log line into a LogEntry.
// This is a basic parser; assumes format like "2023-01-01T00:00:00Z INFO message"
func (s *LogStreamer) parseLogLine(container models.ContainerInfo, line string) models.LogEntry {
	// Simple parsing; in real app, use more robust parsing
	entry := models.LogEntry{
		ContainerID: container.ID,
		Service:     container.Service,
		Raw:         line,
		Timestamp:   time.Now(), // Placeholder; parse from log if available
		Level:       "INFO",     // Placeholder
		Message:     line,
	}
	return entry
}
