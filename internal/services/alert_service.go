package services

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/iot-dashboard/internal/models"
)

// AlertService handles alert rules and alerts
type AlertService struct {
	db     *sql.DB
	broker *SSEBroker
}

// NewAlertService creates a new AlertService
func NewAlertService(db *sql.DB) *AlertService {
	return &AlertService{db: db}
}

// SetBroker sets the SSE broker for broadcasting alerts
func (s *AlertService) SetBroker(broker *SSEBroker) {
	s.broker = broker
}

// ListAlertRules returns all alert rules
func (s *AlertService) ListAlertRules() ([]models.AlertRule, error) {
	query := `SELECT id, sensor_id, name, type, condition, value, duration, severity, enabled, created_at, updated_at
		FROM alert_rules ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list alert rules: %w", err)
	}
	defer rows.Close()

	var rules []models.AlertRule
	for rows.Next() {
		var r models.AlertRule
		if err := rows.Scan(&r.ID, &r.SensorID, &r.Name, &r.Type, &r.Condition,
			&r.Value, &r.Duration, &r.Severity, &r.Enabled, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

// GetAlertRule returns a single alert rule
func (s *AlertService) GetAlertRule(id uuid.UUID) (*models.AlertRule, error) {
	query := `SELECT id, sensor_id, name, type, condition, value, duration, severity, enabled, created_at, updated_at
		FROM alert_rules WHERE id = $1`

	var r models.AlertRule
	err := s.db.QueryRow(query, id).Scan(&r.ID, &r.SensorID, &r.Name, &r.Type, &r.Condition,
		&r.Value, &r.Duration, &r.Severity, &r.Enabled, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

// CreateAlertRule creates a new alert rule
func (s *AlertService) CreateAlertRule(rule *models.AlertRule) error {
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	query := `INSERT INTO alert_rules (id, sensor_id, name, type, condition, value, duration, severity, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := s.db.Exec(query, rule.ID, rule.SensorID, rule.Name, rule.Type, rule.Condition,
		rule.Value, rule.Duration, rule.Severity, rule.Enabled, rule.CreatedAt, rule.UpdatedAt)
	return err
}

// UpdateAlertRule updates an alert rule
func (s *AlertService) UpdateAlertRule(rule *models.AlertRule) error {
	rule.UpdatedAt = time.Now()
	query := `UPDATE alert_rules SET name=$1, type=$2, condition=$3, value=$4,
		duration=$5, severity=$6, enabled=$7, updated_at=$8 WHERE id=$9`

	_, err := s.db.Exec(query, rule.Name, rule.Type, rule.Condition, rule.Value,
		rule.Duration, rule.Severity, rule.Enabled, rule.UpdatedAt, rule.ID)
	return err
}

// DeleteAlertRule deletes an alert rule
func (s *AlertService) DeleteAlertRule(id uuid.UUID) error {
	_, err := s.db.Exec(`DELETE FROM alert_rules WHERE id = $1`, id)
	return err
}

// ListAlerts returns alerts with optional filtering
func (s *AlertService) ListAlerts(onlyActive bool, limit int) ([]models.Alert, error) {
	var query string
	var args []interface{}

	if onlyActive {
		query = `SELECT a.id, a.rule_id, a.sensor_id, s.name, a.message, a.severity, a.value, a.acknowledged, a.triggered_at, a.acked_at
			FROM alerts a JOIN sensors s ON a.sensor_id = s.id
			WHERE a.acknowledged = false ORDER BY a.triggered_at DESC LIMIT $1`
		args = []interface{}{limit}
	} else {
		query = `SELECT a.id, a.rule_id, a.sensor_id, s.name, a.message, a.severity, a.value, a.acknowledged, a.triggered_at, a.acked_at
			FROM alerts a JOIN sensors s ON a.sensor_id = s.id
			ORDER BY a.triggered_at DESC LIMIT $1`
		args = []interface{}{limit}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var a models.Alert
		if err := rows.Scan(&a.ID, &a.RuleID, &a.SensorID, &a.SensorName, &a.Message,
			&a.Severity, &a.Value, &a.Acknowledged, &a.TriggeredAt, &a.AckedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *AlertService) AcknowledgeAlert(id uuid.UUID) error {
	now := time.Now()
	_, err := s.db.Exec(`UPDATE alerts SET acknowledged = true, acked_at = $1 WHERE id = $2`, now, id)
	return err
}

// AcknowledgeAll marks all alerts as acknowledged
func (s *AlertService) AcknowledgeAll() error {
	now := time.Now()
	_, err := s.db.Exec(`UPDATE alerts SET acknowledged = true, acked_at = $1 WHERE acknowledged = false`, now)
	return err
}

// CreateAlert creates a new alert
func (s *AlertService) CreateAlert(alert *models.Alert) error {
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	alert.TriggeredAt = time.Now()

	query := `INSERT INTO alerts (id, rule_id, sensor_id, message, severity, value, acknowledged, triggered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.Exec(query, alert.ID, alert.RuleID, alert.SensorID, alert.Message,
		alert.Severity, alert.Value, false, alert.TriggeredAt)
	return err
}

// EvaluateRules checks all enabled rules against the latest reading
func (s *AlertService) EvaluateRules(sensorID uuid.UUID, value float64) {
	query := `SELECT id, sensor_id, name, type, condition, value, duration, severity, enabled
		FROM alert_rules WHERE sensor_id = $1 AND enabled = true`

	rows, err := s.db.Query(query, sensorID)
	if err != nil {
		log.Printf("Error querying alert rules: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var rule models.AlertRule
		if err := rows.Scan(&rule.ID, &rule.SensorID, &rule.Name, &rule.Type, &rule.Condition,
			&rule.Value, &rule.Duration, &rule.Severity, &rule.Enabled); err != nil {
			log.Printf("Error scanning rule: %v", err)
			continue
		}

		triggered := false
		var message string

		switch rule.Type {
		case models.AlertRuleThreshold:
			triggered, message = s.evaluateThreshold(rule, value)
		case models.AlertRuleRateOfChange:
			triggered, message = s.evaluateRateOfChange(rule, sensorID, value)
		}

		if triggered {
			// Check for recent duplicate (don't spam alerts)
			var count int
			s.db.QueryRow(`SELECT COUNT(*) FROM alerts WHERE rule_id = $1 AND triggered_at > NOW() - INTERVAL '5 minutes'`,
				rule.ID).Scan(&count)
			if count > 0 {
				continue
			}

			// Look up sensor name for the alert
			var sensorName string
			s.db.QueryRow(`SELECT name FROM sensors WHERE id = $1`, sensorID).Scan(&sensorName)

			alert := &models.Alert{
				RuleID:     rule.ID,
				SensorID:   sensorID,
				SensorName: sensorName,
				Message:    message,
				Severity:   rule.Severity,
				Value:      value,
			}
			if err := s.CreateAlert(alert); err != nil {
				log.Printf("Error creating alert: %v", err)
			} else {
				log.Printf("Alert triggered: %s (value=%.2f)", message, value)
				// Broadcast alert via SSE
				if s.broker != nil {
					s.broker.BroadcastAlert(*alert)
				}
			}
		}
	}
}

func (s *AlertService) evaluateThreshold(rule models.AlertRule, value float64) (bool, string) {
	switch rule.Condition {
	case "above":
		if value > rule.Value {
			return true, fmt.Sprintf("%s: value %.2f exceeds threshold %.2f", rule.Name, value, rule.Value)
		}
	case "below":
		if value < rule.Value {
			return true, fmt.Sprintf("%s: value %.2f below threshold %.2f", rule.Name, value, rule.Value)
		}
	case "equals":
		if math.Abs(value-rule.Value) < 0.01 {
			return true, fmt.Sprintf("%s: value %.2f equals threshold %.2f", rule.Name, value, rule.Value)
		}
	}
	return false, ""
}

func (s *AlertService) evaluateRateOfChange(rule models.AlertRule, sensorID uuid.UUID, currentValue float64) (bool, string) {
	var prevValue float64
	err := s.db.QueryRow(`SELECT value FROM sensor_readings
		WHERE sensor_id = $1 ORDER BY time DESC OFFSET 1 LIMIT 1`, sensorID).Scan(&prevValue)
	if err != nil {
		return false, ""
	}

	rateOfChange := math.Abs(currentValue - prevValue)
	if rateOfChange > rule.Value {
		return true, fmt.Sprintf("%s: rate of change %.2f exceeds limit %.2f", rule.Name, rateOfChange, rule.Value)
	}
	return false, ""
}

// CountActiveAlerts returns the number of unacknowledged alerts
func (s *AlertService) CountActiveAlerts() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM alerts WHERE acknowledged = false`).Scan(&count)
	return count, err
}
