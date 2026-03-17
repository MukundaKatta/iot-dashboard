package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/iot-dashboard/internal/models"
)

// SSEEvent represents an event to broadcast
type SSEEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// SSEBroker manages SSE client connections
type SSEBroker struct {
	clients    map[chan SSEEvent]bool
	register   chan chan SSEEvent
	unregister chan chan SSEEvent
	broadcast  chan SSEEvent
	mu         sync.RWMutex
}

// NewSSEBroker creates a new SSE broker
func NewSSEBroker() *SSEBroker {
	b := &SSEBroker{
		clients:    make(map[chan SSEEvent]bool),
		register:   make(chan chan SSEEvent),
		unregister: make(chan chan SSEEvent),
		broadcast:  make(chan SSEEvent, 256),
	}
	go b.run()
	return b
}

func (b *SSEBroker) run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = true
			b.mu.Unlock()
			log.Printf("SSE client connected (total: %d)", len(b.clients))

		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client)
			}
			b.mu.Unlock()
			log.Printf("SSE client disconnected (total: %d)", len(b.clients))

		case event := <-b.broadcast:
			b.mu.RLock()
			for client := range b.clients {
				select {
				case client <- event:
				default:
					// Client buffer full, skip
				}
			}
			b.mu.RUnlock()
		}
	}
}

// Broadcast sends an event to all connected clients
func (b *SSEBroker) Broadcast(event SSEEvent) {
	b.broadcast <- event
}

// BroadcastReading sends a sensor reading to all clients
func (b *SSEBroker) BroadcastReading(reading models.SensorReading) {
	b.Broadcast(SSEEvent{
		Event: "reading",
		Data:  reading,
	})
}

// BroadcastAlert sends an alert to all clients
func (b *SSEBroker) BroadcastAlert(alert models.Alert) {
	b.Broadcast(SSEEvent{
		Event: "alert",
		Data:  alert,
	})
}

// BroadcastSensorStatus sends a sensor status update
func (b *SSEBroker) BroadcastSensorStatus(sensor models.Sensor) {
	b.Broadcast(SSEEvent{
		Event: "sensor_status",
		Data:  sensor,
	})
}

// ServeHTTP handles SSE connections
func (b *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	clientChan := make(chan SSEEvent, 64)
	b.register <- clientChan

	defer func() {
		b.unregister <- clientChan
	}()

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-clientChan:
			if !ok {
				return
			}
			data, err := json.Marshal(event.Data)
			if err != nil {
				log.Printf("SSE marshal error: %v", err)
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, string(data))
			flusher.Flush()
		}
	}
}

// ClientCount returns the number of connected clients
func (b *SSEBroker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}
