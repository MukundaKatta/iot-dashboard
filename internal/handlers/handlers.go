package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/iot-dashboard/internal/models"
	"github.com/iot-dashboard/internal/services"
	"github.com/iot-dashboard/internal/templates"
)

// Handler holds all HTTP handler dependencies
type Handler struct {
	sensorSvc  *services.SensorService
	readingSvc *services.ReadingService
	alertSvc   *services.AlertService
	groupSvc   *services.GroupService
	broker     *services.SSEBroker
	tmpl       *templates.Engine
}

// NewHandler creates a new Handler
func NewHandler(
	sensorSvc *services.SensorService,
	readingSvc *services.ReadingService,
	alertSvc *services.AlertService,
	groupSvc *services.GroupService,
	broker *services.SSEBroker,
	tmpl *templates.Engine,
) *Handler {
	return &Handler{
		sensorSvc:  sensorSvc,
		readingSvc: readingSvc,
		alertSvc:   alertSvc,
		groupSvc:   groupSvc,
		broker:     broker,
		tmpl:       tmpl,
	}
}

// isHTMX checks if the request is an HTMX request
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	http.Error(w, msg, status)
}

// --- Page Handlers ---

// Dashboard renders the main dashboard
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	sensors, err := h.sensorSvc.ListSensors()
	if err != nil {
		log.Printf("Error listing sensors: %v", err)
		sensors = []models.Sensor{}
	}

	stats, err := h.sensorSvc.GetDashboardStats()
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		stats = &models.DashboardStats{}
	}

	alerts, err := h.alertSvc.ListAlerts(true, 5)
	if err != nil {
		log.Printf("Error listing alerts: %v", err)
		alerts = []models.Alert{}
	}

	data := map[string]interface{}{
		"ActivePage":   "dashboard",
		"Sensors":      sensors,
		"Stats":        stats,
		"RecentAlerts": alerts,
	}

	if isHTMX(r) {
		// Return just the content for HTMX navigation
		h.renderPageContent(w, "dashboard", data)
		return
	}

	if err := h.tmpl.Render(w, "dashboard", data); err != nil {
		log.Printf("Error rendering dashboard: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to render page")
	}
}

// SensorsList renders the sensors list page
func (h *Handler) SensorsList(w http.ResponseWriter, r *http.Request) {
	sensors, err := h.sensorSvc.ListSensors()
	if err != nil {
		log.Printf("Error listing sensors: %v", err)
		sensors = []models.Sensor{}
	}

	data := map[string]interface{}{
		"ActivePage": "sensors",
		"Sensors":    sensors,
	}

	if isHTMX(r) {
		h.renderPageContent(w, "sensors", data)
		return
	}

	if err := h.tmpl.Render(w, "sensors", data); err != nil {
		log.Printf("Error rendering sensors: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to render page")
	}
}

// SensorDetail renders a sensor's detail page
func (h *Handler) SensorDetail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid sensor ID")
		return
	}

	sensor, err := h.sensorSvc.GetSensor(id)
	if err != nil || sensor == nil {
		writeError(w, http.StatusNotFound, "Sensor not found")
		return
	}

	data := map[string]interface{}{
		"ActivePage": "sensors",
		"Sensor":     sensor,
	}

	if isHTMX(r) {
		h.renderPageContent(w, "sensor_detail", data)
		return
	}

	if err := h.tmpl.Render(w, "sensor_detail", data); err != nil {
		log.Printf("Error rendering sensor detail: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to render page")
	}
}

// AlertsPage renders the alerts page
func (h *Handler) AlertsPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"ActivePage": "alerts",
	}

	if isHTMX(r) {
		h.renderPageContent(w, "alerts", data)
		return
	}

	if err := h.tmpl.Render(w, "alerts", data); err != nil {
		log.Printf("Error rendering alerts: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to render page")
	}
}

// SettingsPage renders the settings page
func (h *Handler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"ActivePage": "settings",
	}

	if isHTMX(r) {
		h.renderPageContent(w, "settings", data)
		return
	}

	if err := h.tmpl.Render(w, "settings", data); err != nil {
		log.Printf("Error rendering settings: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to render page")
	}
}

// renderPageContent renders just the page content for HTMX navigation
func (h *Handler) renderPageContent(w http.ResponseWriter, page string, data map[string]interface{}) {
	if err := h.tmpl.RenderContent(w, page, data); err != nil {
		log.Printf("Error rendering page content %s: %v", page, err)
		writeError(w, http.StatusInternalServerError, "Failed to render")
	}
}

