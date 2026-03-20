package mcp

import (
	_ "embed"
	"strings"
)

//go:embed ui/chart.html
var chartTemplate string

//go:embed ui/styles.css
var chartStyles string

//go:embed ui/chart.min.js
var chartLib string

//go:embed ui/date-adapter.js
var dateAdapter string

//go:embed ui/app.js
var chartApp string

//go:embed ui/trace.html
var traceTemplate string

//go:embed ui/trace.css
var traceStyles string

//go:embed ui/trace.js
var traceApp string

var chartHTML = buildChartHTML()
var traceHTML = buildTraceHTML()

func buildChartHTML() string {
	r := strings.NewReplacer(
		"{{STYLES}}", chartStyles,
		"{{CHART_LIB}}", chartLib,
		"{{DATE_ADAPTER}}", dateAdapter,
		"{{APP}}", chartApp,
	)
	return r.Replace(chartTemplate)
}

func buildTraceHTML() string {
	r := strings.NewReplacer(
		"{{TRACE_STYLES}}", traceStyles,
		"{{TRACE_APP}}", traceApp,
	)
	return r.Replace(traceTemplate)
}
