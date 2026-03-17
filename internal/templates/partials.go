package templates

const navPartialTmpl = `{{define "nav"}}
<nav class="bg-white shadow-sm border-b border-gray-200" x-data="{ mobileOpen: false }">
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="flex h-16 justify-between">
            <div class="flex">
                <div class="flex flex-shrink-0 items-center">
                    <span class="text-xl font-bold text-indigo-600">IoT Dashboard</span>
                </div>
                <div class="hidden sm:ml-8 sm:flex sm:space-x-4">
                    <a href="/" class="inline-flex items-center border-b-2 px-3 pt-1 text-sm font-medium {{if eq .ActivePage "dashboard"}}border-indigo-500 text-gray-900{{else}}border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700{{end}}"
                       hx-get="/" hx-target="main" hx-push-url="true" hx-swap="innerHTML">
                        Dashboard
                    </a>
                    <a href="/sensors" class="inline-flex items-center border-b-2 px-3 pt-1 text-sm font-medium {{if eq .ActivePage "sensors"}}border-indigo-500 text-gray-900{{else}}border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700{{end}}"
                       hx-get="/sensors" hx-target="main" hx-push-url="true" hx-swap="innerHTML">
                        Sensors
                    </a>
                    <a href="/alerts" class="relative inline-flex items-center border-b-2 px-3 pt-1 text-sm font-medium {{if eq .ActivePage "alerts"}}border-indigo-500 text-gray-900{{else}}border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700{{end}}"
                       hx-get="/alerts" hx-target="main" hx-push-url="true" hx-swap="innerHTML">
                        Alerts
                        <span id="alert-badge" hx-get="/api/alerts/count" hx-trigger="load, refresh, every 30s" hx-swap="innerHTML"
                              class="ml-1 inline-flex items-center rounded-full bg-red-100 px-2 py-0.5 text-xs font-medium text-red-700"></span>
                    </a>
                    <a href="/settings" class="inline-flex items-center border-b-2 px-3 pt-1 text-sm font-medium {{if eq .ActivePage "settings"}}border-indigo-500 text-gray-900{{else}}border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700{{end}}"
                       hx-get="/settings" hx-target="main" hx-push-url="true" hx-swap="innerHTML">
                        Settings
                    </a>
                </div>
            </div>
            <div class="flex items-center space-x-3">
                <span class="text-xs text-gray-400" id="sse-status">
                    <span class="inline-block h-2 w-2 rounded-full bg-green-400 pulse-dot"></span> Live
                </span>
            </div>
            <div class="-mr-2 flex items-center sm:hidden">
                <button @click="mobileOpen = !mobileOpen" class="inline-flex items-center justify-center rounded-md p-2 text-gray-400 hover:bg-gray-100">
                    <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path x-show="!mobileOpen" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
                        <path x-show="mobileOpen" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                    </svg>
                </button>
            </div>
        </div>
    </div>
    <div x-show="mobileOpen" x-cloak class="sm:hidden">
        <div class="space-y-1 pb-3 pt-2">
            <a href="/" class="block border-l-4 py-2 pl-3 pr-4 text-base font-medium {{if eq .ActivePage "dashboard"}}border-indigo-500 bg-indigo-50 text-indigo-700{{else}}border-transparent text-gray-500{{end}}">Dashboard</a>
            <a href="/sensors" class="block border-l-4 py-2 pl-3 pr-4 text-base font-medium {{if eq .ActivePage "sensors"}}border-indigo-500 bg-indigo-50 text-indigo-700{{else}}border-transparent text-gray-500{{end}}">Sensors</a>
            <a href="/alerts" class="block border-l-4 py-2 pl-3 pr-4 text-base font-medium {{if eq .ActivePage "alerts"}}border-indigo-500 bg-indigo-50 text-indigo-700{{else}}border-transparent text-gray-500{{end}}">Alerts</a>
            <a href="/settings" class="block border-l-4 py-2 pl-3 pr-4 text-base font-medium {{if eq .ActivePage "settings"}}border-indigo-500 bg-indigo-50 text-indigo-700{{else}}border-transparent text-gray-500{{end}}">Settings</a>
        </div>
    </div>
</nav>
{{end}}`

