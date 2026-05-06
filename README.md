# Lograf

**Lograf** is a lightweight Go-based log ingestion and filtering microservice designed for modern containerized deployments. It simplifies the collection of streaming logs from Docker containers, applies rule-based parsing and filtering, and forwards structured events to downstream systems like Elasticsearch, Loki, or local file sinks.

## What it Solves

Developers and operators often need a fast, low-dependency tool to capture container logs, normalize them, and filter noise before sending data to observability pipelines. Lograf provides:

- low-latency ingestion of container stdout/stderr streams
- configurable parsing rules for log line normalization
- flexible output sinks for writing structured events
- lightweight deployment for edge and ephemeral environments

## Architecture

Lograf is built as a modular Go service composed of three core components:

1. **Collector**
   - connects to Docker container streams using the Docker API
   - tails live stdout/stderr output from selected containers
   - emits raw log records into the pipeline

2. **Processor**
   - applies configurable filter rules and parsers
   - enriches output with metadata such as container name, image, and timestamp
   - supports pattern matching, level selection, and payload transformation

3. **Sink**
   - writes processed events to the configured destinations
   - supports file-based output and HTTP-compatible backends
   - designed to be extensible for future integrations

Components interact through in-memory channels, enabling a low-overhead, streaming pipeline while keeping the core service lean and modular.

## Development and Contribution

For AI agents and developers contributing to this project, refer to `Agents.md` for detailed guidelines, workflow, and strict rules (including OCP principles and testing requirements). This file ensures consistent code quality and architecture.

## Getting Started

### Prerequisites

- Go 1.25 or later
- Docker Engine installed and running
- `git` installed

### Install

```bash
git clone https://github.com/example/lograf.git
cd lograf
go build -o bin/lograf ./cmd/app
```

### Configuration

Create a configuration file `config.yaml` with the following structure:

```yaml
collector:
  docker:
    host: "unix:///var/run/docker.sock"
    include_labels:
      - "com.docker.compose.service=log-spammer"

processor:
  parser:
    patterns:
      - name: "timestamped"
        regex: "^(?P<timestamp>[0-9\.]+) (?P<level>[A-Z]+) (?P<message>.*)$"
  filters:
    level:
      include:
        - INFO
        - ERROR

sink:
  type: "file"
  file:
    path: "./logs/structured.log"
    format: "json"
```

### Run

```bash
./bin/lograf --config config.yaml
```

`Lograf` will connect to Docker, collect logs from matching containers, parse each log line, and write structured events to the configured sink.

## Usage

### Command-line Options

```bash
./bin/lograf --help
```

Example output:

```text
Usage of ./bin/lograf:
  -config string
        path to configuration file (default "config.yaml")
  -debug
        enable debug logging
  -version
        print version information and exit
```

### Example Configuration

A complete example configuration for Docker-based log collection:

```yaml
collector:
  docker:
    host: "unix:///var/run/docker.sock"
    include_labels:
      - "com.docker.compose.service=web"
      - "com.docker.compose.service=worker"

processor:
  parser:
    patterns:
      - name: "json"
        regex: "^(?P<message>.*)$"
        type: "json"
  filters:
    exclude:
      - "DEBUG"

sink:
  type: "http"
  http:
    endpoint: "http://localhost:3100/loki/api/v1/push"
    headers:
      Content-Type: "application/json"
```

### Programmatic Use

If you prefer to embed the core pipeline in another Go program, use the package API:

```go
package main

import (
    "log"
    "lograf/pkg/streamer"
)

func main() {
    config := streamer.DefaultConfig()
    config.Collector.Docker.Host = "unix:///var/run/docker.sock"
    config.Sink.Type = "file"
    config.Sink.File.Path = "./logs/structured.log"

    service, err := streamer.NewService(config)
    if err != nil {
        log.Fatal(err)
    }

    if err := service.Start(); err != nil {
        log.Fatal(err)
    }
}
```

> Note: This example demonstrates the intended API shape for a Go-based log pipeline consumer.

## Future Development / Roadmap

Lograf is designed to remain lightweight while enabling flexible extension. Future enhancements may include:

- native support for more output sinks such as Elasticsearch, Loki, and Kafka
- a plugin system for custom parsers and filters
- container metadata enrichment from Kubernetes and ECS tags
- secure TLS-enabled HTTP sink transport and authentication hooks
- metrics endpoint for Prometheus and runtime monitoring
- operator-friendly deployment manifests for Kubernetes
- cloud-native TLS, backpressure handling, and retry policies

## Contributing

Contributions are welcome. To contribute:

1. Fork the repository.
2. Create a feature branch.
3. Add tests for new behavior.
4. Open a pull request with a clear description.

Please follow standard Go idioms and maintain existing package structure.

## License

Lograf is released under the MIT License. See `LICENSE` for details.
