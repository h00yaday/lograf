package streamer

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"

	"lograf/internal/models"
)

type fakeLogReader struct {
	logs map[string]string
}

func (f fakeLogReader) ContainerLogs(ctx context.Context, containerID string, options client.ContainerLogsOptions) (client.ContainerLogsResult, error) {
	logData, ok := f.logs[containerID]
	if !ok {
		return nil, io.EOF
	}
	return io.NopCloser(strings.NewReader(logData)), nil
}

func TestNewLogStreamer_ValidRegex(t *testing.T) {
	s, err := NewLogStreamer(fakeLogReader{}, "INFO|DEBUG")
	require.NoError(t, err)
	require.NotNil(t, s)
}

func TestNewLogStreamer_InvalidRegex(t *testing.T) {
	_, err := NewLogStreamer(fakeLogReader{}, "[")
	require.Error(t, err)
}

func TestParseLogLine_ReturnsLogEntry(t *testing.T) {
	s := &LogStreamer{}
	container := models.ContainerInfo{ID: "id-123", Service: "web"}
	entry := s.parseLogLine(container, "2026-05-06T12:00:00Z INFO hello world")

	require.Equal(t, "id-123", entry.ContainerID)
	require.Equal(t, "web", entry.Service)
	require.Equal(t, "INFO", entry.Level)
	require.Equal(t, "2026-05-06T12:00:00Z INFO hello world", entry.Raw)
	require.Equal(t, "hello world", entry.Message)
	require.Equal(t, time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC), entry.Timestamp)
}

func TestParseLogLine_WithoutTimestamp(t *testing.T) {
	s := &LogStreamer{}
	container := models.ContainerInfo{ID: "id-123", Service: "web"}
	entry := s.parseLogLine(container, "ERROR failed to connect")

	require.Equal(t, "ERROR", entry.Level)
	require.Equal(t, "failed to connect", entry.Message)
	require.WithinDuration(t, time.Now(), entry.Timestamp, time.Second)
}

func TestStreamLogs_FiltersOutNonMatchingLines(t *testing.T) {
	api := fakeLogReader{logs: map[string]string{
		"id-123": "INFO good line\nTRACE ignored line\nDEBUG ok line\n",
	}}

	s, err := NewLogStreamer(api, "INFO|DEBUG")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	containers := []models.ContainerInfo{{ID: "id-123", Service: "web"}}
	logCh, stop := s.StreamLogs(ctx, containers)
	defer stop()

	entries := make([]models.LogEntry, 0, 2)
	for entry := range logCh {
		entries = append(entries, entry)
	}

	require.Len(t, entries, 2)
	require.Equal(t, "good line", entries[0].Message)
	require.Equal(t, "ok line", entries[1].Message)
	require.Equal(t, "INFO", entries[0].Level)
	require.Equal(t, "DEBUG", entries[1].Level)
}
