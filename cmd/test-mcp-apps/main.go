// test-mcp-apps serves the MCP App UIs in a test harness for local development.
// It simulates the MCP Apps protocol, letting you preview and tune the UIs
// without connecting to a real MCP client or backend.
//
// Usage:
//
//	go run ./cmd/test-mcp-apps
//
// Then open http://localhost:9199 in your browser.
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, chartHarness)
	})
	http.HandleFunc("/chart", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, buildChartHTML())
	})
	http.HandleFunc("/trace", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, traceHarness)
	})
	http.HandleFunc("/trace-app", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, buildTraceHTML())
	})

	addr := "127.0.0.1:9199"
	fmt.Fprintf(os.Stderr, "MCP Apps test harness: http://%s\n", addr)
	fmt.Fprintf(os.Stderr, "  Chart:  http://%s/\n", addr)
	fmt.Fprintf(os.Stderr, "  Trace:  http://%s/trace\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func buildChartHTML() string {
	tmpl, err := os.ReadFile("pkg/mcp/ui/chart.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Run this command from the repo root:\n  go run ./cmd/test-mcp-apps\n\nerror: %v\n", err)
		os.Exit(1)
	}
	styles, _ := os.ReadFile("pkg/mcp/ui/styles.css")
	chartLib, _ := os.ReadFile("pkg/mcp/ui/chart.min.js")
	dateAdapter, _ := os.ReadFile("pkg/mcp/ui/date-adapter.js")
	app, _ := os.ReadFile("pkg/mcp/ui/app.js")

	r := strings.NewReplacer(
		"{{STYLES}}", string(styles),
		"{{CHART_LIB}}", string(chartLib),
		"{{DATE_ADAPTER}}", string(dateAdapter),
		"{{APP}}", string(app),
	)
	return r.Replace(string(tmpl))
}

func buildTraceHTML() string {
	tmpl, err := os.ReadFile("pkg/mcp/ui/trace.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Run this command from the repo root:\n  go run ./cmd/test-mcp-apps\n\nerror: %v\n", err)
		os.Exit(1)
	}
	styles, _ := os.ReadFile("pkg/mcp/ui/trace.css")
	app, _ := os.ReadFile("pkg/mcp/ui/trace.js")

	r := strings.NewReplacer(
		"{{TRACE_STYLES}}", string(styles),
		"{{TRACE_APP}}", string(app),
	)
	return r.Replace(string(tmpl))
}

const chartHarness = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Chart Test Harness</title>
<style>
  :root {
    --bg: #f3f4f6; --text: #111827; --surface: #ffffff;
    --border: #d1d5db; --subtle: #6b7280;
  }
  [data-theme="dark"] {
    --bg: #111827; --text: #f3f4f6; --surface: #1f2937;
    --border: #374151; --subtle: #9ca3af;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: system-ui, -apple-system, sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    transition: background 0.2s, color 0.2s;
  }
  .toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 20px;
    border-bottom: 1px solid var(--border);
    background: var(--surface);
    flex-wrap: wrap;
  }
  .toolbar h1 {
    font-size: 15px;
    font-weight: 600;
    margin-right: auto;
  }
  .nav-link {
    font-size: 13px;
    color: var(--subtle);
    text-decoration: none;
  }
  .nav-link:hover { color: var(--text); }
  .btn-group {
    display: flex;
    border: 1px solid var(--border);
    border-radius: 8px;
    overflow: hidden;
  }
  .btn-group button {
    padding: 6px 14px;
    font-size: 13px;
    font-family: inherit;
    background: var(--surface);
    color: var(--text);
    border: none;
    border-right: 1px solid var(--border);
    cursor: pointer;
    transition: background 0.15s;
  }
  .btn-group button:last-child { border-right: none; }
  .btn-group button:hover { background: var(--bg); }
  .btn-group button.active {
    background: #2563eb; color: #fff;
  }
  button.action {
    padding: 6px 16px;
    font-size: 13px;
    font-family: inherit;
    background: #2563eb;
    color: #fff;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: background 0.15s;
  }
  button.action:hover { background: #1d4ed8; }
  select, input[type="text"] {
    padding: 6px 10px;
    font-size: 13px;
    font-family: inherit;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--surface);
    color: var(--text);
  }
  select { cursor: pointer; }
  input[type="text"] { width: 220px; }
  label {
    font-size: 13px;
    color: var(--subtle);
  }
  .frame-area {
    padding: 20px;
  }
  iframe {
    width: 100%;
    height: 0;
    border: none;
    display: block;
    border-radius: 4px;
    transition: height 0.2s;
  }
