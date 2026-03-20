package tempo

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rhobs/obs-mcp/pkg/resultutil"
	tempoclient "github.com/rhobs/obs-mcp/pkg/tempo/client"
	"github.com/rhobs/obs-mcp/pkg/tools"
)

var ShowTraceTool = tools.ToolDef{
	Name: "tempo_show_trace",
	Description: `Retrieve a distributed trace by its trace ID and display it as an interactive waterfall chart.
This tool works like tempo_get_trace_by_id but renders the trace as a visual Gantt/waterfall chart in MCP Apps-capable clients.
Use this when the user wants to visualize a trace's span hierarchy, timing, and service dependencies.`,
	Title: "Show Trace Waterfall",
	Params: []tools.ParamDef{
		tempoNamespaceParameter,
		tempoNameParameter,
		tempoTenantParameter,
		{
			Name:        "traceid",
			Type:        tools.ParamTypeString,
			Description: `The trace ID to retrieve, e.g. "26dad4a0e2b0dd9a440dd5ff203a24a4".`,
			Required:    true,
		},
		{
			Name: "start",
			Type: tools.ParamTypeString,
			Description: `Optional start of the time range in RFC 3339 format, e.g. "2025-01-01T00:00:00Z".
Narrows the time range to improve query performance.`,
		},
		{
			Name: "end",
			Type: tools.ParamTypeString,
			Description: `Optional end of the time range in RFC 3339 format, e.g. "2025-01-02T00:00:00Z".
Narrows the time range to improve query performance.`,
		},
		{
			Name:        "title",
			Type:        tools.ParamTypeString,
			Description: "A descriptive chart title (e.g., \"Payment Processing Trace\")",
		},
		{
			Name:        "description",
			Type:        tools.ParamTypeString,
			Description: "An explanation of the trace's meaning or context (e.g., \"Shows the full request flow for a payment transaction\")",
		},
	},
	ReadOnly:    true,
	Destructive: false,
	Idempotent:  true,
	OpenWorld:   true,
}

// TraceOutput is the structured output sent to the MCP App for rendering.
type TraceOutput struct {
	Spans []SpanOutput `json:"spans"`
}

// SpanOutput represents a single span in a simplified format for visualization.
type SpanOutput struct {
	SpanID        string  `json:"spanId"`
	ParentSpanID  string  `json:"parentSpanId"`
	OperationName string  `json:"operationName"`
	ServiceName   string  `json:"serviceName"`
	StartTimeMs   float64 `json:"startTimeMs"`
	DurationMs    float64 `json:"durationMs"`
	StatusCode    string  `json:"statusCode"`
}

func (t *Toolset) ShowTraceHandler(params ToolParams) *resultutil.Result {
	client, err := t.getTempoClient(params)
	if err != nil {
		return resultutil.NewErrorResult(err)
	}

	args := params.arguments

	traceid := tools.GetString(args, "traceid", "")
	if traceid == "" {
		return resultutil.NewErrorResult(fmt.Errorf("traceid parameter must not be empty"))
	}

	start, err := parseDate(tools.GetString(args, "start", ""))
	if err != nil {
		return resultutil.NewErrorResult(fmt.Errorf("invalid start time: %v", err))
	}

	end, err := parseDate(tools.GetString(args, "end", ""))
	if err != nil {
		return resultutil.NewErrorResult(fmt.Errorf("invalid end time: %v", err))
	}

	opts := tempoclient.QueryV2Options{
		Start: start,
		End:   end,
	}

	traceJSON, err := client.QueryV2JSON(params.context, traceid, opts)
	if err != nil {
		return resultutil.NewErrorResult(err)
	}

	output, err := parseOTLPTrace(traceJSON)
	if err != nil {
		return resultutil.NewErrorResult(fmt.Errorf("failed to parse trace: %v", err))
	}

	return resultutil.NewSuccessResult(output)
}

// parseOTLPTrace transforms an OTLP JSON trace response into the simplified TraceOutput format.
func parseOTLPTrace(otlpJSON string) (*TraceOutput, error) {
	var envelope struct {
		Trace struct {
			ResourceSpans []struct {
				Resource struct {
					Attributes []struct {
						Key   string `json:"key"`
						Value struct {
							StringValue string `json:"stringValue"`
						} `json:"value"`
					} `json:"attributes"`
				} `json:"resource"`
				ScopeSpans []struct {
					Spans []struct {
						TraceID           string `json:"traceId"`
						SpanID            string `json:"spanId"`
						ParentSpanID      string `json:"parentSpanId"`
						Name              string `json:"name"`
						Kind              string `json:"kind"`
						StartTimeUnixNano string `json:"startTimeUnixNano"`
						EndTimeUnixNano   string `json:"endTimeUnixNano"`
						Status struct {
							Code string `json:"code"`
						} `json:"status"`
					} `json:"spans"`
				} `json:"scopeSpans"`
			} `json:"resourceSpans"`
		} `json:"trace"`
	}

	if err := json.Unmarshal([]byte(otlpJSON), &envelope); err != nil {
		return nil, err
	}

	output := &TraceOutput{}

	for _, batch := range envelope.Trace.ResourceSpans {
		serviceName := ""
		for _, attr := range batch.Resource.Attributes {
			if attr.Key == "service.name" {
				serviceName = attr.Value.StringValue
				break
			}
		}

		for _, scopeSpan := range batch.ScopeSpans {
			for _, span := range scopeSpan.Spans {
				startNano, _ := strconv.ParseFloat(span.StartTimeUnixNano, 64)
				endNano, _ := strconv.ParseFloat(span.EndTimeUnixNano, 64)

				output.Spans = append(output.Spans, SpanOutput{
					SpanID:        span.SpanID,
					ParentSpanID:  span.ParentSpanID,
					OperationName: span.Name,
					ServiceName:   serviceName,
					StartTimeMs:   startNano / 1e6,
					DurationMs:    (endNano - startNano) / 1e6,
					StatusCode:    span.Status.Code,
				})
			}
		}
	}

	return output, nil
}
