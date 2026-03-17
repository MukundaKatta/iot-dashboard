package simulator

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iot-dashboard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSensors(t *testing.T) {
	sensors := DefaultSensors()
	assert.Len(t, sensors, 10)

	// Check that we have a variety of types
	typeCount := make(map[models.SensorType]int)
	for _, s := range sensors {
		typeCount[s.Sensor.Type]++
	}
	assert.Equal(t, 3, typeCount[models.SensorTypeTemperature])
	assert.Equal(t, 2, typeCount[models.SensorTypeHumidity])
	assert.Equal(t, 1, typeCount[models.SensorTypePressure])
	assert.Equal(t, 2, typeCount[models.SensorTypeCO2])
	assert.Equal(t, 2, typeCount[models.SensorTypeLight])
}

func TestDefaultAlertRules(t *testing.T) {
	sensorID := uuid.New()

	// Temperature rules
	rules := DefaultAlertRules(sensorID, models.SensorTypeTemperature)
	assert.Len(t, rules, 4)
	assert.Equal(t, sensorID, rules[0].SensorID)

	// Humidity rules
	rules = DefaultAlertRules(sensorID, models.SensorTypeHumidity)
	assert.Len(t, rules, 2)

	// CO2 rules
	rules = DefaultAlertRules(sensorID, models.SensorTypeCO2)
	assert.Len(t, rules, 2)

	// Light has no default rules
	rules = DefaultAlertRules(sensorID, models.SensorTypeLight)
	assert.Nil(t, rules)
}

func TestGenerateValue_Temperature(t *testing.T) {
	s := &Simulator{}
	sim := &SensorSim{
		Sensor: models.Sensor{
			Type:     models.SensorTypeTemperature,
			MinValue: -10,
			MaxValue: 50,
		},
		BaseValue:    22.0,
		Noise:        0.5,
		DriftRate:    0.1,
		CurrentValue: 22.0,
	}

	// Generate values at different times of day
	morningTime := time.Date(2024, 6, 15, 6, 0, 0, 0, time.UTC)
	noonTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	nightTime := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	// Sample multiple values to get averages
	morningAvg := sampleAvg(s, sim, morningTime, 100)
	noonAvg := sampleAvg(s, sim, noonTime, 100)
	nightAvg := sampleAvg(s, sim, nightTime, 100)

	// Noon should be warmer than night (diurnal cycle)
	assert.Greater(t, noonAvg, nightAvg, "Noon should be warmer than midnight")
	// Morning should be between night and noon
	assert.Greater(t, noonAvg, morningAvg, "Noon should be warmer than morning")

	// All values should be within sensor range
	for i := 0; i < 1000; i++ {
		v := s.generateValue(sim, noonTime.Add(time.Duration(i)*time.Minute))
		assert.GreaterOrEqual(t, v, sim.Sensor.MinValue)
		assert.LessOrEqual(t, v, sim.Sensor.MaxValue)
	}
}

func TestGenerateValue_Humidity(t *testing.T) {
	s := &Simulator{}
	sim := &SensorSim{
		Sensor: models.Sensor{
			Type:     models.SensorTypeHumidity,
			MinValue: 0,
			MaxValue: 100,
		},
		BaseValue: 45.0,
		Noise:     2.0,
	}

	for i := 0; i < 100; i++ {
		v := s.generateValue(sim, time.Now())
		assert.GreaterOrEqual(t, v, sim.Sensor.MinValue)
		assert.LessOrEqual(t, v, sim.Sensor.MaxValue)
	}
}

func TestGenerateValue_CO2_WorkHours(t *testing.T) {
	s := &Simulator{}
	sim := &SensorSim{
		Sensor: models.Sensor{
			Type:     models.SensorTypeCO2,
			MinValue: 300,
			MaxValue: 2000,
		},
		BaseValue: 450.0,
		Noise:     20.0,
	}

	// Weekday work hours should have higher CO2
	weekdayNoon := time.Date(2024, 6, 12, 12, 0, 0, 0, time.UTC) // Wednesday
	weekdayNight := time.Date(2024, 6, 12, 2, 0, 0, 0, time.UTC)
	weekendNoon := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC) // Saturday

	workAvg := sampleAvg(s, sim, weekdayNoon, 200)
	nightAvg := sampleAvg(s, sim, weekdayNight, 200)
	weekendAvg := sampleAvg(s, sim, weekendNoon, 200)

	assert.Greater(t, workAvg, nightAvg, "Weekday work hours should have higher CO2")
	assert.Greater(t, workAvg, weekendAvg, "Weekday should have higher CO2 than weekend")
}

func TestGenerateValue_Light_DayNight(t *testing.T) {
	s := &Simulator{}
	sim := &SensorSim{
		Sensor: models.Sensor{
			Type:     models.SensorTypeLight,
			MinValue: 0,
			MaxValue: 100000,
		},
		BaseValue: 5000.0,
		Noise:     500.0,
	}

	dayTime := time.Date(2024, 6, 15, 13, 0, 0, 0, time.UTC)
	nightTime := time.Date(2024, 6, 15, 2, 0, 0, 0, time.UTC)

	dayAvg := sampleAvg(s, sim, dayTime, 100)
	nightAvg := sampleAvg(s, sim, nightTime, 100)

	assert.Greater(t, dayAvg, nightAvg*5, "Daytime light should be much higher than nighttime")
}

func TestGenerateValue_Anomalies(t *testing.T) {
	s := &Simulator{}
	sim := &SensorSim{
		Sensor: models.Sensor{
			Type:     models.SensorTypeTemperature,
			MinValue: -10,
			MaxValue: 50,
		},
		BaseValue: 22.0,
		Noise:     0.5,
	}

	// Generate many values and check none exceed bounds even with anomalies
	now := time.Now()
	for i := 0; i < 10000; i++ {
		v := s.generateValue(sim, now.Add(time.Duration(i)*time.Second))
		require.GreaterOrEqual(t, v, sim.Sensor.MinValue, "Value should not go below min")
		require.LessOrEqual(t, v, sim.Sensor.MaxValue, "Value should not exceed max")
	}
}

func sampleAvg(s *Simulator, sim *SensorSim, t time.Time, n int) float64 {
	total := 0.0
	for i := 0; i < n; i++ {
		total += s.generateValue(sim, t.Add(time.Duration(i)*time.Second))
	}
	return math.Round(total/float64(n)*100) / 100
}
