package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"lograf/pkg/docker"
	"lograf/pkg/streamer"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	// Get containers
	containers, err := dockerClient.GetContainers(ctx)
	if err != nil {
		log.Fatalf("Failed to get containers: %v", err)
	}

	if len(containers) == 0 {
		log.Println("No containers with docker-compose labels found")
		return
	}

	// Create streamer with a basic filter (matches everything for now)
	streamer, err := streamer.NewLogStreamer(dockerClient.GetClient(), ".*")
	if err != nil {
		log.Fatalf("Failed to create streamer: %v", err)
	}

	// Start streaming
	logCh, stopStreaming := streamer.StreamLogs(ctx, containers)
	defer stopStreaming()

	// Print logs to stdout
	for {
		select {
		case entry, ok := <-logCh:
			if !ok {
				log.Println("Log streaming stopped")
				return
			}
			fmt.Printf("[%s] %s: %s\n", entry.Service, entry.Level, entry.Message)
		case <-ctx.Done():
			log.Println("Shutting down...")
			return
		}
	}
}
