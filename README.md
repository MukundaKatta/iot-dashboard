# iot-dashboard

IoT sensor monitoring dashboard project with a mix of real Go backend code and stub scaffolding.

## What's actually here

The Go backend (`cmd/server/`, `internal/`) follows standard Go project layout with a server entrypoint and internal packages. The project uses Go modules (go.mod/go.sum) and includes Docker configuration. The description mentions HTMX, TimescaleDB, and real-time SSE.

There is also a `src/` directory containing a stub core.py file with placeholder methods that return fixed dictionaries - this is unrelated scaffolding and does not connect to the Go code.

## Tech stack

Go, HTMX, TimescaleDB, Docker, SSE

## Status

The Go backend has real project structure. Whether the internal packages contain full implementations or are partially stubbed would require deeper inspection.
