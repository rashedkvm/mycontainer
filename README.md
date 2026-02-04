# IP & Clock Monitor

A lightweight web application that displays server and client IP addresses along with a real-time digital clock.

## Features

- Displays the server/container IP address
- Displays the client/visitor IP address
- Real-time digital clock updating every second
- Minimal, professional UI design

## Requirements

- Docker
- Go 1.21+ (for local development)
- Make

## Make Commands

| Command | Description |
|---------|-------------|
| `make dev` | Run the app locally for development (requires Go) |
| `make build` | Build Docker image for linux/arm64 |
| `make run` | Run the Docker container |
| `make up` | Build and run in one command |
| `make stop` | Stop running containers |
| `make clean` | Remove the Docker image |

## Quick Start

```bash
# Development mode (local)
make dev

# Production mode (Docker)
make up
```

## Access

Open your browser and navigate to:

```
http://localhost:8080
```