</style>
</head>
<body>

<div class="toolbar">
  <h1>Chart Test Harness</h1>
  <a class="nav-link" href="/trace">Trace Waterfall &rarr;</a>

  <label>Theme</label>
  <div class="btn-group" id="theme-btns">
    <button class="active" data-theme="light">Light</button>
    <button data-theme="dark">Dark</button>
  </div>

  <label>Series</label>
  <select id="series-count">
    <option value="1">1</option>
    <option value="3">3</option>
    <option value="5" selected>5</option>
    <option value="10">10</option>
  </select>

  <label>Range</label>
  <select id="time-range">
    <option value="1800">30 min</option>
    <option value="7200" selected>2 hours</option>
    <option value="43200">12 hours</option>
    <option value="86400">24 hours</option>
    <option value="259200">3 days</option>
  </select>

  <label>Title</label>
  <input type="text" id="title-input" value="CPU Usage by Pod (Last 2 Hours)" placeholder="Chart title (optional)">

  <label>Description</label>
  <input type="text" id="description-input" value="Shows CPU usage rate per pod in the openshift-monitoring namespace" placeholder="Chart description (optional)">

  <button class="action" onclick="sendData()">Send Data</button>
  <button class="action" onclick="clearData()" style="background:#dc2626">Clear</button>
</div>

<div class="frame-area">
  <iframe id="f" src="/chart"></iframe>
</div>

<script>
var dark = false;
var f = document.getElementById("f");

// Restore selections from URL params
(function() {
  var p = new URLSearchParams(window.location.search);
  if (p.has("theme")) {
    var t = p.get("theme");
    if (t === "dark" || t === "light") {
      dark = t === "dark";
      document.documentElement.setAttribute("data-theme", t);
      document.querySelectorAll("#theme-btns button").forEach(function(b) {
        b.classList.toggle("active", b.dataset.theme === t);
      });
    }
  }
  if (p.has("series")) {
    var s = document.getElementById("series-count");
    if (s.querySelector('option[value="' + p.get("series") + '"]')) s.value = p.get("series");
  }
  if (p.has("range")) {
    var r = document.getElementById("time-range");
    if (r.querySelector('option[value="' + p.get("range") + '"]')) r.value = p.get("range");
  }
  if (p.has("title")) {
    document.getElementById("title-input").value = p.get("title");
  }
  if (p.has("description")) {
    document.getElementById("description-input").value = p.get("description");
  }
})();

function updateURL() {
  var p = new URLSearchParams();
  p.set("theme", dark ? "dark" : "light");
  p.set("series", document.getElementById("series-count").value);
  p.set("range", document.getElementById("time-range").value);
  var title = document.getElementById("title-input").value.trim();
  if (title) p.set("title", title);
  var description = document.getElementById("description-input").value.trim();
  if (description) p.set("description", description);
  history.replaceState(null, "", "?" + p.toString());
}

// Theme switching
document.getElementById("theme-btns").addEventListener("click", function(e) {
  var btn = e.target.closest("button");
  if (!btn) return;
  var theme = btn.dataset.theme;
  dark = theme === "dark";

  // Update button states
  this.querySelectorAll("button").forEach(function(b) { b.classList.remove("active"); });
  btn.classList.add("active");

  // Update harness theme
  document.documentElement.setAttribute("data-theme", theme);

  // Send to chart iframe
  f.contentWindow.postMessage({
    jsonrpc: "2.0",
    method: "ui/notifications/host-context-changed",
    params: { theme: theme }
  }, "*");

  updateURL();
});

// Handle MCP Apps lifecycle messages from iframe
window.addEventListener("message", function(e) {
  var m = e.data;
  if (!m || !m.jsonrpc) return;
  if (m.method === "ui/initialize") {
    f.contentWindow.postMessage({
      jsonrpc: "2.0",
      id: m.id,
      result: { hostContext: { theme: dark ? "dark" : "light" } }
    }, "*");
  }
  if (m.method === "ui/notifications/initialized") {
    sendData();
  }
  if (m.method === "ui/notifications/size-changed") {
    var h = (m.params && m.params.height) || 0;
    f.style.height = h > 0 ? "calc(100vh - 120px)" : "0";
  }
});

