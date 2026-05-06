package docker

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
)

type fakeDockerAPI struct {
	result client.ContainerListResult
	err    error
}

func (f fakeDockerAPI) ContainerLogs(ctx context.Context, containerID string, options client.ContainerLogsOptions) (client.ContainerLogsResult, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

func (f fakeDockerAPI) ContainerList(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error) {
	return f.result, f.err
}

func (f fakeDockerAPI) Close() error {
	return nil
}

func TestGetContainersFiltersComposeServices(t *testing.T) {
	fake := fakeDockerAPI{result: client.ContainerListResult{
		Items: []container.Summary{
			{
				ID:    "id-123",
				Names: []string{"/log-spammer"},
				Labels: map[string]string{
					"com.docker.compose.service": "log-spammer",
				},
			},
			{
				ID:     "id-456",
				Names:  []string{"/other"},
				Labels: map[string]string{},
			},
		},
	}}

	client := &Client{cli: fake}
	containers, err := client.GetContainers(context.Background())
	require.NoError(t, err)
	require.Len(t, containers, 1)
	require.Equal(t, "id-123", containers[0].ID)
	require.Equal(t, "log-spammer", containers[0].Service)
}

func TestGetContainersReturnsError(t *testing.T) {
	expected := errors.New("api unavailable")
	fake := fakeDockerAPI{err: expected}
	client := &Client{cli: fake}

	_, err := client.GetContainers(context.Background())
	require.Error(t, err)
	require.ErrorIs(t, err, expected)
}

func TestGetClientReturnsLogReader(t *testing.T) {
	client := &Client{cli: fakeDockerAPI{}}
	logReader := client.GetClient()
	require.NotNil(t, logReader)
}