// --- Partial Handlers ---

// PartialsStats returns the stats bar partial
func (h *Handler) PartialsStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.sensorSvc.GetDashboardStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get stats")
		return
	}
	data := map[string]interface{}{"Stats": stats}
	h.tmpl.RenderPartial(w, "stats_bar_partial", data)
}

// PartialsSensors returns the sensor list partial
func (h *Handler) PartialsSensors(w http.ResponseWriter, r *http.Request) {
	sensors, err := h.sensorSvc.ListSensors()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list sensors")
		return
	}

	// Filter by type if specified
	sensorType := r.URL.Query().Get("type")
	if sensorType != "" {
		filtered := make([]models.Sensor, 0)
		for _, s := range sensors {
			if string(s.Type) == sensorType {
				filtered = append(filtered, s)
			}
		}
		sensors = filtered
	}

	data := map[string]interface{}{"Sensors": sensors}
	h.tmpl.RenderPartial(w, "sensor_list_partial", data)
}

// PartialsAlerts returns the alert list partial
func (h *Handler) PartialsAlerts(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") != "false"
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	alerts, err := h.alertSvc.ListAlerts(activeOnly, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list alerts")
		return
	}

	data := map[string]interface{}{"Alerts": alerts}
	h.tmpl.RenderPartial(w, "alert_list_partial", data)
}

// --- API Handlers ---

// APISensorCreate creates a new sensor via API
func (h *Handler) APISensorCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	minVal, _ := strconv.ParseFloat(r.FormValue("min_value"), 64)
	maxVal, _ := strconv.ParseFloat(r.FormValue("max_value"), 64)

	sensor := &models.Sensor{
		Name:     r.FormValue("name"),
		Type:     models.SensorType(r.FormValue("type")),
		Location: r.FormValue("location"),
		Unit:     r.FormValue("unit"),
		MinValue: minVal,
		MaxValue: maxVal,
	}

	if err := h.sensorSvc.CreateSensor(sensor); err != nil {
		log.Printf("Error creating sensor: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create sensor")
		return
	}

	// Return updated sensor list
	h.PartialsSensors(w, r)
}

// APISensorDelete deletes a sensor
func (h *Handler) APISensorDelete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid sensor ID")
		return
	}

	if err := h.sensorSvc.DeleteSensor(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete sensor")
		return
	}

	// Redirect to sensors list
	h.SensorsList(w, r)
}

// APIIngestReading ingests a sensor reading (for real IoT devices)
func (h *Handler) APIIngestReading(w http.ResponseWriter, r *http.Request) {
	var req models.ReadingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	sensorID, err := uuid.Parse(req.SensorID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid sensor_id"})
		return
	}

	reading := &models.SensorReading{
		Time:     time.Now(),
		SensorID: sensorID,
		Value:    req.Value,
		Quality:  req.Quality,
	}

	if err := h.readingSvc.InsertReading(reading); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to insert reading"})
		return
	}

	h.sensorSvc.UpdateSensorStatus(sensorID, models.SensorStatusOnline, &req.Value)
	h.broker.BroadcastReading(*reading)
	h.alertSvc.EvaluateRules(sensorID, req.Value)

	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// APIIngestBatch ingests multiple readings
func (h *Handler) APIIngestBatch(w http.ResponseWriter, r *http.Request) {
	var reqs []models.ReadingRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	readings := make([]models.SensorReading, 0, len(reqs))
	now := time.Now()
	for _, req := range reqs {
		sensorID, err := uuid.Parse(req.SensorID)
		if err != nil {
			continue
		}
		readings = append(readings, models.SensorReading{
			Time:     now,
			SensorID: sensorID,
			Value:    req.Value,
			Quality:  req.Quality,
		})
	}

	if err := h.readingSvc.InsertBatchReadings(readings); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to insert readings"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]int{"inserted": len(readings)})
}

