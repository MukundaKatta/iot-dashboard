package templates

const dashboardPageTmpl = `{{define "title"}} - Overview{{end}}
{{define "content"}}
<div hx-get="/partials/stats" hx-trigger="load, every 30s" hx-swap="innerHTML" id="stats-container">
    {{template "stats_bar" .}}
</div>

<div class="mb-6">
    <h2 class="text-lg font-semibold text-gray-900 mb-4">Live Sensor Readings</h2>
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4" id="sensor-grid"
         hx-get="/partials/sensors" hx-trigger="every 60s" hx-swap="innerHTML">
        {{range .Sensors}}
        {{template "sensor_card" .}}
        {{end}}
    </div>
</div>

<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mt-8">
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <h3 class="text-sm font-semibold text-gray-900 mb-4">Temperature Overview (24h)</h3>
        <canvas id="tempChart" height="200"></canvas>
    </div>
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <h3 class="text-sm font-semibold text-gray-900 mb-4">Humidity Correlation (24h)</h3>
        <canvas id="humidityChart" height="200"></canvas>
    </div>
</div>

<div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 p-5">
    <h3 class="text-sm font-semibold text-gray-900 mb-4">Recent Alerts</h3>
    <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
                <tr>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Severity</th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Sensor</th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Message</th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Value</th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Time</th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Action</th>
                </tr>
            </thead>
            <tbody class="divide-y divide-gray-200" id="recent-alerts"
                   hx-get="/partials/alerts?limit=5" hx-trigger="every 30s" hx-swap="innerHTML">
                {{range .RecentAlerts}}
                {{template "alert_row" .}}
                {{end}}
            </tbody>
        </table>
    </div>
</div>
{{end}}

{{define "scripts"}}
<script>
document.addEventListener('DOMContentLoaded', function() {
    // Load chart data
    fetch('/api/charts/temperature')
        .then(r => r.json())
        .then(data => {
            new Chart(document.getElementById('tempChart'), {
                type: 'line',
                data: {
                    labels: data.labels,
                    datasets: data.datasets.map(ds => ({
                        ...ds,
                        borderWidth: 2,
                        pointRadius: 0,
                        pointHitRadius: 10,
                    }))
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    interaction: { intersect: false, mode: 'index' },
                    scales: {
                        y: { title: { display: true, text: 'Temperature (C)' } },
                        x: { ticks: { maxTicksLimit: 12 } }
                    },
                    plugins: { legend: { position: 'bottom' } }
                }
            });
        });

    fetch('/api/charts/humidity')
        .then(r => r.json())
        .then(data => {
            new Chart(document.getElementById('humidityChart'), {
                type: 'line',
                data: {
                    labels: data.labels,
                    datasets: data.datasets.map(ds => ({
                        ...ds,
                        borderWidth: 2,
                        pointRadius: 0,
                        pointHitRadius: 10,
                    }))
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    interaction: { intersect: false, mode: 'index' },
                    scales: {
                        y: { title: { display: true, text: 'Humidity (%)' } },
                        x: { ticks: { maxTicksLimit: 12 } }
                    },
                    plugins: { legend: { position: 'bottom' } }
                }
            });
        });
});
</script>
{{end}}`

