package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/iot-dashboard/internal/handlers"
	"github.com/iot-dashboard/internal/middleware"
	"github.com/iot-dashboard/internal/services"
	"github.com/iot-dashboard/internal/simulator"
	"github.com/iot-dashboard/internal/templates"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting IoT Dashboard Server...")

	// Configuration from environment
	port := envOrDefault("PORT", "8080")
	dbHost := envOrDefault("DB_HOST", "localhost")
	dbPort := envOrDefaultInt("DB_PORT", 5432)
	dbUser := envOrDefault("DB_USER", "iot")
	dbPass := envOrDefault("DB_PASSWORD", "iotpassword")
	dbName := envOrDefault("DB_NAME", "iot_dashboard")
	simInterval := envOrDefaultDuration("SIM_INTERVAL", 5*time.Second)
	enableSim := envOrDefault("ENABLE_SIMULATOR", "true")

	// Database connection
	db, err := services.NewDB(services.DBConfig{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPass,
		DBName:   dbName,
		SSLMode:  "disable",
	})
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := services.RunMigrations(db); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	// Initialize services
	sensorSvc := services.NewSensorService(db)
	readingSvc := services.NewReadingService(db)
	alertSvc := services.NewAlertService(db)
	groupSvc := services.NewGroupService(db)
	broker := services.NewSSEBroker()

	// Wire the broker into alert service for SSE broadcasting
	alertSvc.SetBroker(broker)

	// Template engine
	tmpl := templates.NewEngine()
	if err := tmpl.LoadTemplates(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Handler
	h := handlers.NewHandler(sensorSvc, readingSvc, alertSvc, groupSvc, broker, tmpl)

	// Router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(chimw.RealIP)
	r.Use(chimw.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "HX-Request", "HX-Target", "HX-Trigger"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Pages
	r.Get("/", h.Dashboard)
	r.Get("/sensors", h.SensorsList)
	r.Get("/sensors/{id}", h.SensorDetail)
	r.Get("/alerts", h.AlertsPage)
	r.Get("/settings", h.SettingsPage)

	// Partials (HTMX)
	r.Get("/partials/stats", h.PartialsStats)
	r.Get("/partials/sensors", h.PartialsSensors)
	r.Get("/partials/alerts", h.PartialsAlerts)

	// API
	r.Route("/api", func(r chi.Router) {
		// SSE
		r.Get("/events", broker.ServeHTTP)

		// Sensors
		r.Post("/sensors", h.APISensorCreate)
		r.Delete("/sensors/{id}", h.APISensorDelete)
		r.Get("/sensors/{id}/sparkline", h.APISensorSparkline)
		r.Get("/sensors/{id}/chart", h.APISensorChart)
		r.Get("/sensors/{id}/readings", h.APIReadings)
		r.Get("/sensors/{id}/rules", h.APISensorRules)

		// Readings ingestion
		r.Post("/readings", h.APIIngestReading)
		r.Post("/readings/batch", h.APIIngestBatch)

		// Alerts
		r.Get("/alerts/count", h.APIAlertsCount)
		r.Post("/alerts/{id}/acknowledge", h.APIAcknowledgeAlert)
		r.Post("/alerts/acknowledge-all", h.APIAcknowledgeAll)

		// Alert rules
		r.Get("/alert-rules", h.APIAlertRules)
		r.Delete("/alert-rules/{id}", h.APIDeleteAlertRule)

		// Groups
		r.Get("/groups", h.APIGroups)
		r.Post("/groups", h.APICreateGroup)
		r.Delete("/groups/{id}", h.APIDeleteGroup)

		// Charts
		r.Get("/charts/temperature", h.APIChartTemperature)
		r.Get("/charts/humidity", h.APIChartHumidity)

		// System
		r.Get("/system/sse-clients", h.APISSEClients)
	})

	// Start simulator
	if enableSim == "true" {
		sim := simulator.NewSimulator(sensorSvc, readingSvc, alertSvc, broker, simInterval)
		if err := sim.Setup(); err != nil {
			log.Fatalf("Simulator setup failed: %v", err)
		}
		go sim.Start()
		log.Println("Simulator started")
	}

	// Start offline detection loop
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			offlineSensors, err := sensorSvc.DetectOfflineSensors(2 * time.Minute)
			if err != nil {
				log.Printf("Offline detection error: %v", err)
				continue
			}
			for _, s := range offlineSensors {
				broker.BroadcastSensorStatus(s)
				log.Printf("Sensor %s marked offline", s.Name)
			}
		}
	}()

	// HTTP Server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second, // Longer for SSE
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Printf("Server listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server stopped")
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envOrDefaultInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func envOrDefaultDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
