package templates

import (
	"fmt"
	"html/template"
	"io"
	"math"
	"strings"
	"time"
)

// Engine manages HTML templates
type Engine struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
}

// NewEngine creates a new template engine
func NewEngine() *Engine {
	e := &Engine{
		templates: make(map[string]*template.Template),
	}

	e.funcMap = template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"formatTimeShort": func(t time.Time) string {
			return t.Format("15:04")
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"timeAgo": func(t time.Time) string {
			d := time.Since(t)
			switch {
			case d < time.Minute:
				return "just now"
			case d < time.Hour:
				return fmt.Sprintf("%dm ago", int(d.Minutes()))
			case d < 24*time.Hour:
				return fmt.Sprintf("%dh ago", int(d.Hours()))
			default:
				return fmt.Sprintf("%dd ago", int(d.Hours()/24))
			}
		},
		"timeAgoPtr": func(t *time.Time) string {
			if t == nil {
				return "never"
			}
			d := time.Since(*t)
			switch {
			case d < time.Minute:
				return "just now"
			case d < time.Hour:
				return fmt.Sprintf("%dm ago", int(d.Minutes()))
			case d < 24*time.Hour:
				return fmt.Sprintf("%dh ago", int(d.Hours()))
			default:
				return fmt.Sprintf("%dd ago", int(d.Hours()/24))
			}
		},
		"formatFloat": func(f float64) string {
			return fmt.Sprintf("%.2f", f)
		},
		"formatFloatPtr": func(f *float64) string {
			if f == nil {
				return "N/A"
			}
			return fmt.Sprintf("%.2f", *f)
		},
		"round": func(f float64) float64 {
			return math.Round(f*100) / 100
		},
		"pct": func(value, min, max float64) float64 {
			if max == min {
				return 50
			}
			return ((value - min) / (max - min)) * 100
		},
		"severityColor": func(s string) string {
			switch s {
			case "critical":
				return "red"
			case "warning":
				return "yellow"
			default:
				return "blue"
			}
		},
		"statusColor": func(s string) string {
			switch s {
			case "online":
				return "green"
			case "offline":
				return "gray"
			case "error":
				return "red"
			default:
				return "gray"
			}
		},
		"sensorIcon": func(s string) string {
			switch s {
			case "temperature":
				return "🌡️"
			case "humidity":
				return "💧"
			case "pressure":
				return "🌀"
			case "co2":
				return "💨"
			case "light":
				return "☀️"
			default:
				return "📡"
			}
		},
		"sparklineSVG": func(values []float64, width, height int) template.HTML {
			if len(values) == 0 {
				return template.HTML("")
			}
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

			var points []string
			step := float64(width) / float64(len(values)-1)
			for i, v := range values {
				x := float64(i) * step
				y := float64(height) - ((v - minVal) / valRange * float64(height-4)) - 2
				points = append(points, fmt.Sprintf("%.1f,%.1f", x, y))
			}
			path := strings.Join(points, " ")
			svg := fmt.Sprintf(`<svg width="%d" height="%d" class="sparkline"><polyline points="%s" fill="none" stroke="currentColor" stroke-width="1.5" /></svg>`,
				width, height, path)
			return template.HTML(svg)
		},
		"jsonArray": func(values []float64) string {
			if len(values) == 0 {
				return "[]"
			}
			parts := make([]string, len(values))
			for i, v := range values {
				parts[i] = fmt.Sprintf("%.2f", v)
			}
			return "[" + strings.Join(parts, ",") + "]"
		},
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}

	return e
}