const sensorsPageTmpl = `{{define "title"}} - Sensors{{end}}
{{define "content"}}
<div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold text-gray-900">Sensors</h1>
    <div x-data="{ showModal: false }">
        <button @click="showModal = true"
                class="inline-flex items-center rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700">
            + Add Sensor
        </button>

        <!-- Add Sensor Modal -->
        <div x-show="showModal" x-cloak @keydown.escape.window="showModal = false"
             class="fixed inset-0 z-50 overflow-y-auto" aria-modal="true">
            <div class="flex min-h-screen items-center justify-center p-4">
                <div class="fixed inset-0 bg-gray-500/75 transition-opacity" @click="showModal = false"></div>
                <div class="relative w-full max-w-lg rounded-xl bg-white p-6 shadow-xl" @click.stop>
                    <h2 class="text-lg font-semibold text-gray-900 mb-4">Add New Sensor</h2>
                    <form hx-post="/api/sensors" hx-target="#sensor-list" hx-swap="innerHTML" @htmx:after-request="showModal = false"
                          x-data="{ name: '', type: 'temperature', location: '', unit: 'C', minVal: 0, maxVal: 100 }">
                        <div class="space-y-4">
                            <div>
                                <label class="block text-sm font-medium text-gray-700">Name</label>
                                <input type="text" name="name" x-model="name" required
                                       class="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500">
                            </div>
                            <div>
                                <label class="block text-sm font-medium text-gray-700">Type</label>
                                <select name="type" x-model="type"
                                        class="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm">
                                    <option value="temperature">Temperature</option>
                                    <option value="humidity">Humidity</option>
                                    <option value="pressure">Pressure</option>
                                    <option value="co2">CO2</option>
                                    <option value="light">Light</option>
                                </select>
                            </div>
                            <div>
                                <label class="block text-sm font-medium text-gray-700">Location</label>
                                <input type="text" name="location" x-model="location"
                                       class="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500">
                            </div>
                            <div class="grid grid-cols-3 gap-4">
                                <div>
                                    <label class="block text-sm font-medium text-gray-700">Unit</label>
                                    <input type="text" name="unit" x-model="unit"
                                           class="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm">
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-700">Min</label>
                                    <input type="number" name="min_value" x-model="minVal" step="0.01"
                                           class="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm">
                                </div>
                                <div>
                                    <label class="block text-sm font-medium text-gray-700">Max</label>
                                    <input type="number" name="max_value" x-model="maxVal" step="0.01"
                                           class="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm">
                                </div>
                            </div>
                        </div>
                        <div class="mt-6 flex justify-end space-x-3">
                            <button type="button" @click="showModal = false"
                                    class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">Cancel</button>
                            <button type="submit"
                                    class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700">Create</button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    </div>
</div>

<div class="mb-4" x-data="{ filter: 'all' }">
    <div class="flex space-x-2">
        <button @click="filter = 'all'" :class="filter === 'all' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-500 hover:text-gray-700'"
                class="rounded-lg px-3 py-1.5 text-sm font-medium"
                hx-get="/partials/sensors" hx-target="#sensor-list" hx-swap="innerHTML">All</button>
        <button @click="filter = 'temperature'" :class="filter === 'temperature' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-500'"
                class="rounded-lg px-3 py-1.5 text-sm font-medium"
                hx-get="/partials/sensors?type=temperature" hx-target="#sensor-list" hx-swap="innerHTML">Temperature</button>
        <button @click="filter = 'humidity'" :class="filter === 'humidity' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-500'"
                class="rounded-lg px-3 py-1.5 text-sm font-medium"
                hx-get="/partials/sensors?type=humidity" hx-target="#sensor-list" hx-swap="innerHTML">Humidity</button>
        <button @click="filter = 'co2'" :class="filter === 'co2' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-500'"
                class="rounded-lg px-3 py-1.5 text-sm font-medium"
                hx-get="/partials/sensors?type=co2" hx-target="#sensor-list" hx-swap="innerHTML">CO2</button>
    </div>
</div>

<div id="sensor-list" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
    {{range .Sensors}}
    {{template "sensor_card" .}}
    {{end}}
</div>
{{end}}
{{define "scripts"}}{{end}}`

