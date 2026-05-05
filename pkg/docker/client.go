package docker

import (
	"context"
	"fmt"
	"log"

	"lograf/internal/models"

	"github.com/moby/moby/client"
)

// Client wraps the Docker client.
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client.
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return &Client{cli: cli}, nil
}

// GetClient returns the underlying Docker client.
func (c *Client) GetClient() *client.Client {
	return c.cli
}

// GetContainers fetches active containers with docker-compose labels.
func (c *Client) GetContainers(ctx context.Context) ([]models.ContainerInfo, error) {
	result, err := c.cli.ContainerList(ctx, client.ContainerListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var infos []models.ContainerInfo
	for _, container := range result.Items {
		labels := container.Labels
		service := labels["com.docker.compose.service"]
		if service == "" {
			continue // Skip non-compose containers
		}

		info := models.ContainerInfo{
			ID:      container.ID,
			Name:    container.Names[0], // Use first name
			Labels:  labels,
			Service: service,
		}
		infos = append(infos, info)
	}

	log.Printf("Found %d containers with compose labels", len(infos))
	return infos, nil
}

// Close closes the Docker client.
func (c *Client) Close() error {
	return c.cli.Close()
}
