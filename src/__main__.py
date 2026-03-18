"""CLI for iot-dashboard."""
import sys, json, argparse
from .core import IotDashboard

def main():
    parser = argparse.ArgumentParser(description="IoT sensor monitoring dashboard with Go, HTMX, TimescaleDB, and real-time SSE")
    parser.add_argument("command", nargs="?", default="status", choices=["status", "run", "info"])
    parser.add_argument("--input", "-i", default="")
    args = parser.parse_args()
    instance = IotDashboard()
    if args.command == "status":
        print(json.dumps(instance.get_stats(), indent=2))
    elif args.command == "run":
        print(json.dumps(instance.detect(input=args.input or "test"), indent=2, default=str))
    elif args.command == "info":
        print(f"iot-dashboard v0.1.0 — IoT sensor monitoring dashboard with Go, HTMX, TimescaleDB, and real-time SSE")

if __name__ == "__main__":
    main()