const sensorCardPartialTmpl = `{{define "sensor_card"}}
<div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5 hover:shadow-md transition-shadow cursor-pointer fade-in"
     hx-get="/sensors/{{.ID}}" hx-target="main" hx-push-url="true" hx-swap="innerHTML">
    <div class="flex items-center justify-between mb-3">
        <div class="flex items-center space-x-2">
            <span class="text-xl">{{sensorIcon (printf "%s" .Type)}}</span>
            <h3 class="text-sm font-semibold text-gray-900 truncate">{{.Name}}</h3>
        </div>
        <div id="status-{{.ID}}" class="h-3 w-3 rounded-full {{if eq (printf "%s" .Status) "online"}}bg-green-400 pulse-dot{{else if eq (printf "%s" .Status) "error"}}bg-red-400{{else}}bg-gray-400{{end}}"></div>
    </div>
    <div class="flex items-end justify-between">
        <div>
            <span id="gauge-{{.ID}}" class="text-3xl font-bold text-gray-900">{{formatFloatPtr .LastReading}}</span>
            <span class="text-sm text-gray-500 ml-1">{{.Unit}}</span>
        </div>
        <div class="text-right">
            <div class="text-xs text-gray-400">{{timeAgoPtr .LastSeenAt}}</div>
        </div>
    </div>
    <div class="mt-3 text-xs text-gray-500">{{.Location}}</div>
    <div class="mt-2" id="sparkline-{{.ID}}" hx-get="/api/sensors/{{.ID}}/sparkline" hx-trigger="load" hx-swap="innerHTML"></div>
</div>
{{end}}`

const alertRowPartialTmpl = `{{define "alert_row"}}
<tr class="hover:bg-gray-50 fade-in {{if not .Acknowledged}}bg-{{severityColor (printf "%s" .Severity)}}-50{{end}}">
    <td class="whitespace-nowrap px-4 py-3">
        <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium
            {{if eq (printf "%s" .Severity) "critical"}}bg-red-100 text-red-800
            {{else if eq (printf "%s" .Severity) "warning"}}bg-yellow-100 text-yellow-800
            {{else}}bg-blue-100 text-blue-800{{end}}">
            {{.Severity}}
        </span>
    </td>
    <td class="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{{.SensorName}}</td>
    <td class="px-4 py-3 text-sm text-gray-700">{{.Message}}</td>
    <td class="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{{formatFloat .Value}}</td>
    <td class="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{{timeAgo .TriggeredAt}}</td>
    <td class="whitespace-nowrap px-4 py-3 text-sm">
        {{if not .Acknowledged}}
        <button hx-post="/api/alerts/{{.ID}}/acknowledge" hx-target="closest tr" hx-swap="outerHTML"
                class="text-indigo-600 hover:text-indigo-900 text-xs font-medium">Acknowledge</button>
        {{else}}
        <span class="text-green-600 text-xs">Acknowledged</span>
        {{end}}
    </td>
</tr>
{{end}}`

const statsBarPartialTmpl = `{{define "stats_bar"}}
<div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="text-sm font-medium text-gray-500">Total Sensors</div>
        <div class="mt-1 text-3xl font-bold text-gray-900">{{.Stats.TotalSensors}}</div>
    </div>
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="text-sm font-medium text-gray-500">Online</div>
        <div class="mt-1 text-3xl font-bold text-green-600">{{.Stats.OnlineSensors}}</div>
    </div>
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="text-sm font-medium text-gray-500">Active Alerts</div>
        <div class="mt-1 text-3xl font-bold {{if gt .Stats.ActiveAlerts 0}}text-red-600{{else}}text-gray-900{{end}}">{{.Stats.ActiveAlerts}}</div>
    </div>
    <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="text-sm font-medium text-gray-500">Readings (24h)</div>
        <div class="mt-1 text-3xl font-bold text-gray-900">{{.Stats.ReadingsToday}}</div>
    </div>
</div>
{{end}}`

const sensorListPartialTmpl = `{{range .Sensors}}{{template "sensor_card" .}}{{end}}{{if not .Sensors}}<div class="col-span-full text-center py-8 text-gray-500">No sensors found.</div>{{end}}`

const alertListPartialTmpl = `{{range .Alerts}}{{template "alert_row" .}}{{end}}{{if not .Alerts}}<tr><td colspan="6" class="px-4 py-8 text-center text-gray-500">No alerts found.</td></tr>{{end}}`
