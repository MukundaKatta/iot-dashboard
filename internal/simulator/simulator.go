package simulator

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/iot-dashboard/internal/models"
	"github.com/iot-dashboard/internal/services"
)

// SensorSim holds simulation state for one sensor
type SensorSim struct {
	Sensor       models.Sensor
	BaseValue    float64
	Noise        float64
	DriftRate    float64
	CurrentValue float64
	Phase        float64 // for diurnal cycle
}

// Simulator generates realistic sensor data
type Simulator struct {
	sensors        []*SensorSim
	sensorService  *services.SensorService
	readingService *services.ReadingService
	alertService   *services.AlertService
	broker         *services.SSEBroker
	interval       time.Duration
	stopCh         chan struct{}
}

// NewSimulator creates a new sensor simulator
func NewSimulator(
	sensorSvc *services.SensorService,
	readingSvc *services.ReadingService,
	alertSvc *services.AlertService,
	broker *services.SSEBroker,
	interval time.Duration,
) *Simulator {
	return &Simulator{
		sensorService:  sensorSvc,
		readingService: readingSvc,
		alertService:   alertSvc,
		broker:         broker,
		interval:       interval,
		stopCh:         make(chan struct{}),
	}
}

// DefaultSensors returns a set of predefined sensors for simulation
func DefaultSensors() []SensorSim {
	return []SensorSim{
		{
			Sensor: models.Sensor{
				Name: "Office Temp A", Type: models.SensorTypeTemperature,
				Location: "Building A - Floor 1", Unit: "°C", MinValue: -10, MaxValue: 50,
			},
			BaseValue: 22.0, Noise: 0.5, DriftRate: 0.1,
		},
		{
			Sensor: models.Sensor{
				Name: "Office Temp B", Type: models.SensorTypeTemperature,
				Location: "Building A - Floor 2", Unit: "°C", MinValue: -10, MaxValue: 50,
			},
			BaseValue: 21.5, Noise: 0.4, DriftRate: 0.08,
		},
		{
			Sensor: models.Sensor{
				Name: "Server Room Temp", Type: models.SensorTypeTemperature,
				Location: "Building B - Basement", Unit: "°C", MinValue: 10, MaxValue: 45,
			},
			BaseValue: 19.0, Noise: 0.3, DriftRate: 0.05,
		},
		{
			Sensor: models.Sensor{
				Name: "Office Humidity A", Type: models.SensorTypeHumidity,
				Location: "Building A - Floor 1", Unit: "%", MinValue: 0, MaxValue: 100,
			},
			BaseValue: 45.0, Noise: 2.0, DriftRate: 0.5,
		},
		{
			Sensor: models.Sensor{
				Name: "Server Room Humidity", Type: models.SensorTypeHumidity,
				Location: "Building B - Basement", Unit: "%", MinValue: 0, MaxValue: 100,
			},
			BaseValue: 40.0, Noise: 1.5, DriftRate: 0.3,
		},
		{
			Sensor: models.Sensor{
				Name: "Barometric Pressure", Type: models.SensorTypePressure,
				Location: "Building A - Roof", Unit: "hPa", MinValue: 950, MaxValue: 1050,
			},
			BaseValue: 1013.25, Noise: 0.5, DriftRate: 0.2,
		},
		{
			Sensor: models.Sensor{
				Name: "Office CO2", Type: models.SensorTypeCO2,
				Location: "Building A - Floor 1", Unit: "ppm", MinValue: 300, MaxValue: 2000,
			},
			BaseValue: 450.0, Noise: 20.0, DriftRate: 5.0,
		},
		{
			Sensor: models.Sensor{
				Name: "Conference Room CO2", Type: models.SensorTypeCO2,
				Location: "Building A - Floor 2", Unit: "ppm", MinValue: 300, MaxValue: 2000,
			},
			BaseValue: 500.0, Noise: 30.0, DriftRate: 8.0,
		},
		{
			Sensor: models.Sensor{
				Name: "Office Light Level", Type: models.SensorTypeLight,
				Location: "Building A - Floor 1", Unit: "lux", MinValue: 0, MaxValue: 10000,
			},
			BaseValue: 350.0, Noise: 20.0, DriftRate: 10.0,
		},
		{
			Sensor: models.Sensor{
				Name: "Outdoor Light", Type: models.SensorTypeLight,
				Location: "Building A - Roof", Unit: "lux", MinValue: 0, MaxValue: 100000,
			},
			BaseValue: 5000.0, Noise: 500.0, DriftRate: 200.0,
		},
	}
}

