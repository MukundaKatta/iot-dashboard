package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iot-dashboard/internal/models"
)

// SensorService handles sensor-related database operations
type SensorService struct {
	db *sql.DB
}

// NewSensorService creates a new SensorService
func NewSensorService(db *sql.DB) *SensorService {
	return &SensorService{db: db}
}

// ListSensors returns all sensors
func (s *SensorService) ListSensors() ([]models.Sensor, error) {
	query := `SELECT id, name, type, location, group_id, status, unit,
		min_value, max_value, last_reading, last_seen_at, created_at, updated_at
		FROM sensors ORDER BY name`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query sensors: %w", err)
	}
	defer rows.Close()

	var sensors []models.Sensor
	for rows.Next() {
		var sensor models.Sensor
		err := rows.Scan(
			&sensor.ID, &sensor.Name, &sensor.Type, &sensor.Location,
			&sensor.GroupID, &sensor.Status, &sensor.Unit,
			&sensor.MinValue, &sensor.MaxValue, &sensor.LastReading,
			&sensor.LastSeenAt, &sensor.CreatedAt, &sensor.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan sensor: %w", err)
		}
		sensors = append(sensors, sensor)
	}
	return sensors, rows.Err()
}

// GetSensor returns a single sensor by ID
func (s *SensorService) GetSensor(id uuid.UUID) (*models.Sensor, error) {
	query := `SELECT id, name, type, location, group_id, status, unit,
		min_value, max_value, last_reading, last_seen_at, created_at, updated_at
		FROM sensors WHERE id = $1`

	var sensor models.Sensor
	err := s.db.QueryRow(query, id).Scan(
		&sensor.ID, &sensor.Name, &sensor.Type, &sensor.Location,
		&sensor.GroupID, &sensor.Status, &sensor.Unit,
		&sensor.MinValue, &sensor.MaxValue, &sensor.LastReading,
		&sensor.LastSeenAt, &sensor.CreatedAt, &sensor.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get sensor: %w", err)
	}
	return &sensor, nil
}

// CreateSensor creates a new sensor
func (s *SensorService) CreateSensor(sensor *models.Sensor) error {
	if sensor.ID == uuid.Nil {
		sensor.ID = uuid.New()
	}
	sensor.CreatedAt = time.Now()
	sensor.UpdatedAt = time.Now()
	sensor.Status = models.SensorStatusOffline

	query := `INSERT INTO sensors (id, name, type, location, group_id, status, unit, min_value, max_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := s.db.Exec(query,
		sensor.ID, sensor.Name, sensor.Type, sensor.Location,
		sensor.GroupID, sensor.Status, sensor.Unit,
		sensor.MinValue, sensor.MaxValue, sensor.CreatedAt, sensor.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create sensor: %w", err)
	}
	return nil
}

// UpdateSensor updates an existing sensor
func (s *SensorService) UpdateSensor(sensor *models.Sensor) error {
	sensor.UpdatedAt = time.Now()
	query := `UPDATE sensors SET name=$1, type=$2, location=$3, group_id=$4,
		unit=$5, min_value=$6, max_value=$7, updated_at=$8 WHERE id=$9`

	_, err := s.db.Exec(query,
		sensor.Name, sensor.Type, sensor.Location, sensor.GroupID,
		sensor.Unit, sensor.MinValue, sensor.MaxValue, sensor.UpdatedAt, sensor.ID,
	)
	if err != nil {
		return fmt.Errorf("update sensor: %w", err)
	}
	return nil
}

// DeleteSensor deletes a sensor
func (s *SensorService) DeleteSensor(id uuid.UUID) error {
	_, err := s.db.Exec(`DELETE FROM sensors WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete sensor: %w", err)
	}
	return nil
}

// UpdateSensorStatus updates sensor status and last seen
func (s *SensorService) UpdateSensorStatus(id uuid.UUID, status models.SensorStatus, lastReading *float64) error {
	now := time.Now()
	query := `UPDATE sensors SET status=$1, last_reading=$2, last_seen_at=$3, updated_at=$3 WHERE id=$4`
	_, err := s.db.Exec(query, status, lastReading, now, id)
	if err != nil {
		return fmt.Errorf("update sensor status: %w", err)
	}
	return nil
}

// GetDashboardStats returns overview statistics
func (s *SensorService) GetDashboardStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	err := s.db.QueryRow(`SELECT COUNT(*) FROM sensors`).Scan(&stats.TotalSensors)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(`SELECT COUNT(*) FROM sensors WHERE status = 'online'`).Scan(&stats.OnlineSensors)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(`SELECT COUNT(*) FROM alerts WHERE acknowledged = false`).Scan(&stats.ActiveAlerts)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(`SELECT COUNT(*) FROM sensor_readings WHERE time > NOW() - INTERVAL '24 hours'`).Scan(&stats.ReadingsToday)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// DetectOfflineSensors marks sensors that haven't reported recently as offline
func (s *SensorService) DetectOfflineSensors(timeout time.Duration) ([]models.Sensor, error) {
	threshold := time.Now().Add(-timeout)
	query := `UPDATE sensors SET status = 'offline', updated_at = NOW()
		WHERE status = 'online' AND last_seen_at < $1
		RETURNING id, name, type, location, group_id, status, unit,
			min_value, max_value, last_reading, last_seen_at, created_at, updated_at`

	rows, err := s.db.Query(query, threshold)
	if err != nil {
		return nil, fmt.Errorf("detect offline: %w", err)
	}
	defer rows.Close()

	var sensors []models.Sensor
	for rows.Next() {
		var sensor models.Sensor
		err := rows.Scan(
			&sensor.ID, &sensor.Name, &sensor.Type, &sensor.Location,
			&sensor.GroupID, &sensor.Status, &sensor.Unit,
			&sensor.MinValue, &sensor.MaxValue, &sensor.LastReading,
			&sensor.LastSeenAt, &sensor.CreatedAt, &sensor.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sensors = append(sensors, sensor)
	}
	return sensors, rows.Err()
}