// Sample metric definitions
var METRICS = [
  { name: "container_cpu_usage_seconds_total", labels: { namespace: "openshift-monitoring", pod: "prometheus-k8s-0" }, base: 4.5, amp: 1.0 },
  { name: "container_cpu_usage_seconds_total", labels: { namespace: "openshift-monitoring", pod: "prometheus-k8s-1" }, base: 0.7, amp: 0.3 },
  { name: "container_cpu_usage_seconds_total", labels: { namespace: "openshift-monitoring", pod: "node-exporter-k62n8" }, base: 0.5, amp: 0.4 },
  { name: "container_cpu_usage_seconds_total", labels: { namespace: "openshift-monitoring", pod: "node-exporter-cxgcx" }, base: 0.4, amp: 0.3 },
  { name: "container_cpu_usage_seconds_total", labels: { namespace: "openshift-monitoring", pod: "kube-apiserver-ip-10-0-7-32" }, base: 0.6, amp: 0.5 },
  { name: "container_cpu_usage_seconds_total", labels: { namespace: "openshift-monitoring", pod: "alertmanager-main-0" }, base: 1.2, amp: 0.4 },
  { name: "container_cpu_usage_seconds_total", labels: { namespace: "openshift-monitoring", pod: "thanos-querier-0" }, base: 0.8, amp: 0.3 },
  { name: "container_cpu_usage_seconds_total", labels: { namespace: "openshift-monitoring", pod: "grafana-7b9bc44b8f-x2k9m" }, base: 0.3, amp: 0.2 },
  { name: "container_cpu_usage_seconds_total", labels: { namespace: "openshift-monitoring", pod: "kube-state-metrics-5c9b7c4f6d-t8r2l" }, base: 0.9, amp: 0.35 },
  { name: "container_cpu_usage_seconds_total", labels: { namespace: "openshift-monitoring", pod: "telemeter-client-8d4f7b6c9-q3w5r" }, base: 0.2, amp: 0.15 },
];

function sendData() {
  var count = parseInt(document.getElementById("series-count").value);
  var range = parseInt(document.getElementById("time-range").value);
  var now = Math.floor(Date.now() / 1000);
  var start = now - range;
  var step = Math.max(15, Math.floor(range / 120)); // ~120 data points

  var selected = METRICS.slice(0, count);
  var result = selected.map(function(s) {
    var metric = Object.assign({ __name__: s.name }, s.labels);
    var base = s.base;
    var amp = s.amp;
    var values = [];
    for (var t = start; t <= now; t += step) {
      var progress = (t - start) / range;
      var trend = Math.sin(progress * Math.PI * 2) * amp;
      var noise = (Math.random() - 0.5) * amp * 0.3;
      var v = base + trend + noise;
      values.push([t, String(Math.max(0, v))]);
    }
    return { metric: metric, values: values };
  });

  var queryName = selected[0] ? selected[0].name : "up";
  var query = "topk(" + count + ", sum(rate(" + queryName + "[5m])) by (pod, namespace))";
  var title = document.getElementById("title-input").value.trim();

  updateURL();

  // Send tool-input (query + optional title/description)
  var description = document.getElementById("description-input").value.trim();
  var toolArgs = { query: query };
  if (title) toolArgs.title = title;
  if (description) toolArgs.description = description;
  f.contentWindow.postMessage({
    jsonrpc: "2.0",
    method: "ui/notifications/tool-input",
    params: { arguments: toolArgs }
  }, "*");

  // Send tool-result (data)
  var sc = { resultType: "matrix", result: result };
  f.contentWindow.postMessage({
    jsonrpc: "2.0",
    method: "ui/notifications/tool-result",
    params: { structuredContent: sc }
  }, "*");
}