// APISensorSparkline returns sparkline SVG for a sensor
func (h *Handler) APISensorSparkline(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid sensor ID")
		return
	}

	values, err := h.readingSvc.GetSparklineData(id, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get sparkline data")
		return
	}

	data := map[string]interface{}{"Values": values}
	// Render inline SVG sparkline
	w.Header().Set("Content-Type", "text/html")
	if len(values) < 2 {
		fmt.Fprint(w, `<span class="text-xs text-gray-400">No data</span>`)
		return
	}

	// Build SVG sparkline
	minVal, maxVal := values[0], values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	valRange := maxVal - minVal
	if valRange == 0 {
		valRange = 1
	}
	width := 120
	height := 30
	step := float64(width) / float64(len(values)-1)
	var points string
	for i, v := range values {
		x := float64(i) * step
		y := float64(height) - ((v - minVal) / valRange * float64(height-4)) - 2
		if i > 0 {
			points += " "
		}
		points += fmt.Sprintf("%.1f,%.1f", x, y)
	}

	_ = data
	fmt.Fprintf(w, `<svg width="%d" height="%d" class="sparkline"><polyline points="%s" fill="none" stroke="#6366f1" stroke-width="1.5" /></svg>`,
		width, height, points)
}

// APISensorChart returns chart data as JSON (rendered into hidden div, parsed by JS)
func (h *Handler) APISensorChart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid sensor ID")
		return
	}

	rangeStr := r.URL.Query().Get("range")
	var hours int
	switch rangeStr {
	case "1h":
		hours = 1
	case "6h":
		hours = 6
	case "7d":
		hours = 168
	default:
		hours = 24
	}

	sensor, err := h.sensorSvc.GetSensor(id)
	if err != nil || sensor == nil {
		writeError(w, http.StatusNotFound, "Sensor not found")
		return
	}

	start := time.Now().Add(-time.Duration(hours) * time.Hour)
	readings, err := h.readingSvc.GetReadingsInRange(id, start, time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get readings")
		return
	}

	// Downsample if too many points
	maxPoints := 200
	if len(readings) > maxPoints {
		step := len(readings) / maxPoints
		downsampled := make([]models.SensorReading, 0, maxPoints)
		for i := 0; i < len(readings); i += step {
			downsampled = append(downsampled, readings[i])
		}
		readings = downsampled
	}

	chartData := models.ChartData{
		Labels: make([]string, len(readings)),
		Datasets: []models.Dataset{{
			Label:           sensor.Name,
			Data:            make([]float64, len(readings)),
			BorderColor:     "#6366f1",
			BackgroundColor: "rgba(99, 102, 241, 0.1)",
			Fill:            true,
			Tension:         0.3,
		}},
	}

	for i, r := range readings {
		if hours <= 24 {
			chartData.Labels[i] = r.Time.Format("15:04")
		} else {
			chartData.Labels[i] = r.Time.Format("Jan 02 15:04")
		}
		chartData.Datasets[0].Data[i] = r.Value
	}

	// For HTMX: return JSON in text for JS to parse
	if isHTMX(r) {
		w.Header().Set("Content-Type", "text/html")
		jsonData, _ := json.Marshal(chartData)
		w.Write(jsonData)
		return
	}

	writeJSON(w, http.StatusOK, chartData)
}

// APIChartTemperature returns temperature chart data for dashboard
func (h *Handler) APIChartTemperature(w http.ResponseWriter, r *http.Request) {
	sensors, _ := h.sensorSvc.ListSensors()

	colors := []string{"#6366f1", "#ec4899", "#14b8a6", "#f59e0b", "#ef4444"}
	chartData := models.ChartData{
		Labels:   []string{},
		Datasets: []models.Dataset{},
	}

	colorIdx := 0
	for _, sensor := range sensors {
		if sensor.Type != models.SensorTypeTemperature {
			continue
		}

		aggs, err := h.readingSvc.GetHourlyAggregates(sensor.ID, 24)
		if err != nil || len(aggs) == 0 {
			continue
		}

		if len(chartData.Labels) == 0 {
			for _, a := range aggs {
				chartData.Labels = append(chartData.Labels, a.Bucket.Format("15:04"))
			}
		}

		data := make([]float64, len(aggs))
		for i, a := range aggs {
			data[i] = a.AvgValue
		}

		chartData.Datasets = append(chartData.Datasets, models.Dataset{
			Label:           sensor.Name,
			Data:            data,
			BorderColor:     colors[colorIdx%len(colors)],
			BackgroundColor: "transparent",
			Fill:            false,
			Tension:         0.3,
		})
		colorIdx++
	}

	writeJSON(w, http.StatusOK, chartData)
}