// DefaultAlertRules returns predefined alert rules for a sensor
func DefaultAlertRules(sensorID uuid.UUID, sensorType models.SensorType) []models.AlertRule {
	switch sensorType {
	case models.SensorTypeTemperature:
		return []models.AlertRule{
			{SensorID: sensorID, Name: "High Temperature", Type: models.AlertRuleThreshold, Condition: "above", Value: 30, Severity: models.AlertSeverityWarning, Enabled: true},
			{SensorID: sensorID, Name: "Critical Temperature", Type: models.AlertRuleThreshold, Condition: "above", Value: 35, Severity: models.AlertSeverityCritical, Enabled: true},
			{SensorID: sensorID, Name: "Low Temperature", Type: models.AlertRuleThreshold, Condition: "below", Value: 15, Severity: models.AlertSeverityWarning, Enabled: true},
			{SensorID: sensorID, Name: "Temp Spike", Type: models.AlertRuleRateOfChange, Condition: "above", Value: 3.0, Severity: models.AlertSeverityWarning, Enabled: true},
		}
	case models.SensorTypeHumidity:
		return []models.AlertRule{
			{SensorID: sensorID, Name: "High Humidity", Type: models.AlertRuleThreshold, Condition: "above", Value: 70, Severity: models.AlertSeverityWarning, Enabled: true},
			{SensorID: sensorID, Name: "Low Humidity", Type: models.AlertRuleThreshold, Condition: "below", Value: 25, Severity: models.AlertSeverityWarning, Enabled: true},
		}
	case models.SensorTypeCO2:
		return []models.AlertRule{
			{SensorID: sensorID, Name: "High CO2", Type: models.AlertRuleThreshold, Condition: "above", Value: 1000, Severity: models.AlertSeverityWarning, Enabled: true},
			{SensorID: sensorID, Name: "Dangerous CO2", Type: models.AlertRuleThreshold, Condition: "above", Value: 1500, Severity: models.AlertSeverityCritical, Enabled: true},
		}
	default:
		return nil
	}
}

// Setup initializes sensors in the database and generates historical data
func (s *Simulator) Setup() error {
	log.Println("Setting up simulator...")

	// Check if sensors already exist
	existing, err := s.sensorService.ListSensors()
	if err != nil {
		return fmt.Errorf("list sensors: %w", err)
	}
	if len(existing) > 0 {
		log.Printf("Found %d existing sensors, loading them", len(existing))
		for i := range existing {
			sim := &SensorSim{
				Sensor:    existing[i],
				BaseValue: s.getBaseForType(existing[i].Type),
				Noise:     s.getNoiseForType(existing[i].Type),
				DriftRate: 0.1,
			}
			if existing[i].LastReading != nil {
				sim.CurrentValue = *existing[i].LastReading
			} else {
				sim.CurrentValue = sim.BaseValue
			}
			s.sensors = append(s.sensors, sim)
		}
		return nil
	}

	// Create default sensors
	defaults := DefaultSensors()
	for i := range defaults {
		sim := &defaults[i]
		sim.Sensor.ID = uuid.New()
		sim.CurrentValue = sim.BaseValue
		sim.Phase = rand.Float64() * 2 * math.Pi

		if err := s.sensorService.CreateSensor(&sim.Sensor); err != nil {
			return fmt.Errorf("create sensor %s: %w", sim.Sensor.Name, err)
		}

		// Create default alert rules
		rules := DefaultAlertRules(sim.Sensor.ID, sim.Sensor.Type)
		for _, rule := range rules {
			if err := s.alertService.CreateAlertRule(&rule); err != nil {
				log.Printf("Warning: failed to create alert rule: %v", err)
			}
		}

		s.sensors = append(s.sensors, sim)
		log.Printf("Created sensor: %s (%s)", sim.Sensor.Name, sim.Sensor.ID)
	}

	// Generate 24 hours of historical data
	log.Println("Generating 24h of historical data...")
	if err := s.generateHistoricalData(24 * time.Hour); err != nil {
		log.Printf("Warning: failed to generate historical data: %v", err)
	}

	return nil
}

