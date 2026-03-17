package templates

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine()
	assert.NotNil(t, e)
	assert.NotNil(t, e.funcMap)
}

func TestLoadTemplates(t *testing.T) {
	e := NewEngine()
	err := e.LoadTemplates()
	require.NoError(t, err)

	// Check that all page templates are loaded
	expectedPages := []string{"dashboard", "sensors", "sensor_detail", "alerts", "settings"}
	for _, page := range expectedPages {
		_, ok := e.templates[page]
		assert.True(t, ok, "Template %s should be loaded", page)
	}

	// Check content-only templates for HTMX navigation
	for _, page := range expectedPages {
		contentName := page + "_content"
		_, ok := e.templates[contentName]
		assert.True(t, ok, "Content template %s should be loaded", contentName)
	}

	// Check partial templates
	expectedPartials := []string{"sensor_card_partial", "alert_row_partial", "stats_bar_partial", "sensor_list_partial", "alert_list_partial"}
	for _, partial := range expectedPartials {
		_, ok := e.templates[partial]
		assert.True(t, ok, "Partial template %s should be loaded", partial)
	}
}

func TestRenderMissingTemplate(t *testing.T) {
	e := NewEngine()
	err := e.LoadTemplates()
	require.NoError(t, err)

	var buf bytes.Buffer
	err = e.Render(&buf, "nonexistent", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRenderPartialMissing(t *testing.T) {
	e := NewEngine()
	err := e.LoadTemplates()
	require.NoError(t, err)

	var buf bytes.Buffer
	err = e.RenderPartial(&buf, "nonexistent_partial", nil)
	assert.Error(t, err)
}

func TestRenderContentMissing(t *testing.T) {
	e := NewEngine()
	err := e.LoadTemplates()
	require.NoError(t, err)

	var buf bytes.Buffer
	err = e.RenderContent(&buf, "nonexistent", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestTemplateFuncMap(t *testing.T) {
	e := NewEngine()

	// Test formatFloat
	fn, ok := e.funcMap["formatFloat"]
	assert.True(t, ok)
	result := fn.(func(float64) string)(22.456)
	assert.Equal(t, "22.46", result)

	// Test severityColor
	fn, ok = e.funcMap["severityColor"]
	assert.True(t, ok)
	assert.Equal(t, "red", fn.(func(string) string)("critical"))
	assert.Equal(t, "yellow", fn.(func(string) string)("warning"))
	assert.Equal(t, "blue", fn.(func(string) string)("info"))

	// Test statusColor
	fn, ok = e.funcMap["statusColor"]
	assert.True(t, ok)
	assert.Equal(t, "green", fn.(func(string) string)("online"))
	assert.Equal(t, "gray", fn.(func(string) string)("offline"))
	assert.Equal(t, "red", fn.(func(string) string)("error"))

	// Test sensorIcon
	fn, ok = e.funcMap["sensorIcon"]
	assert.True(t, ok)
	assert.NotEmpty(t, fn.(func(string) string)("temperature"))

	// Test add/sub
	addFn := e.funcMap["add"].(func(int, int) int)
	assert.Equal(t, 5, addFn(2, 3))
	subFn := e.funcMap["sub"].(func(int, int) int)
	assert.Equal(t, 1, subFn(3, 2))

	// Test pct
	pctFn := e.funcMap["pct"].(func(float64, float64, float64) float64)
	assert.InDelta(t, 50.0, pctFn(50, 0, 100), 0.01)
	assert.InDelta(t, 0.0, pctFn(0, 0, 100), 0.01)
	assert.InDelta(t, 100.0, pctFn(100, 0, 100), 0.01)
	assert.InDelta(t, 50.0, pctFn(5, 5, 5), 0.01) // edge case: min==max

	// Test jsonArray
	jsonFn := e.funcMap["jsonArray"].(func([]float64) string)
	assert.Equal(t, "[]", jsonFn([]float64{}))
	assert.Equal(t, "[1.00,2.50,3.00]", jsonFn([]float64{1, 2.5, 3}))

	// Test formatFloatPtr
	fmtPtrFn := e.funcMap["formatFloatPtr"].(func(*float64) string)
	assert.Equal(t, "N/A", fmtPtrFn(nil))
	v := 22.5
	assert.Equal(t, "22.50", fmtPtrFn(&v))
}

func TestSeq(t *testing.T) {
	e := NewEngine()
	seqFn := e.funcMap["seq"].(func(int) []int)

	result := seqFn(5)
	assert.Equal(t, []int{0, 1, 2, 3, 4}, result)

	result = seqFn(0)
	assert.Empty(t, result)
}