// APIChartHumidity returns humidity chart data for dashboard
func (h *Handler) APIChartHumidity(w http.ResponseWriter, r *http.Request) {
	sensors, _ := h.sensorSvc.ListSensors()

	colors := []string{"#3b82f6", "#8b5cf6", "#06b6d4"}
	chartData := models.ChartData{
		Labels:   []string{},
		Datasets: []models.Dataset{},
	}

	colorIdx := 0
	for _, sensor := range sensors {
		if sensor.Type != models.SensorTypeHumidity {
			continue
		}

		aggs, err := h.readingSvc.GetHourlyAggregates(sensor.ID, 24)
		if err != nil || len(aggs) == 0 {
			continue
		}

		if len(chartData.Labels) == 0 {
			for _, a := range aggs {
				chartData.Labels = append(chartData.Labels, a.Bucket.Format("15:04"))
			}
		}

		data := make([]float64, len(aggs))
		for i, a := range aggs {
			data[i] = a.AvgValue
		}

		chartData.Datasets = append(chartData.Datasets, models.Dataset{
			Label:           sensor.Name,
			Data:            data,
			BorderColor:     colors[colorIdx%len(colors)],
			BackgroundColor: "transparent",
			Fill:            false,
			Tension:         0.3,
		})
		colorIdx++
	}

	writeJSON(w, http.StatusOK, chartData)
}

// APIAlertsCount returns the count of active alerts
func (h *Handler) APIAlertsCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.alertSvc.CountActiveAlerts()
	if err != nil {
		w.Write([]byte("0"))
		return
	}
	if count == 0 {
		w.Write([]byte(""))
		return
	}
	fmt.Fprintf(w, "%d", count)
}

// APIAcknowledgeAlert acknowledges a single alert
func (h *Handler) APIAcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid alert ID")
		return
	}

	if err := h.alertSvc.AcknowledgeAlert(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to acknowledge")
		return
	}

	// Return the updated row
	alerts, _ := h.alertSvc.ListAlerts(false, 1)
	for _, a := range alerts {
		if a.ID == id {
			data := map[string]interface{}{"Alert": a}
			_ = data
			fmt.Fprintf(w, `<tr class="hover:bg-gray-50 bg-green-50 fade-in">
				<td class="px-4 py-3"><span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-green-100 text-green-800">%s</span></td>
				<td class="px-4 py-3 text-sm text-gray-900">%s</td>
				<td class="px-4 py-3 text-sm text-gray-700">%s</td>
				<td class="px-4 py-3 text-sm text-gray-500">%.2f</td>
				<td class="px-4 py-3 text-sm text-gray-500">%s</td>
				<td class="px-4 py-3 text-sm text-green-600 text-xs">Acknowledged</td>
			</tr>`, a.Severity, a.SensorName, a.Message, a.Value, a.TriggeredAt.Format("15:04:05"))
			return
		}
	}
	w.Write([]byte(""))
}

// APIAcknowledgeAll acknowledges all alerts
func (h *Handler) APIAcknowledgeAll(w http.ResponseWriter, r *http.Request) {
	if err := h.alertSvc.AcknowledgeAll(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to acknowledge all")
		return
	}
	h.PartialsAlerts(w, r)
}

// APIAlertRules returns alert rules list
func (h *Handler) APIAlertRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.alertSvc.ListAlertRules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list rules")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if len(rules) == 0 {
		fmt.Fprint(w, `<p class="text-sm text-gray-500">No alert rules configured.</p>`)
		return
	}

	fmt.Fprint(w, `<div class="overflow-x-auto"><table class="min-w-full divide-y divide-gray-200">
		<thead class="bg-gray-50">
			<tr>
				<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
				<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
				<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Condition</th>
				<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Value</th>
				<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Severity</th>
				<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Enabled</th>
				<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Action</th>
			</tr>
		</thead><tbody class="divide-y divide-gray-200">`)

	for _, rule := range rules {
		enabledBadge := `<span class="text-green-600 text-xs font-medium">Yes</span>`
		if !rule.Enabled {
			enabledBadge = `<span class="text-gray-400 text-xs font-medium">No</span>`
		}
		fmt.Fprintf(w, `<tr class="hover:bg-gray-50">
			<td class="px-4 py-2 text-sm text-gray-900">%s</td>
			<td class="px-4 py-2 text-sm text-gray-500">%s</td>
			<td class="px-4 py-2 text-sm text-gray-500">%s</td>
			<td class="px-4 py-2 text-sm text-gray-500">%.2f</td>
			<td class="px-4 py-2"><span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-%s-100 text-%s-800">%s</span></td>
			<td class="px-4 py-2">%s</td>
			<td class="px-4 py-2">
				<button hx-delete="/api/alert-rules/%s" hx-target="#alert-rules-list" hx-swap="innerHTML" hx-confirm="Delete this rule?"
						class="text-red-600 hover:text-red-800 text-xs font-medium">Delete</button>
			</td>
		</tr>`, rule.Name, rule.Type, rule.Condition, rule.Value,
			severityColor(string(rule.Severity)), severityColor(string(rule.Severity)), rule.Severity,
			enabledBadge, rule.ID)
	}
	fmt.Fprint(w, `</tbody></table></div>`)
}