func (s *Simulator) getBaseForType(t models.SensorType) float64 {
	switch t {
	case models.SensorTypeTemperature:
		return 22.0
	case models.SensorTypeHumidity:
		return 45.0
	case models.SensorTypePressure:
		return 1013.25
	case models.SensorTypeCO2:
		return 450.0
	case models.SensorTypeLight:
		return 350.0
	default:
		return 50.0
	}
}

func (s *Simulator) getNoiseForType(t models.SensorType) float64 {
	switch t {
	case models.SensorTypeTemperature:
		return 0.5
	case models.SensorTypeHumidity:
		return 2.0
	case models.SensorTypePressure:
		return 0.5
	case models.SensorTypeCO2:
		return 20.0
	case models.SensorTypeLight:
		return 20.0
	default:
		return 1.0
	}
}

func (s *Simulator) generateHistoricalData(duration time.Duration) error {
	now := time.Now()
	start := now.Add(-duration)
	step := 30 * time.Second // one reading every 30 seconds

	batch := make([]models.SensorReading, 0, 1000)

	for t := start; t.Before(now); t = t.Add(step) {
		for _, sim := range s.sensors {
			value := s.generateValue(sim, t)
			batch = append(batch, models.SensorReading{
				Time:     t,
				SensorID: sim.Sensor.ID,
				Value:    value,
				Quality:  90 + rand.Intn(11),
			})

			if len(batch) >= 1000 {
				if err := s.readingService.InsertBatchReadings(batch); err != nil {
					return fmt.Errorf("insert batch: %w", err)
				}
				batch = batch[:0]
			}
		}
	}

	if len(batch) > 0 {
		if err := s.readingService.InsertBatchReadings(batch); err != nil {
			return fmt.Errorf("insert final batch: %w", err)
		}
	}

	// Set all sensors to online with last value
	for _, sim := range s.sensors {
		val := sim.CurrentValue
		s.sensorService.UpdateSensorStatus(sim.Sensor.ID, models.SensorStatusOnline, &val)
	}

	log.Printf("Generated historical data: %d sensors x %.0f hours", len(s.sensors), duration.Hours())
	return nil
}

