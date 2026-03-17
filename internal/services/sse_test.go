package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iot-dashboard/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSSEBroker_ClientManagement(t *testing.T) {
	broker := NewSSEBroker()

	// Initially no clients
	assert.Equal(t, 0, broker.ClientCount())

	// Register a client
	client := make(chan SSEEvent, 64)
	broker.register <- client
	time.Sleep(50 * time.Millisecond) // Let the goroutine process
	assert.Equal(t, 1, broker.ClientCount())

	// Register another
	client2 := make(chan SSEEvent, 64)
	broker.register <- client2
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 2, broker.ClientCount())

	// Unregister one
	broker.unregister <- client
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, broker.ClientCount())

	// Unregister the other
	broker.unregister <- client2
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, broker.ClientCount())
}

func TestSSEBroker_Broadcast(t *testing.T) {
	broker := NewSSEBroker()

	// Register a client
	client := make(chan SSEEvent, 64)
	broker.register <- client
	time.Sleep(50 * time.Millisecond)

	// Broadcast a reading
	reading := models.SensorReading{
		SensorID: uuid.New(),
		Value:    22.5,
		Quality:  95,
		Time:     time.Now(),
	}
	broker.BroadcastReading(reading)

	// Client should receive the event
	select {
	case event := <-client:
		assert.Equal(t, "reading", event.Event)
		r, ok := event.Data.(models.SensorReading)
		assert.True(t, ok)
		assert.Equal(t, reading.Value, r.Value)
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for broadcast")
	}

	broker.unregister <- client
}

func TestSSEBroker_BroadcastAlert(t *testing.T) {
	broker := NewSSEBroker()

	client := make(chan SSEEvent, 64)
	broker.register <- client
	time.Sleep(50 * time.Millisecond)

	alert := models.Alert{
		ID:       uuid.New(),
		SensorID: uuid.New(),
		Message:  "Temperature too high",
		Severity: models.AlertSeverityCritical,
		Value:    35.5,
	}
	broker.BroadcastAlert(alert)

	select {
	case event := <-client:
		assert.Equal(t, "alert", event.Event)
		a, ok := event.Data.(models.Alert)
		assert.True(t, ok)
		assert.Equal(t, alert.Message, a.Message)
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for alert broadcast")
	}

	broker.unregister <- client
}

func TestSSEBroker_BroadcastSensorStatus(t *testing.T) {
	broker := NewSSEBroker()

	client := make(chan SSEEvent, 64)
	broker.register <- client
	time.Sleep(50 * time.Millisecond)

	sensor := models.Sensor{
		ID:     uuid.New(),
		Name:   "Test Sensor",
		Status: models.SensorStatusOffline,
	}
	broker.BroadcastSensorStatus(sensor)

	select {
	case event := <-client:
		assert.Equal(t, "sensor_status", event.Event)
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for status broadcast")
	}

	broker.unregister <- client
}
