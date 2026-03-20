package tempo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOTLPTrace(t *testing.T) {
	otlpJSON := `{
		"trace": {
			"resourceSpans": [
				{
					"resource": {
						"attributes": [
							{"key": "service.name", "value": {"stringValue": "api-gateway"}}
						]
					},
					"scopeSpans": [
						{
							"spans": [
								{
									"traceId": "JtrdoOKw3ZpEDdX/IDokok==",
									"spanId": "AAAAAAAAAAE=",
									"parentSpanId": "",
									"name": "POST /api/v1/payments/create",
									"kind": "SPAN_KIND_SERVER",
									"startTimeUnixNano": "1706000000000000000",
									"endTimeUnixNano": "1706000001850000000",
									"status": {}
								},
								{
									"traceId": "JtrdoOKw3ZpEDdX/IDokok==",
									"spanId": "AAAAAAAAAAI=",
									"parentSpanId": "AAAAAAAAAAE=",
									"name": "AuthMiddleware",
									"kind": "SPAN_KIND_CLIENT",
									"startTimeUnixNano": "1706000000050000000",
									"endTimeUnixNano": "1706000000135000000",
									"status": {}
								}
							]
						}
					]
				},
				{
					"resource": {
						"attributes": [
							{"key": "service.name", "value": {"stringValue": "payment-service"}}
						]
					},
					"scopeSpans": [
						{
							"spans": [
								{
									"traceId": "JtrdoOKw3ZpEDdX/IDokok==",
									"spanId": "AAAAAAAAAAM=",
									"parentSpanId": "AAAAAAAAAAE=",
									"name": "ProcessPayment",
									"kind": "SPAN_KIND_SERVER",
									"startTimeUnixNano": "1706000000150000000",
									"endTimeUnixNano": "1706000001350000000",
									"status": {"code": "STATUS_CODE_ERROR"}
								}
							]
						}
					]
				}
			]
		}
	}`

	output, err := parseOTLPTrace(otlpJSON)
	require.NoError(t, err)

	require.Len(t, output.Spans, 3)

	// First span: root span from api-gateway
	require.Equal(t, "AAAAAAAAAAE=", output.Spans[0].SpanID)
	require.Equal(t, "", output.Spans[0].ParentSpanID)
	require.Equal(t, "POST /api/v1/payments/create", output.Spans[0].OperationName)
	require.Equal(t, "api-gateway", output.Spans[0].ServiceName)
	require.InDelta(t, 1706000000000.0, output.Spans[0].StartTimeMs, 0.1)
	require.InDelta(t, 1850.0, output.Spans[0].DurationMs, 0.1)
	require.Equal(t, "", output.Spans[0].StatusCode)

	// Second span: child of root, same service
	require.Equal(t, "AAAAAAAAAAI=", output.Spans[1].SpanID)
	require.Equal(t, "AAAAAAAAAAE=", output.Spans[1].ParentSpanID)
	require.Equal(t, "AuthMiddleware", output.Spans[1].OperationName)
	require.Equal(t, "api-gateway", output.Spans[1].ServiceName)
	require.InDelta(t, 85.0, output.Spans[1].DurationMs, 0.1)

	// Third span: different service
	require.Equal(t, "AAAAAAAAAAM=", output.Spans[2].SpanID)
	require.Equal(t, "AAAAAAAAAAE=", output.Spans[2].ParentSpanID)
	require.Equal(t, "ProcessPayment", output.Spans[2].OperationName)
	require.Equal(t, "payment-service", output.Spans[2].ServiceName)
	require.InDelta(t, 1200.0, output.Spans[2].DurationMs, 0.1)
	require.Equal(t, "STATUS_CODE_ERROR", output.Spans[2].StatusCode)
}

func TestParseOTLPTrace_EmptyBatches(t *testing.T) {
	output, err := parseOTLPTrace(`{"trace": {"resourceSpans": []}}`)
	require.NoError(t, err)
	require.Empty(t, output.Spans)
}

func TestParseOTLPTrace_InvalidJSON(t *testing.T) {
	_, err := parseOTLPTrace(`not json`)
	require.Error(t, err)
}

func TestParseOTLPTrace_MissingServiceName(t *testing.T) {
	otlpJSON := `{
		"trace": {
			"resourceSpans": [{
				"resource": {"attributes": []},
				"scopeSpans": [{
					"spans": [{
						"traceId": "abc",
						"spanId": "def",
						"name": "test-op",
						"startTimeUnixNano": "1000000000000",
						"endTimeUnixNano": "2000000000000",
						"status": {}
					}]
				}]
			}]
		}
	}`

	output, err := parseOTLPTrace(otlpJSON)
	require.NoError(t, err)
	require.Len(t, output.Spans, 1)
	require.Equal(t, "", output.Spans[0].ServiceName)
	require.Equal(t, "test-op", output.Spans[0].OperationName)
	require.InDelta(t, 1000000.0, output.Spans[0].DurationMs, 0.1)
}