// generateValue creates a realistic sensor value for a given time
func (s *Simulator) generateValue(sim *SensorSim, t time.Time) float64 {
	hourOfDay := float64(t.Hour()) + float64(t.Minute())/60.0

	var value float64

	switch sim.Sensor.Type {
	case models.SensorTypeTemperature:
		// Diurnal cycle: warmer during day, cooler at night
		diurnal := 3.0 * math.Sin((hourOfDay-6)*math.Pi/12)
		// Seasonal component (simplified)
		dayOfYear := float64(t.YearDay())
		seasonal := 2.0 * math.Sin((dayOfYear-80)*2*math.Pi/365)
		value = sim.BaseValue + diurnal + seasonal + sim.Noise*rand.NormFloat64()

	case models.SensorTypeHumidity:
		// Inverse of temperature - higher humidity at night
		diurnal := -5.0 * math.Sin((hourOfDay-6)*math.Pi/12)
		value = sim.BaseValue + diurnal + sim.Noise*rand.NormFloat64()
		value = math.Max(10, math.Min(95, value))

	case models.SensorTypePressure:
		// Slow drift with small noise
		drift := 5.0 * math.Sin(float64(t.Unix())/86400*2*math.Pi)
		value = sim.BaseValue + drift + sim.Noise*rand.NormFloat64()

	case models.SensorTypeCO2:
		// Higher during work hours (9-17)
		var occupancy float64
		if hourOfDay >= 9 && hourOfDay <= 17 {
			occupancy = 200 * math.Sin((hourOfDay-9)*math.Pi/8)
		}
		// Weekend = lower
		if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
			occupancy *= 0.2
		}
		value = sim.BaseValue + occupancy + sim.Noise*rand.NormFloat64()
		value = math.Max(350, value)

	case models.SensorTypeLight:
		// Strong diurnal - dark at night, bright during day
		if hourOfDay >= 7 && hourOfDay <= 19 {
			daylight := math.Sin((hourOfDay - 7) * math.Pi / 12)
			value = sim.BaseValue * daylight * 2
		} else {
			value = sim.BaseValue * 0.05
		}
		// Cloud cover simulation
		value *= 0.7 + 0.3*rand.Float64()
		value += sim.Noise * rand.NormFloat64()
		value = math.Max(0, value)

	default:
		value = sim.BaseValue + sim.Noise*rand.NormFloat64()
	}

	// Random anomalies (0.5% chance)
	if rand.Float64() < 0.005 {
		anomalyMagnitude := (sim.Sensor.MaxValue - sim.Sensor.MinValue) * 0.15
		if rand.Float64() > 0.5 {
			value += anomalyMagnitude
		} else {
			value -= anomalyMagnitude
		}
	}

	// Clamp to sensor range
	value = math.Max(sim.Sensor.MinValue, math.Min(sim.Sensor.MaxValue, value))
	sim.CurrentValue = value

	return math.Round(value*100) / 100
}

// Start begins the simulation loop
func (s *Simulator) Start() {
	log.Printf("Starting simulator with %d sensors (interval: %v)", len(s.sensors), s.interval)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			log.Println("Simulator stopped")
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *Simulator) tick() {
	now := time.Now()

	for _, sim := range s.sensors {
		// Simulate occasional sensor offline (0.1% chance per tick)
		if rand.Float64() < 0.001 {
			s.sensorService.UpdateSensorStatus(sim.Sensor.ID, models.SensorStatusOffline, nil)
			sim.Sensor.Status = models.SensorStatusOffline
			s.broker.BroadcastSensorStatus(sim.Sensor)
			continue
		}

		// Bring offline sensors back online (10% chance per tick)
		if sim.Sensor.Status == models.SensorStatusOffline && rand.Float64() < 0.1 {
			sim.Sensor.Status = models.SensorStatusOnline
		}

		if sim.Sensor.Status == models.SensorStatusOffline {
			continue
		}

		value := s.generateValue(sim, now)
		quality := 85 + rand.Intn(16)

		reading := models.SensorReading{
			Time:     now,
			SensorID: sim.Sensor.ID,
			Value:    value,
			Quality:  quality,
		}

		if err := s.readingService.InsertReading(&reading); err != nil {
			log.Printf("Error inserting reading for %s: %v", sim.Sensor.Name, err)
			continue
		}

		s.sensorService.UpdateSensorStatus(sim.Sensor.ID, models.SensorStatusOnline, &value)
		sim.Sensor.Status = models.SensorStatusOnline

		// Broadcast to SSE clients
		s.broker.BroadcastReading(reading)

		// Evaluate alert rules
		s.alertService.EvaluateRules(sim.Sensor.ID, value)
	}
}

// Stop stops the simulator
func (s *Simulator) Stop() {
	close(s.stopCh)
}

// GetSensors returns the list of simulated sensors
func (s *Simulator) GetSensors() []*SensorSim {
	return s.sensors
}
