# IoT Sensor Dashboard

A lightweight, real-time IoT sensor monitoring dashboard built with Go, HTMX, and TimescaleDB.

## Features

- **Live Sensor Data** — Real-time updates via Server-Sent Events (SSE)
- **Sensor Groups** — Organize sensors into logical groups
- **Time-Series Storage** — TimescaleDB hypertables for efficient sensor readings
- **Alerting** — Configurable threshold-based alerts per sensor
- **Sensor Simulator** — Built-in data simulator for development and testing
- **Server-Side Rendering** — Fast HTML responses with Go templates + HTMX
- **No JavaScript Framework** — Minimal client JS via HTMX for dynamic updates
- **REST API** — JSON endpoints for sensor CRUD, readings, and groups
- **Dockerized** — Single docker-compose for app + TimescaleDB

## Tech Stack

- **Language:** Go 1.22
- **Router:** chi v5
- **Frontend:** HTMX + Go html/template
- **Database:** TimescaleDB (PostgreSQL extension)
- **Real-Time:** Server-Sent Events (SSE)
- **Testing:** Go testing + testify
- **Containerization:** Docker + Docker Compose

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose

### Installation

```bash
git clone <repo-url>
cd iot-dashboard
```

### Run

```bash
# With Docker (recommended)
docker-compose up

# Or manually (requires TimescaleDB running)
go run cmd/server/main.go
```

### Environment Variables

Configure via environment variables or a `.env` file:

| Variable       | Description              | Default               |
|----------------|--------------------------|-----------------------|
| `DATABASE_URL` | TimescaleDB connection   | `postgres://...`      |
| `PORT`         | HTTP server port         | `8080`                |
| `SIMULATE`     | Enable sensor simulator  | `true`                |

## Project Structure

```
cmd/
└── server/
    └── main.go              # Application entry point
internal/
├── handlers/
│   └── handlers.go          # HTTP route handlers
├── middleware/
│   └── middleware.go         # Logging, CORS, recovery
├── models/
│   └── models.go            # Sensor, Reading, Group, Alert structs
├── services/
│   ├── database.go          # DB connection and migrations
│   ├── sensor_service.go    # Sensor CRUD operations
│   ├── reading_service.go   # Time-series read/write
│   ├── group_service.go     # Sensor group management
│   ├── alert_service.go     # Alert rules and evaluation
│   └── sse.go               # Server-Sent Events broker
├── simulator/
│   └── simulator.go         # Fake sensor data generator
└── templates/
    ├── engine.go            # Template rendering engine
    ├── layout.go            # Base HTML layout
    ├── pages.go             # Page templates
    └── partials.go          # Reusable HTMX fragments
```

## License

MIT
