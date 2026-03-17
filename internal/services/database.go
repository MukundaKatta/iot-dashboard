package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// DBConfig holds database configuration
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDB creates a new database connection
func NewDB(cfg DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Retry connection with backoff
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("Database connected successfully")
			return db, nil
		}
		log.Printf("Waiting for database... attempt %d/30: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after 30 attempts: %w", err)
}

// RunMigrations executes the database schema setup
func RunMigrations(db *sql.DB) error {
	migrations := []string{
		// Enable TimescaleDB
		`CREATE EXTENSION IF NOT EXISTS timescaledb;`,

		// Sensor groups
		`CREATE TABLE IF NOT EXISTS sensor_groups (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,

		// Sensors
		`CREATE TABLE IF NOT EXISTS sensors (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			location VARCHAR(255) DEFAULT '',
			group_id UUID REFERENCES sensor_groups(id) ON DELETE SET NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'offline',
			unit VARCHAR(20) NOT NULL DEFAULT '',
			min_value DOUBLE PRECISION NOT NULL DEFAULT 0,
			max_value DOUBLE PRECISION NOT NULL DEFAULT 100,
			last_reading DOUBLE PRECISION,
			last_seen_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,

		// Sensor readings - hypertable
		`CREATE TABLE IF NOT EXISTS sensor_readings (
			time TIMESTAMPTZ NOT NULL,
			sensor_id UUID NOT NULL REFERENCES sensors(id) ON DELETE CASCADE,
			value DOUBLE PRECISION NOT NULL,
			quality INT NOT NULL DEFAULT 100
		);`,

		// Convert to hypertable (idempotent check)
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM timescaledb_information.hypertables
				WHERE hypertable_name = 'sensor_readings'
			) THEN
				PERFORM create_hypertable('sensor_readings', 'time');
			END IF;
		END $$;`,

		// Index on sensor_readings
		`CREATE INDEX IF NOT EXISTS idx_sensor_readings_sensor_time
			ON sensor_readings (sensor_id, time DESC);`,

		// Alert rules
		`CREATE TABLE IF NOT EXISTS alert_rules (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			sensor_id UUID NOT NULL REFERENCES sensors(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'threshold',
			condition VARCHAR(20) NOT NULL DEFAULT 'above',
			value DOUBLE PRECISION NOT NULL DEFAULT 0,
			duration INT NOT NULL DEFAULT 0,
			severity VARCHAR(20) NOT NULL DEFAULT 'warning',
			enabled BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,

		// Alerts
		`CREATE TABLE IF NOT EXISTS alerts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			rule_id UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
			sensor_id UUID NOT NULL REFERENCES sensors(id) ON DELETE CASCADE,
			message TEXT NOT NULL,
			severity VARCHAR(20) NOT NULL DEFAULT 'warning',
			value DOUBLE PRECISION NOT NULL DEFAULT 0,
			acknowledged BOOLEAN NOT NULL DEFAULT false,
			triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			acked_at TIMESTAMPTZ
		);`,

		// Index on alerts
		`CREATE INDEX IF NOT EXISTS idx_alerts_sensor ON alerts (sensor_id, triggered_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_unacked ON alerts (acknowledged, triggered_at DESC);`,

		// Continuous aggregate: hourly
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM timescaledb_information.continuous_aggregates
				WHERE view_name = 'sensor_readings_hourly'
			) THEN
				EXECUTE $agg$
				CREATE MATERIALIZED VIEW sensor_readings_hourly
				WITH (timescaledb.continuous) AS
				SELECT
					time_bucket('1 hour', time) AS bucket,
					sensor_id,
					AVG(value) AS avg_value,
					MIN(value) AS min_value,
					MAX(value) AS max_value,
					COUNT(*) AS count
				FROM sensor_readings
				GROUP BY bucket, sensor_id
				WITH NO DATA
				$agg$;
			END IF;
		END $$;`,

		// Continuous aggregate: daily
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM timescaledb_information.continuous_aggregates
				WHERE view_name = 'sensor_readings_daily'
			) THEN
				EXECUTE $agg$
				CREATE MATERIALIZED VIEW sensor_readings_daily
				WITH (timescaledb.continuous) AS
				SELECT
					time_bucket('1 day', time) AS bucket,
					sensor_id,
					AVG(value) AS avg_value,
					MIN(value) AS min_value,
					MAX(value) AS max_value,
					COUNT(*) AS count
				FROM sensor_readings
				GROUP BY bucket, sensor_id
				WITH NO DATA
				$agg$;
			END IF;
		END $$;`,

		// Refresh policies
		`DO $$
		BEGIN
			BEGIN
				PERFORM add_continuous_aggregate_policy('sensor_readings_hourly',
					start_offset => INTERVAL '3 hours',
					end_offset => INTERVAL '1 hour',
					schedule_interval => INTERVAL '1 hour');
			EXCEPTION WHEN OTHERS THEN NULL;
			END;
		END $$;`,

		`DO $$
		BEGIN
			BEGIN
				PERFORM add_continuous_aggregate_policy('sensor_readings_daily',
					start_offset => INTERVAL '3 days',
					end_offset => INTERVAL '1 day',
					schedule_interval => INTERVAL '1 day');
			EXCEPTION WHEN OTHERS THEN NULL;
			END;
		END $$;`,

		// Retention policy: keep raw data for 30 days
		`DO $$
		BEGIN
			BEGIN
				PERFORM add_retention_policy('sensor_readings', INTERVAL '30 days');
			EXCEPTION WHEN OTHERS THEN NULL;
			END;
		END $$;`,
	}

	for i, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	log.Println("Migrations completed successfully")
	return nil
}
