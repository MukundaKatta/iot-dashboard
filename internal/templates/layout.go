package templates

const baseLayoutTmpl = `<!DOCTYPE html>
<html lang="en" class="h-full bg-gray-50">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>IoT Dashboard{{block "title" .}} {{end}}</title>
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <script src="https://unpkg.com/htmx-ext-sse@2.2.2/sse.js"></script>
    <script defer src="https://unpkg.com/alpinejs@3.14.8/dist/cdn.min.js"></script>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.7/dist/chart.umd.min.js"></script>
    <style>
        [x-cloak] { display: none !important; }
        .sparkline { color: #6366f1; }
        .htmx-indicator { opacity: 0; transition: opacity 200ms ease-in; }
        .htmx-request .htmx-indicator { opacity: 1; }
        .htmx-request.htmx-indicator { opacity: 1; }
        .fade-in { animation: fadeIn 0.3s ease-in; }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }
        .pulse-dot { animation: pulse 2s infinite; }
        @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
    </style>
</head>
<body class="h-full" hx-ext="sse" sse-connect="/api/events">
    <div class="min-h-full">
        {{template "nav" .}}
        <main class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-8">
            {{block "content" .}}{{end}}
        </main>
    </div>

    <!-- Toast notifications from SSE -->
    <div id="toast-container" class="fixed bottom-4 right-4 z-50 space-y-2"
         x-data="{ toasts: [] }"
         @toast.window="toasts.push({id: Date.now(), message: $event.detail.message, type: $event.detail.type}); setTimeout(() => toasts.shift(), 5000)">
        <template x-for="toast in toasts" :key="toast.id">
            <div x-show="true" x-transition
                 class="rounded-lg p-4 shadow-lg text-white text-sm max-w-sm fade-in"
                 :class="toast.type === 'error' ? 'bg-red-600' : toast.type === 'warning' ? 'bg-yellow-500' : 'bg-indigo-600'">
                <span x-text="toast.message"></span>
            </div>
        </template>
    </div>

    <script>
    // SSE event handling for live updates
    document.body.addEventListener('sse:reading', function(e) {
        try {
            const data = JSON.parse(e.detail.data);
            // Update sensor card gauge if visible
            const gauge = document.getElementById('gauge-' + data.sensor_id);
            if (gauge) {
                gauge.textContent = parseFloat(data.value).toFixed(2);
            }
            // Update sparkline data
            const event = new CustomEvent('new-reading', { detail: data });
            document.dispatchEvent(event);
        } catch(err) {
            console.error('SSE reading parse error:', err);
        }
    });

    document.body.addEventListener('sse:alert', function(e) {
        try {
            const data = JSON.parse(e.detail.data);
            window.dispatchEvent(new CustomEvent('toast', {
                detail: { message: data.message, type: data.severity === 'critical' ? 'error' : 'warning' }
            }));
            // Refresh alert badge
            htmx.trigger('#alert-badge', 'refresh');
        } catch(err) {}
    });

    document.body.addEventListener('sse:sensor_status', function(e) {
        try {
            const data = JSON.parse(e.detail.data);
            const statusDot = document.getElementById('status-' + data.id);
            if (statusDot) {
                statusDot.className = data.status === 'online' ? 'h-3 w-3 rounded-full bg-green-400 pulse-dot' : 'h-3 w-3 rounded-full bg-gray-400';
            }
        } catch(err) {}
    });
    </script>
    {{block "scripts" .}}{{end}}
</body>
</html>`