function clearData() {
  f.contentWindow.postMessage({
    jsonrpc: "2.0",
    method: "ui/resource-teardown",
    params: {}
  }, "*");
  f.style.height = "0";
}
</script>
</body>
</html>
`

const traceHarness = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Trace Waterfall Test Harness</title>
<style>
  :root {
    --bg: #f3f4f6; --text: #111827; --surface: #ffffff;
    --border: #d1d5db; --subtle: #6b7280;
  }
  [data-theme="dark"] {
    --bg: #111827; --text: #f3f4f6; --surface: #1f2937;
    --border: #374151; --subtle: #9ca3af;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: system-ui, -apple-system, sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    transition: background 0.2s, color 0.2s;
  }
  .toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 20px;
    border-bottom: 1px solid var(--border);
    background: var(--surface);
    flex-wrap: wrap;
  }
  .toolbar h1 {
    font-size: 15px;
    font-weight: 600;
    margin-right: auto;
  }
  .nav-link {
    font-size: 13px;
    color: var(--subtle);
    text-decoration: none;
  }
  .nav-link:hover { color: var(--text); }
  .btn-group {
    display: flex;
    border: 1px solid var(--border);
    border-radius: 8px;
    overflow: hidden;
  }
  .btn-group button {
    padding: 6px 14px;
    font-size: 13px;
    font-family: inherit;
    background: var(--surface);
    color: var(--text);
    border: none;
    border-right: 1px solid var(--border);
    cursor: pointer;
    transition: background 0.15s;
  }
  .btn-group button:last-child { border-right: none; }
  .btn-group button:hover { background: var(--bg); }
  .btn-group button.active {
    background: #2563eb; color: #fff;
  }
  button.action {
    padding: 6px 16px;
    font-size: 13px;
    font-family: inherit;
    background: #2563eb;
    color: #fff;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: background 0.15s;
  }
  button.action:hover { background: #1d4ed8; }
  select {
    padding: 6px 10px;
    font-size: 13px;
    font-family: inherit;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--surface);
    color: var(--text);
    cursor: pointer;
  }
  label {
    font-size: 13px;
    color: var(--subtle);
  }
  .frame-area {
    padding: 20px;
  }
  iframe {
    width: 100%;
    height: 0;
    border: none;
    display: block;
    border-radius: 4px;
    transition: height 0.2s;
  }
</style>
</head>
<body>

<div class="toolbar">
  <h1>Trace Waterfall Test Harness</h1>
  <a class="nav-link" href="/">&larr; Timeseries Chart</a>

  <label>Theme</label>
  <div class="btn-group" id="theme-btns">
    <button class="active" data-theme="light">Light</button>
    <button data-theme="dark">Dark</button>
  </div>

  <label>Scenario</label>
  <select id="scenario">
    <option value="payment" selected>Payment Flow (8 spans)</option>
    <option value="simple">Simple (3 spans)</option>
    <option value="deep">Deep Nesting (12 spans)</option>
    <option value="error">With Errors</option>
  </select>

  <label>Title</label>
  <input type="text" id="title-input" value="Payment Processing Trace" placeholder="Chart title (optional)" style="padding:6px 10px;font-size:13px;font-family:inherit;border:1px solid var(--border);border-radius:8px;background:var(--surface);color:var(--text);width:200px">

  <label>Description</label>
  <input type="text" id="description-input" value="Full request flow for POST /api/v1/payments/create" placeholder="Description (optional)" style="padding:6px 10px;font-size:13px;font-family:inherit;border:1px solid var(--border);border-radius:8px;background:var(--surface);color:var(--text);width:260px">

  <button class="action" onclick="sendData()">Send Data</button>
  <button class="action" onclick="clearData()" style="background:#dc2626">Clear</button>
</div>

<div class="frame-area">
  <iframe id="f" src="/trace-app"></iframe>
</div>

<script>
var dark = false;
var f = document.getElementById("f");

// Restore from URL
(function() {
  var p = new URLSearchParams(window.location.search);
  if (p.has("theme")) {
    var t = p.get("theme");
    if (t === "dark" || t === "light") {
      dark = t === "dark";
      document.documentElement.setAttribute("data-theme", t);
      document.querySelectorAll("#theme-btns button").forEach(function(b) {
        b.classList.toggle("active", b.dataset.theme === t);
      });
    }
  }
  if (p.has("scenario")) {
    var sel = document.getElementById("scenario");
    if (sel.querySelector('option[value="' + p.get("scenario") + '"]')) sel.value = p.get("scenario");
  }
  if (p.has("title")) {
    document.getElementById("title-input").value = p.get("title");
  }
  if (p.has("description")) {
    document.getElementById("description-input").value = p.get("description");
  }
})();

function updateURL() {
  var p = new URLSearchParams();
  p.set("theme", dark ? "dark" : "light");
  p.set("scenario", document.getElementById("scenario").value);
  var title = document.getElementById("title-input").value.trim();
  if (title) p.set("title", title);
  var description = document.getElementById("description-input").value.trim();
  if (description) p.set("description", description);
  history.replaceState(null, "", "?" + p.toString());
}

// Theme switching
document.getElementById("theme-btns").addEventListener("click", function(e) {
  var btn = e.target.closest("button");
  if (!btn) return;
  var theme = btn.dataset.theme;
  dark = theme === "dark";
  this.querySelectorAll("button").forEach(function(b) { b.classList.remove("active"); });
  btn.classList.add("active");
  document.documentElement.setAttribute("data-theme", theme);
  f.contentWindow.postMessage({
    jsonrpc: "2.0",
    method: "ui/notifications/host-context-changed",
    params: { theme: theme }
  }, "*");
  updateURL();
});

// Handle MCP Apps lifecycle
window.addEventListener("message", function(e) {
  var m = e.data;
  if (!m || !m.jsonrpc) return;
  if (m.method === "ui/initialize") {
    f.contentWindow.postMessage({
      jsonrpc: "2.0",
      id: m.id,
      result: { hostContext: { theme: dark ? "dark" : "light" } }
    }, "*");
  }
  if (m.method === "ui/notifications/initialized") {
    sendData();
  }
  if (m.method === "ui/notifications/size-changed") {
    var h = (m.params && m.params.height) || 0;
    if (h > 0) f.style.height = h + "px";
  }
});

// Sample trace scenarios
var SCENARIOS = {
  payment: {
    spans: [
      { spanId: "span-1", parentSpanId: "", operationName: "POST /api/v1/payments/create", serviceName: "api-gateway", startTimeMs: 0, durationMs: 1850, statusCode: 0 },
      { spanId: "span-2", parentSpanId: "span-1", operationName: "AuthMiddleware", serviceName: "api-gateway", startTimeMs: 10, durationMs: 85, statusCode: 0 },
      { spanId: "span-3", parentSpanId: "span-1", operationName: "ProcessPayment", serviceName: "payment-service", startTimeMs: 100, durationMs: 1200, statusCode: 0 },
      { spanId: "span-4", parentSpanId: "span-3", operationName: "ValidatePaymentData", serviceName: "payment-service", startTimeMs: 110, durationMs: 45, statusCode: 0 },
      { spanId: "span-5", parentSpanId: "span-3", operationName: "GetUserProfile", serviceName: "user-service", startTimeMs: 200, durationMs: 950, statusCode: 0 },
      { spanId: "span-6", parentSpanId: "span-5", operationName: "DB::GetUser", serviceName: "user-service", startTimeMs: 210, durationMs: 900, statusCode: 0 },
      { spanId: "span-7", parentSpanId: "span-3", operationName: "RecordPaymentAttempt", serviceName: "payment-service", startTimeMs: 1200, durationMs: 80, statusCode: 0 },
      { spanId: "span-8", parentSpanId: "span-1", operationName: "SerializeResponse", serviceName: "api-gateway", startTimeMs: 1350, durationMs: 30, statusCode: 0 }
    ]
  },
  simple: {
    spans: [
      { spanId: "s1", parentSpanId: "", operationName: "GET /api/users", serviceName: "api-gateway", startTimeMs: 0, durationMs: 250, statusCode: 0 },
      { spanId: "s2", parentSpanId: "s1", operationName: "FetchUsers", serviceName: "user-service", startTimeMs: 20, durationMs: 180, statusCode: 0 },
      { spanId: "s3", parentSpanId: "s2", operationName: "SELECT * FROM users", serviceName: "postgres", startTimeMs: 30, durationMs: 120, statusCode: 0 }
    ]
  },
  deep: {
    spans: [
      { spanId: "d1", parentSpanId: "", operationName: "POST /checkout", serviceName: "frontend", startTimeMs: 0, durationMs: 3200, statusCode: 0 },
      { spanId: "d2", parentSpanId: "d1", operationName: "ValidateCart", serviceName: "cart-service", startTimeMs: 50, durationMs: 400, statusCode: 0 },
      { spanId: "d3", parentSpanId: "d2", operationName: "GetCartItems", serviceName: "cart-service", startTimeMs: 60, durationMs: 150, statusCode: 0 },
      { spanId: "d4", parentSpanId: "d2", operationName: "CheckInventory", serviceName: "inventory-service", startTimeMs: 220, durationMs: 200, statusCode: 0 },
      { spanId: "d5", parentSpanId: "d4", operationName: "Redis::MGET", serviceName: "redis", startTimeMs: 230, durationMs: 30, statusCode: 0 },
      { spanId: "d6", parentSpanId: "d1", operationName: "ProcessPayment", serviceName: "payment-service", startTimeMs: 500, durationMs: 1800, statusCode: 0 },
      { spanId: "d7", parentSpanId: "d6", operationName: "ChargeCard", serviceName: "stripe-client", startTimeMs: 520, durationMs: 1500, statusCode: 0 },
      { spanId: "d8", parentSpanId: "d7", operationName: "HTTP POST /v1/charges", serviceName: "stripe-client", startTimeMs: 530, durationMs: 1400, statusCode: 0 },
      { spanId: "d9", parentSpanId: "d6", operationName: "SaveTransaction", serviceName: "payment-service", startTimeMs: 2050, durationMs: 200, statusCode: 0 },
      { spanId: "d10", parentSpanId: "d9", operationName: "INSERT INTO transactions", serviceName: "postgres", startTimeMs: 2060, durationMs: 150, statusCode: 0 },
      { spanId: "d11", parentSpanId: "d1", operationName: "SendConfirmation", serviceName: "notification-service", startTimeMs: 2400, durationMs: 600, statusCode: 0 },
      { spanId: "d12", parentSpanId: "d11", operationName: "SendEmail", serviceName: "email-service", startTimeMs: 2420, durationMs: 500, statusCode: 0 }
    ]
  },
  error: {
    spans: [
      { spanId: "e1", parentSpanId: "", operationName: "POST /api/orders", serviceName: "api-gateway", startTimeMs: 0, durationMs: 800, statusCode: 2 },
      { spanId: "e2", parentSpanId: "e1", operationName: "AuthMiddleware", serviceName: "api-gateway", startTimeMs: 10, durationMs: 50, statusCode: 0 },
      { spanId: "e3", parentSpanId: "e1", operationName: "CreateOrder", serviceName: "order-service", startTimeMs: 70, durationMs: 700, statusCode: 2 },
      { spanId: "e4", parentSpanId: "e3", operationName: "ValidateOrder", serviceName: "order-service", startTimeMs: 80, durationMs: 30, statusCode: 0 },
      { spanId: "e5", parentSpanId: "e3", operationName: "ReserveInventory", serviceName: "inventory-service", startTimeMs: 120, durationMs: 600, statusCode: 2 },
      { spanId: "e6", parentSpanId: "e5", operationName: "DB::UPDATE inventory", serviceName: "postgres", startTimeMs: 130, durationMs: 550, statusCode: 2 }
    ]
  }
};

function sendData() {
  var scenario = document.getElementById("scenario").value;
  var data = SCENARIOS[scenario];
  updateURL();

  // Offset startTimeMs to current time
  var now = Date.now();
  var spans = data.spans.map(function(s) {
    return Object.assign({}, s, { startTimeMs: now - 5000 + s.startTimeMs });
  });

  // Set initial iframe height so content can render and ResizeObserver can measure
  f.style.height = "600px";

  // Send tool-input (title + description)
  var title = document.getElementById("title-input").value.trim();
  var description = document.getElementById("description-input").value.trim();
  var toolArgs = {};
  if (title) toolArgs.title = title;
  if (description) toolArgs.description = description;
  f.contentWindow.postMessage({
    jsonrpc: "2.0",
    method: "ui/notifications/tool-input",
    params: { arguments: toolArgs }
  }, "*");

  // Send tool-result (data)
  f.contentWindow.postMessage({
    jsonrpc: "2.0",
    method: "ui/notifications/tool-result",
    params: { structuredContent: { spans: spans } }
  }, "*");
}

function clearData() {
  f.contentWindow.postMessage({
    jsonrpc: "2.0",
    method: "ui/resource-teardown",
    params: {}
  }, "*");
  f.style.height = "0";
}
</script>
</body>
</html>
`