// LoadTemplates parses all templates
func (e *Engine) LoadTemplates() error {
	layouts := map[string]string{
		"base": baseLayoutTmpl,
	}

	partials := map[string]string{
		"nav":         navPartialTmpl,
		"sensor_card": sensorCardPartialTmpl,
		"alert_row":   alertRowPartialTmpl,
		"stats_bar":   statsBarPartialTmpl,
	}

	pages := map[string]string{
		"dashboard":     dashboardPageTmpl,
		"sensors":       sensorsPageTmpl,
		"sensor_detail": sensorDetailPageTmpl,
		"alerts":        alertsPageTmpl,
		"settings":      settingsPageTmpl,
	}

	// Register partial-only templates for HTMX responses.
	// These need all sub-partial {{define}} blocks available.

	// For stats_bar_partial, the template defines {{define "stats_bar"}}...{{end}}.
	// We register it and will look up the named block when rendering.
	// For sensor_list_partial and alert_list_partial, they reference
	// {{template "sensor_card" .}} and {{template "alert_row" .}}.

	partialOnlyTemplates := map[string]struct {
		content    string
		renderName string // the {{define}} block to execute, or "" for top-level
	}{
		"stats_bar_partial":   {statsBarPartialTmpl, "stats_bar"},
		"sensor_card_partial": {sensorCardPartialTmpl, "sensor_card"},
		"alert_row_partial":   {alertRowPartialTmpl, "alert_row"},
		"sensor_list_partial": {sensorListPartialTmpl, ""},
		"alert_list_partial":  {alertListPartialTmpl, ""},
	}

	var err error
	for name, info := range partialOnlyTemplates {
		t := template.New(name).Funcs(e.funcMap)

		// Include all sub-partial definitions
		for _, partialContent := range partials {
			t, err = t.Parse(partialContent)
			if err != nil {
				return fmt.Errorf("parse sub-partial for %s: %w", name, err)
			}
		}

		// Parse the main content
		t, err = t.Parse(info.content)
		if err != nil {
			return fmt.Errorf("parse partial-only %s: %w", name, err)
		}

		// If there's a specific {{define}} block to render, look it up
		if info.renderName != "" {
			named := t.Lookup(info.renderName)
			if named == nil {
				return fmt.Errorf("partial %s: defined block %q not found", name, info.renderName)
			}
			e.templates[name] = named
		} else {
			e.templates[name] = t
		}
	}

	// Build full page templates: base + partials + page
	for pageName, pageContent := range pages {
		t := template.New("base").Funcs(e.funcMap)

		// Parse base layout
		t, err := t.Parse(layouts["base"])
		if err != nil {
			return fmt.Errorf("parse base for %s: %w", pageName, err)
		}

		// Parse all partials
		for _, partialContent := range partials {
			t, err = t.Parse(partialContent)
			if err != nil {
				return fmt.Errorf("parse partial for %s: %w", pageName, err)
			}
		}

		// Parse page content
		t, err = t.Parse(pageContent)
		if err != nil {
			return fmt.Errorf("parse page %s: %w", pageName, err)
		}

		e.templates[pageName] = t
	}

	// Build content-only templates for HTMX navigation (no base layout).
	// These wrap the page content block so we can render just the inner HTML.
	contentWrapper := `{{block "content" .}}{{end}}{{block "scripts" .}}{{end}}`
	for pageName, pageContent := range pages {
		contentName := pageName + "_content"
		t := template.New(contentName).Funcs(e.funcMap)

		// Parse all partials (needed for templates that reference them)
		for _, partialContent := range partials {
			t, err = t.Parse(partialContent)
			if err != nil {
				return fmt.Errorf("parse partial for content %s: %w", pageName, err)
			}
		}

		// Parse the page (which defines "content" and "scripts" blocks)
		t, err = t.Parse(pageContent)
		if err != nil {
			return fmt.Errorf("parse page content %s: %w", pageName, err)
		}

		// Parse the wrapper that invokes the content block
		t, err = t.Parse(contentWrapper)
		if err != nil {
			return fmt.Errorf("parse content wrapper %s: %w", pageName, err)
		}

		e.templates[contentName] = t
	}

	return nil
}

// Render renders a full page template
func (e *Engine) Render(w io.Writer, name string, data interface{}) error {
	t, ok := e.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return t.Execute(w, data)
}

// RenderContent renders only the content block of a page (for HTMX navigation)
func (e *Engine) RenderContent(w io.Writer, name string, data interface{}) error {
	contentName := name + "_content"
	t, ok := e.templates[contentName]
	if !ok {
		return fmt.Errorf("content template %s not found", contentName)
	}
	return t.Execute(w, data)
}

// RenderPartial renders a partial template (for HTMX responses)
func (e *Engine) RenderPartial(w io.Writer, name string, data interface{}) error {
	t, ok := e.templates[name]
	if !ok {
		return fmt.Errorf("partial template %s not found", name)
	}
	return t.Execute(w, data)
}
