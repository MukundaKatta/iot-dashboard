package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iot-dashboard/internal/models"
)

// GroupService handles sensor group operations
type GroupService struct {
	db *sql.DB
}

// NewGroupService creates a new GroupService
func NewGroupService(db *sql.DB) *GroupService {
	return &GroupService{db: db}
}

// ListGroups returns all sensor groups
func (s *GroupService) ListGroups() ([]models.SensorGroup, error) {
	rows, err := s.db.Query(`SELECT id, name, description, created_at FROM sensor_groups ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	var groups []models.SensorGroup
	for rows.Next() {
		var g models.SensorGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// CreateGroup creates a new sensor group
func (s *GroupService) CreateGroup(group *models.SensorGroup) error {
	if group.ID == uuid.Nil {
		group.ID = uuid.New()
	}
	group.CreatedAt = time.Now()

	_, err := s.db.Exec(`INSERT INTO sensor_groups (id, name, description, created_at) VALUES ($1, $2, $3, $4)`,
		group.ID, group.Name, group.Description, group.CreatedAt)
	return err
}

// DeleteGroup deletes a sensor group
func (s *GroupService) DeleteGroup(id uuid.UUID) error {
	_, err := s.db.Exec(`DELETE FROM sensor_groups WHERE id = $1`, id)
	return err
}
