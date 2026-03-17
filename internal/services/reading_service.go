package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iot-dashboard/internal/models"
)

// ReadingService handles sensor reading operations
type ReadingService struct {
	db *sql.DB
}

// NewReadingService creates a new ReadingService
func NewReadingService(db *sql.DB) *ReadingService {
	return &ReadingService{db: db}
}

// InsertReading inserts a new sensor reading
func (s *ReadingService) InsertReading(reading *models.SensorReading) error {
	query := `INSERT INTO sensor_readings (time, sensor_id, value, quality) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(query, reading.Time, reading.SensorID, reading.Value, reading.Quality)
	if err != nil {
		return fmt.Errorf("insert reading: %w", err)
	}
	return nil
}

// InsertBatchReadings inserts multiple readings in a single transaction
func (s *ReadingService) InsertBatchReadings(readings []models.SensorReading) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO sensor_readings (time, sensor_id, value, quality) VALUES ($1, $2, $3, $4)`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, r := range readings {
		_, err := stmt.Exec(r.Time, r.SensorID, r.Value, r.Quality)
		if err != nil {
			return fmt.Errorf("exec reading: %w", err)
		}
	}

	return tx.Commit()
}

// GetLatestReadings returns the most recent readings for a sensor
func (s *ReadingService) GetLatestReadings(sensorID uuid.UUID, limit int) ([]models.SensorReading, error) {
	query := `SELECT time, sensor_id, value, quality
		FROM sensor_readings WHERE sensor_id = $1
		ORDER BY time DESC LIMIT $2`

	rows, err := s.db.Query(query, sensorID, limit)
	if err != nil {
		return nil, fmt.Errorf("query readings: %w", err)
	}
	defer rows.Close()

	var readings []models.SensorReading
	for rows.Next() {
		var r models.SensorReading
		if err := rows.Scan(&r.Time, &r.SensorID, &r.Value, &r.Quality); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}

// GetReadingsInRange returns readings within a time range
func (s *ReadingService) GetReadingsInRange(sensorID uuid.UUID, start, end time.Time) ([]models.SensorReading, error) {
	query := `SELECT time, sensor_id, value, quality
		FROM sensor_readings
		WHERE sensor_id = $1 AND time >= $2 AND time <= $3
		ORDER BY time ASC`

	rows, err := s.db.Query(query, sensorID, start, end)
	if err != nil {
		return nil, fmt.Errorf("query readings range: %w", err)
	}
	defer rows.Close()

	var readings []models.SensorReading
	for rows.Next() {
		var r models.SensorReading
		if err := rows.Scan(&r.Time, &r.SensorID, &r.Value, &r.Quality); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}

// GetHourlyAggregates returns hourly aggregate data
func (s *ReadingService) GetHourlyAggregates(sensorID uuid.UUID, hours int) ([]models.HourlyAggregate, error) {
	query := `SELECT bucket, sensor_id, avg_value, min_value, max_value, count
		FROM sensor_readings_hourly
		WHERE sensor_id = $1 AND bucket >= NOW() - make_interval(hours => $2)
		ORDER BY bucket ASC`

	rows, err := s.db.Query(query, sensorID, hours)
	if err != nil {
		// Fallback to raw query if continuous aggregate not ready
		return s.getHourlyAggregatesFallback(sensorID, hours)
	}
	defer rows.Close()

	var aggs []models.HourlyAggregate
	for rows.Next() {
		var a models.HourlyAggregate
		if err := rows.Scan(&a.Bucket, &a.SensorID, &a.AvgValue, &a.MinValue, &a.MaxValue, &a.Count); err != nil {
			return nil, err
		}
		aggs = append(aggs, a)
	}
	if len(aggs) == 0 {
		return s.getHourlyAggregatesFallback(sensorID, hours)
	}
	return aggs, rows.Err()
}

func (s *ReadingService) getHourlyAggregatesFallback(sensorID uuid.UUID, hours int) ([]models.HourlyAggregate, error) {
	query := `SELECT time_bucket('1 hour', time) AS bucket, sensor_id,
		AVG(value) AS avg_value, MIN(value) AS min_value, MAX(value) AS max_value, COUNT(*) AS count
		FROM sensor_readings
		WHERE sensor_id = $1 AND time >= NOW() - make_interval(hours => $2)
		GROUP BY bucket, sensor_id
		ORDER BY bucket ASC`

	rows, err := s.db.Query(query, sensorID, hours)
	if err != nil {
		return nil, fmt.Errorf("hourly fallback: %w", err)
	}
	defer rows.Close()

	var aggs []models.HourlyAggregate
	for rows.Next() {
		var a models.HourlyAggregate
		if err := rows.Scan(&a.Bucket, &a.SensorID, &a.AvgValue, &a.MinValue, &a.MaxValue, &a.Count); err != nil {
			return nil, err
		}
		aggs = append(aggs, a)
	}
	return aggs, rows.Err()
}

// GetDailyAggregates returns daily aggregate data
func (s *ReadingService) GetDailyAggregates(sensorID uuid.UUID, days int) ([]models.DailyAggregate, error) {
	query := `SELECT time_bucket('1 day', time) AS bucket, sensor_id,
		AVG(value) AS avg_value, MIN(value) AS min_value, MAX(value) AS max_value, COUNT(*) AS count
		FROM sensor_readings
		WHERE sensor_id = $1 AND time >= NOW() - make_interval(days => $2)
		GROUP BY bucket, sensor_id
		ORDER BY bucket ASC`

	rows, err := s.db.Query(query, sensorID, days)
	if err != nil {
		return nil, fmt.Errorf("daily aggs: %w", err)
	}
	defer rows.Close()

	var aggs []models.DailyAggregate
	for rows.Next() {
		var a models.DailyAggregate
		if err := rows.Scan(&a.Bucket, &a.SensorID, &a.AvgValue, &a.MinValue, &a.MaxValue, &a.Count); err != nil {
			return nil, err
		}
		aggs = append(aggs, a)
	}
	return aggs, rows.Err()
}

// GetAllLatestReadings returns the most recent reading for every sensor
func (s *ReadingService) GetAllLatestReadings() (map[uuid.UUID]models.SensorReading, error) {
	query := `SELECT DISTINCT ON (sensor_id) time, sensor_id, value, quality
		FROM sensor_readings
		ORDER BY sensor_id, time DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("all latest: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]models.SensorReading)
	for rows.Next() {
		var r models.SensorReading
		if err := rows.Scan(&r.Time, &r.SensorID, &r.Value, &r.Quality); err != nil {
			return nil, err
		}
		result[r.SensorID] = r
	}
	return result, rows.Err()
}

// GetSparklineData returns the last N readings for sparkline display
func (s *ReadingService) GetSparklineData(sensorID uuid.UUID, points int) ([]float64, error) {
	query := `SELECT value FROM (
		SELECT value, time FROM sensor_readings
		WHERE sensor_id = $1 ORDER BY time DESC LIMIT $2
	) sub ORDER BY time ASC`

	rows, err := s.db.Query(query, sensorID, points)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []float64
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