// APISensorRules returns alert rules for a specific sensor
func (h *Handler) APISensorRules(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid sensor ID")
		return
	}

	rules, err := h.alertSvc.ListAlertRules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list rules")
		return
	}

	w.Header().Set("Content-Type", "text/html")

	sensorRules := make([]models.AlertRule, 0)
	for _, rule := range rules {
		if rule.SensorID == id {
			sensorRules = append(sensorRules, rule)
		}
	}

	if len(sensorRules) == 0 {
		fmt.Fprint(w, `<p class="text-sm text-gray-500">No alert rules for this sensor.</p>`)
		return
	}

	fmt.Fprint(w, `<div class="space-y-2">`)
	for _, rule := range sensorRules {
		badgeColor := severityColor(string(rule.Severity))
		fmt.Fprintf(w, `<div class="flex items-center justify-between py-2 px-3 rounded-lg bg-gray-50">
			<div class="flex items-center space-x-3">
				<span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-%s-100 text-%s-800">%s</span>
				<span class="text-sm text-gray-900">%s</span>
				<span class="text-xs text-gray-500">%s %s %.2f</span>
			</div>
			<span class="text-xs %s">%s</span>
		</div>`, badgeColor, badgeColor, rule.Severity, rule.Name, rule.Type, rule.Condition, rule.Value,
			map[bool]string{true: "text-green-600", false: "text-gray-400"}[rule.Enabled],
			map[bool]string{true: "Enabled", false: "Disabled"}[rule.Enabled])
	}
	fmt.Fprint(w, `</div>`)
}

// APIDeleteAlertRule deletes an alert rule
func (h *Handler) APIDeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid rule ID")
		return
	}

	if err := h.alertSvc.DeleteAlertRule(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete rule")
		return
	}

	h.APIAlertRules(w, r)
}

// APIGroups returns groups list
func (h *Handler) APIGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.groupSvc.ListGroups()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list groups")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if len(groups) == 0 {
		fmt.Fprint(w, `<p class="text-sm text-gray-500">No sensor groups defined.</p>`)
		return
	}

	fmt.Fprint(w, `<div class="space-y-2">`)
	for _, g := range groups {
		fmt.Fprintf(w, `<div class="flex items-center justify-between py-2 px-3 rounded-lg bg-gray-50">
			<div>
				<span class="text-sm font-medium text-gray-900">%s</span>
				<span class="text-xs text-gray-500 ml-2">%s</span>
			</div>
			<button hx-delete="/api/groups/%s" hx-target="#groups-list" hx-swap="innerHTML" hx-confirm="Delete this group?"
					class="text-red-600 hover:text-red-800 text-xs font-medium">Delete</button>
		</div>`, g.Name, g.Description, g.ID)
	}
	fmt.Fprint(w, `</div>`)
}

// APICreateGroup creates a sensor group
func (h *Handler) APICreateGroup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form")
		return
	}

	group := &models.SensorGroup{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}

	if err := h.groupSvc.CreateGroup(group); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	h.APIGroups(w, r)
}

// APIDeleteGroup deletes a sensor group
func (h *Handler) APIDeleteGroup(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	if err := h.groupSvc.DeleteGroup(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete group")
		return
	}

	h.APIGroups(w, r)
}

// APISSEClients returns the number of SSE clients
func (h *Handler) APISSEClients(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%d", h.broker.ClientCount())
}

// APIReadings returns readings for a sensor (JSON)
func (h *Handler) APIReadings(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid sensor ID"})
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	readings, err := h.readingSvc.GetLatestReadings(id, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get readings"})
		return
	}

	writeJSON(w, http.StatusOK, readings)
}

func severityColor(s string) string {
	switch s {
	case "critical":
		return "red"
	case "warning":
		return "yellow"
	default:
		return "blue"
	}
}
