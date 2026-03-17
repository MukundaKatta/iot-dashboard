package models

import (
	"time"

	"github.com/google/uuid"
)

// SensorType represents the type of sensor
type SensorType string

const (
	SensorTypeTemperature SensorType = "temperature"
	SensorTypeHumidity    SensorType = "humidity"
	SensorTypePressure    SensorType = "pressure"
	SensorTypeCO2         SensorType = "co2"
	SensorTypeLight       SensorType = "light"
)

// SensorStatus represents the operational status of a sensor
type SensorStatus string

const (
	SensorStatusOnline  SensorStatus = "online"
	SensorStatusOffline SensorStatus = "offline"
	SensorStatusError   SensorStatus = "error"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertRuleType represents the type of alert rule
type AlertRuleType string

const (
	AlertRuleThreshold    AlertRuleType = "threshold"
	AlertRuleRateOfChange AlertRuleType = "rate_of_change"
	AlertRuleOffline      AlertRuleType = "offline"
)

// Sensor represents an IoT sensor device
type Sensor struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Type        SensorType   `json:"type" db:"type"`
	Location    string       `json:"location" db:"location"`
	GroupID     *uuid.UUID   `json:"group_id,omitempty" db:"group_id"`
	Status      SensorStatus `json:"status" db:"status"`
	Unit        string       `json:"unit" db:"unit"`
	MinValue    float64      `json:"min_value" db:"min_value"`
	MaxValue    float64      `json:"max_value" db:"max_value"`
	LastReading *float64     `json:"last_reading,omitempty" db:"last_reading"`
	LastSeenAt  *time.Time   `json:"last_seen_at,omitempty" db:"last_seen_at"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

// SensorReading represents a single sensor measurement
type SensorReading struct {
	Time     time.Time `json:"time" db:"time"`
	SensorID uuid.UUID `json:"sensor_id" db:"sensor_id"`
	Value    float64   `json:"value" db:"value"`
	Quality  int       `json:"quality" db:"quality"` // 0-100 signal quality
}

// SensorGroup represents a logical grouping of sensors
type SensorGroup struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	SensorID  uuid.UUID     `json:"sensor_id" db:"sensor_id"`
	Name      string        `json:"name" db:"name"`
	Type      AlertRuleType `json:"type" db:"type"`
	Condition string        `json:"condition" db:"condition"` // "above", "below", "equals"
	Value     float64       `json:"value" db:"value"`
	Duration  int           `json:"duration" db:"duration"` // seconds - for sustained threshold
	Severity  AlertSeverity `json:"severity" db:"severity"`
	Enabled   bool          `json:"enabled" db:"enabled"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" db:"updated_at"`
}

// Alert represents a triggered alert
type Alert struct {
	ID           uuid.UUID     `json:"id" db:"id"`
	RuleID       uuid.UUID     `json:"rule_id" db:"rule_id"`
	SensorID     uuid.UUID     `json:"sensor_id" db:"sensor_id"`
	SensorName   string        `json:"sensor_name" db:"sensor_name"`
	Message      string        `json:"message" db:"message"`
	Severity     AlertSeverity `json:"severity" db:"severity"`
	Value        float64       `json:"value" db:"value"`
	Acknowledged bool          `json:"acknowledged" db:"acknowledged"`
	TriggeredAt  time.Time     `json:"triggered_at" db:"triggered_at"`
	AckedAt      *time.Time    `json:"acked_at,omitempty" db:"acked_at"`
}

// HourlyAggregate represents hourly rollup data
type HourlyAggregate struct {
	Bucket   time.Time `json:"bucket" db:"bucket"`
	SensorID uuid.UUID `json:"sensor_id" db:"sensor_id"`
	AvgValue float64   `json:"avg_value" db:"avg_value"`
	MinValue float64   `json:"min_value" db:"min_value"`
	MaxValue float64   `json:"max_value" db:"max_value"`
	Count    int64     `json:"count" db:"count"`
}

// DailyAggregate represents daily rollup data
type DailyAggregate struct {
	Bucket   time.Time `json:"bucket" db:"bucket"`
	SensorID uuid.UUID `json:"sensor_id" db:"sensor_id"`
	AvgValue float64   `json:"avg_value" db:"avg_value"`
	MinValue float64   `json:"min_value" db:"min_value"`
	MaxValue float64   `json:"max_value" db:"max_value"`
	Count    int64     `json:"count" db:"count"`
}

// DashboardStats holds overview statistics
type DashboardStats struct {
	TotalSensors  int `json:"total_sensors"`
	OnlineSensors int `json:"online_sensors"`
	ActiveAlerts  int `json:"active_alerts"`
	ReadingsToday int `json:"readings_today"`
}

// ReadingRequest is the API payload for ingesting readings
type ReadingRequest struct {
	SensorID string  `json:"sensor_id"`
	Value    float64 `json:"value"`
	Quality  int     `json:"quality"`
}

// TimeRange represents a query time range
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// ChartData represents data formatted for Chart.js
type ChartData struct {
	Labels   []string  `json:"labels"`
	Datasets []Dataset `json:"datasets"`
}

// Dataset represents a Chart.js dataset
type Dataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BorderColor     string    `json:"borderColor"`
	BackgroundColor string    `json:"backgroundColor"`
	Fill            bool      `json:"fill"`
	Tension         float64   `json:"tension"`
}