const sensorDetailPageTmpl = `{{define "title"}} - {{.Sensor.Name}}{{end}}
{{define "content"}}
<div class="mb-6">
    <div class="flex items-center space-x-3">
        <a href="/sensors" hx-get="/sensors" hx-target="main" hx-push-url="true" hx-swap="innerHTML"
           class="text-gray-400 hover:text-gray-600">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
        </a>
        <span class="text-2xl">{{sensorIcon (printf "%s" .Sensor.Type)}}</span>
        <h1 class="text-2xl font-bold text-gray-900">{{.Sensor.Name}}</h1>
        <div id="status-{{.Sensor.ID}}" class="h-3 w-3 rounded-full {{if eq (printf "%s" .Sensor.Status) "online"}}bg-green-400 pulse-dot{{else}}bg-gray-400{{end}}"></div>
    </div>
    <p class="mt-1 text-sm text-gray-500">{{.Sensor.Location}} &middot; {{.Sensor.Type}} &middot; Range: {{formatFloat .Sensor.MinValue}} - {{formatFloat .Sensor.MaxValue}} {{.Sensor.Unit}}</p>
</div>

<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5 text-center">
        <div class="text-sm text-gray-500 mb-1">Current Reading</div>
        <div id="gauge-{{.Sensor.ID}}" class="text-4xl font-bold text-indigo-600">{{formatFloatPtr .Sensor.LastReading}}</div>
        <div class="text-sm text-gray-400">{{.Sensor.Unit}}</div>
    </div>
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5 text-center">
        <div class="text-sm text-gray-500 mb-1">Last Seen</div>
        <div class="text-lg font-semibold text-gray-900">{{timeAgoPtr .Sensor.LastSeenAt}}</div>
    </div>
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5 text-center">
        <div class="text-sm text-gray-500 mb-1">Status</div>
        <div class="text-lg font-semibold {{if eq (printf "%s" .Sensor.Status) "online"}}text-green-600{{else}}text-gray-500{{end}}">{{.Sensor.Status}}</div>
    </div>
</div>

<div x-data="{ timeRange: '24h' }" class="space-y-6">
    <div class="flex space-x-2 mb-2">
        <button @click="timeRange = '1h'" :class="timeRange === '1h' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-500'"
                class="rounded-lg px-3 py-1.5 text-sm font-medium"
                hx-get="/api/sensors/{{.Sensor.ID}}/chart?range=1h" hx-target="#chart-data" hx-swap="innerHTML">1h</button>
        <button @click="timeRange = '6h'" :class="timeRange === '6h' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-500'"
                class="rounded-lg px-3 py-1.5 text-sm font-medium"
                hx-get="/api/sensors/{{.Sensor.ID}}/chart?range=6h" hx-target="#chart-data" hx-swap="innerHTML">6h</button>
        <button @click="timeRange = '24h'" :class="timeRange === '24h' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-500'"
                class="rounded-lg px-3 py-1.5 text-sm font-medium"
                hx-get="/api/sensors/{{.Sensor.ID}}/chart?range=24h" hx-target="#chart-data" hx-swap="innerHTML">24h</button>
        <button @click="timeRange = '7d'" :class="timeRange === '7d' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-500'"
                class="rounded-lg px-3 py-1.5 text-sm font-medium"
                hx-get="/api/sensors/{{.Sensor.ID}}/chart?range=7d" hx-target="#chart-data" hx-swap="innerHTML">7d</button>
    </div>

    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <h3 class="text-sm font-semibold text-gray-900 mb-4">{{.Sensor.Name}} Over Time</h3>
        <div style="height: 300px;">
            <canvas id="sensorChart"></canvas>
        </div>
        <div id="chart-data" class="hidden" hx-get="/api/sensors/{{.Sensor.ID}}/chart?range=24h" hx-trigger="load" hx-swap="innerHTML"></div>
    </div>
</div>

<div class="mt-8 bg-white rounded-xl shadow-sm border border-gray-200 p-5">
    <h3 class="text-sm font-semibold text-gray-900 mb-4">Alert Rules</h3>
    <div hx-get="/api/sensors/{{.Sensor.ID}}/rules" hx-trigger="load" hx-swap="innerHTML" id="sensor-rules"></div>
</div>

<div class="mt-4 flex justify-end">
    <button hx-delete="/api/sensors/{{.Sensor.ID}}" hx-confirm="Are you sure you want to delete this sensor?"
            hx-target="main" hx-swap="innerHTML" hx-push-url="/sensors"
            class="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700">Delete Sensor</button>
</div>
{{end}}

{{define "scripts"}}
<script>
let sensorChart = null;

function updateSensorChart(chartData) {
    const ctx = document.getElementById('sensorChart');
    if (!ctx) return;

    if (sensorChart) {
        sensorChart.destroy();
    }

    sensorChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: chartData.labels,
            datasets: chartData.datasets.map(ds => ({
                ...ds,
                borderWidth: 2,
                pointRadius: 0,
                pointHitRadius: 10,
            }))
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: { intersect: false, mode: 'index' },
            scales: {
                y: { title: { display: true, text: '{{.Sensor.Unit}}' } },
                x: { ticks: { maxTicksLimit: 12 } }
            },
            plugins: { legend: { position: 'bottom' } }
        }
    });
}

// Listen for chart data loaded via HTMX
document.body.addEventListener('htmx:afterSwap', function(e) {
    if (e.detail.target.id === 'chart-data') {
        try {
            const data = JSON.parse(e.detail.target.textContent);
            updateSensorChart(data);
        } catch(err) {
            console.error('Chart data parse error:', err);
        }
    }
});

// Update gauge from SSE
document.addEventListener('new-reading', function(e) {
    if (e.detail.sensor_id === '{{.Sensor.ID}}') {
        const gauge = document.getElementById('gauge-{{.Sensor.ID}}');
        if (gauge) gauge.textContent = parseFloat(e.detail.value).toFixed(2);
    }
});
</script>
{{end}}`

const alertsPageTmpl = `{{define "title"}} - Alerts{{end}}
{{define "content"}}
<div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold text-gray-900">Alerts</h1>
    <div class="flex space-x-3">
        <button hx-post="/api/alerts/acknowledge-all" hx-target="#alerts-table-body" hx-swap="innerHTML"
                class="inline-flex items-center rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50">
            Acknowledge All
        </button>
    </div>
</div>

<div x-data="{ filter: 'active' }" class="mb-4">
    <div class="flex space-x-2">
        <button @click="filter = 'active'" :class="filter === 'active' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-500'"
                class="rounded-lg px-3 py-1.5 text-sm font-medium"
                hx-get="/partials/alerts?active=true" hx-target="#alerts-table-body" hx-swap="innerHTML">Active</button>
        <button @click="filter = 'all'" :class="filter === 'all' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-500'"
                class="rounded-lg px-3 py-1.5 text-sm font-medium"
                hx-get="/partials/alerts?active=false" hx-target="#alerts-table-body" hx-swap="innerHTML">All</button>
    </div>
</div>

<div class="bg-white rounded-xl shadow-sm border border-gray-200">
    <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
                <tr>
                    <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Severity</th>
                    <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Sensor</th>
                    <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Message</th>
                    <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Value</th>
                    <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Time</th>
                    <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Action</th>
                </tr>
            </thead>
            <tbody id="alerts-table-body" class="divide-y divide-gray-200"
                   hx-get="/partials/alerts?active=true" hx-trigger="load" hx-swap="innerHTML">
            </tbody>
        </table>
    </div>
</div>
{{end}}
{{define "scripts"}}{{end}}`

const settingsPageTmpl = `{{define "title"}} - Settings{{end}}
{{define "content"}}
<h1 class="text-2xl font-bold text-gray-900 mb-6">Settings</h1>

<div class="space-y-6">
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Alert Rules</h2>
        <div id="alert-rules-list" hx-get="/api/alert-rules" hx-trigger="load" hx-swap="innerHTML">
            <div class="text-sm text-gray-500">Loading alert rules...</div>
        </div>
    </div>

    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Sensor Groups</h2>
        <div x-data="{ showForm: false }" class="space-y-4">
            <div id="groups-list" hx-get="/api/groups" hx-trigger="load" hx-swap="innerHTML">
                <div class="text-sm text-gray-500">Loading groups...</div>
            </div>
            <button @click="showForm = !showForm"
                    class="text-sm text-indigo-600 hover:text-indigo-800 font-medium">+ Add Group</button>
            <form x-show="showForm" x-cloak hx-post="/api/groups" hx-target="#groups-list" hx-swap="innerHTML"
                  @htmx:after-request="showForm = false" class="flex space-x-3 items-end">
                <div class="flex-1">
                    <label class="block text-sm font-medium text-gray-700">Name</label>
                    <input type="text" name="name" required class="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm">
                </div>
                <div class="flex-1">
                    <label class="block text-sm font-medium text-gray-700">Description</label>
                    <input type="text" name="description" class="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm">
                </div>
                <button type="submit" class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700">Create</button>
            </form>
        </div>
    </div>

    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">System Info</h2>
        <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
                <span class="text-gray-500">SSE Connections:</span>
                <span class="font-medium" hx-get="/api/system/sse-clients" hx-trigger="load, every 10s" hx-swap="innerHTML">0</span>
            </div>
            <div>
                <span class="text-gray-500">Simulator:</span>
                <span class="font-medium text-green-600">Running</span>
            </div>
        </div>
    </div>
</div>
{{end}}
{{define "scripts"}}{{end}}`
