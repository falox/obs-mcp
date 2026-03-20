package mcp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/client-go/dynamic"

	"github.com/rhobs/obs-mcp/pkg/k8s"
	"github.com/rhobs/obs-mcp/pkg/prometheus"
	"github.com/rhobs/obs-mcp/pkg/tempo"
	"github.com/rhobs/obs-mcp/pkg/tools"
)

// ObsMCPOptions contains configuration options for the MCP server
type ObsMCPOptions struct {
	AuthMode               AuthMode
	Toolsets               []string
	MetricsBackendURL      string
	AlertmanagerURL        string
	Insecure               bool
	Guardrails             *prometheus.Guardrails
	FullRangeQueryResponse bool
}

const (
	mcpEndpoint            = "/mcp"
	healthEndpoint         = "/health"
	serverName             = "obs-mcp"
	serverVersion          = "1.0.0"
	defaultShutdownTimeout = 10 * time.Second
)

func NewMCPServer(opts ObsMCPOptions) (*server.MCPServer, error) {
	hooks := &server.Hooks{}
	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		slog.Debug("MCP tool call", "tool", message.Params.Name, "arguments", message.Params.Arguments)
	})
	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result any) {
		slog.Debug("MCP tool result", "tool", message.Params.Name, "result", result)
	})

	mcpServer := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithLogging(),
		server.WithHooks(hooks),
		server.WithToolCapabilities(true),
		server.WithInstructions(tools.ServerPrompt),
	)

	if err := SetupTools(mcpServer, opts); err != nil {
		return nil, err
	}

	// Register UI resources for MCP Apps
	mcpServer.AddResource(
		mcp.Resource{
			URI:         "ui://timeseries-chart",
			Name:        "Timeseries Chart",
			Description: "Interactive timeseries line chart for range query results",
			MIMEType:    "text/html;profile=mcp-app",
		},
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "ui://timeseries-chart",
					MIMEType: "text/html;profile=mcp-app",
					Text:     chartHTML,
				},
			}, nil
		},
	)

	// Register trace waterfall UI resource for MCP Apps
	mcpServer.AddResource(
		mcp.Resource{
			URI:         "ui://trace-waterfall",
			Name:        "Trace Waterfall",
			Description: "Interactive waterfall chart for distributed trace spans",
			MIMEType:    "text/html;profile=mcp-app",
		},
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "ui://trace-waterfall",
					MIMEType: "text/html;profile=mcp-app",
					Text:     traceHTML,
				},
			}, nil
		},
	)

	return mcpServer, nil
}

func SetupTools(mcpServer *server.MCPServer, opts ObsMCPOptions) error {
	if slices.Contains(opts.Toolsets, "prometheus") {
		// Create tool definitions
		listMetricsTool := CreateListMetricsTool()
		executeInstantQueryTool := CreateExecuteInstantQueryTool()
		executeRangeQueryTool := CreateExecuteRangeQueryTool()
		showTimeseriesTool := CreateShowTimeseriesTool()
		getLabelNamesTool := CreateGetLabelNamesTool()
		getLabelValuesTool := CreateGetLabelValuesTool()
		getSeriesTool := CreateGetSeriesTool()
		getAlertsTool := CreateGetAlertsTool()
		getSilencesTool := CreateGetSilencesTool()

		// Create handlers
		listMetricsHandler := ListMetricsHandler(opts)
		executeInstantQueryHandler := ExecuteInstantQueryHandler(opts)
		executeRangeQueryHandler := ExecuteRangeQueryHandler(opts)
		showTimeseriesHandler := ShowTimeseriesHandler(opts)
		getLabelNamesHandler := GetLabelNamesHandler(opts)
		getLabelValuesHandler := GetLabelValuesHandler(opts)
		getSeriesHandler := GetSeriesHandler(opts)
		getAlertsHandler := GetAlertsHandler(opts)
		getSilencesHandler := GetSilencesHandler(opts)

		// Add tools to server
		mcpServer.AddTool(listMetricsTool, listMetricsHandler)
		mcpServer.AddTool(executeInstantQueryTool, executeInstantQueryHandler)
		mcpServer.AddTool(executeRangeQueryTool, executeRangeQueryHandler)
		mcpServer.AddTool(showTimeseriesTool, showTimeseriesHandler)
		mcpServer.AddTool(getLabelNamesTool, getLabelNamesHandler)
		mcpServer.AddTool(getLabelValuesTool, getLabelValuesHandler)
		mcpServer.AddTool(getSeriesTool, getSeriesHandler)
		mcpServer.AddTool(getAlertsTool, getAlertsHandler)
		mcpServer.AddTool(getSilencesTool, getSilencesHandler)
	}

	if slices.Contains(opts.Toolsets, "tempo") {
		tempoToolset := &tempo.Toolset{}
		restConfig, err := k8s.GetClientConfig()
		if err != nil {
			return err
		}
		dynamicClient, err := dynamic.NewForConfig(restConfig)
		if err != nil {
			return err
		}
		mcpServer.AddTool(tempo.ListInstancesTool.ToMCPTool(), tempo.ToMCPHandler(restConfig, dynamicClient, tempoToolset.ListInstancesHandler))
		mcpServer.AddTool(tempo.GetTraceByIDTool.ToMCPTool(), tempo.ToMCPHandler(restConfig, dynamicClient, tempoToolset.GetTraceByIDHandler))
		mcpServer.AddTool(CreateShowTraceTool(), tempo.ToMCPHandler(restConfig, dynamicClient, tempoToolset.ShowTraceHandler))
		mcpServer.AddTool(tempo.SearchTracesTool.ToMCPTool(), tempo.ToMCPHandler(restConfig, dynamicClient, tempoToolset.SearchTracesHandler))
		mcpServer.AddTool(tempo.SearchTagsTool.ToMCPTool(), tempo.ToMCPHandler(restConfig, dynamicClient, tempoToolset.SearchTagsHandler))
		mcpServer.AddTool(tempo.SearchTagValuesTool.ToMCPTool(), tempo.ToMCPHandler(restConfig, dynamicClient, tempoToolset.SearchTagValuesHandler))
	}

	return nil
}

func authFromRequest(ctx context.Context, r *http.Request) context.Context {
	authHeaderValue := r.Header.Get(string(AuthHeaderKey))
	token, found := strings.CutPrefix(authHeaderValue, "Bearer ")
	if !found {
		return ctx
	}
	return context.WithValue(ctx, AuthHeaderKey, token)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Incoming request", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
		slog.Debug("Request headers", "headers", r.Header)
		if r.ContentLength > 0 {
			slog.Info("Request content length", "content_length", r.ContentLength)
		}
		next.ServeHTTP(w, r)
	})
}

func Serve(ctx context.Context, mcpServer *server.MCPServer, listenAddr string) error {
	mux := http.NewServeMux()

	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: loggingMiddleware(mux),
	}

	streamableHTTPServer := server.NewStreamableHTTPServer(mcpServer,
		server.WithStreamableHTTPServer(httpServer),
		server.WithStateLess(true),
		server.WithHTTPContextFunc(authFromRequest),
	)
	mux.Handle(mcpEndpoint, streamableHTTPServer)

	mux.Handle("/", streamableHTTPServer)

	mux.HandleFunc(healthEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("HTTP server starting", "listen_addr", listenAddr, "mcp_endpoint", mcpEndpoint)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case sig := <-sigChan:
		slog.Warn("Received signal, initiating graceful shutdown", "signal", sig)
		cancel()
	case <-ctx.Done():
		slog.Warn("Context cancelled, initiating graceful shutdown")
	case err := <-serverErr:
		slog.Error("HTTP server error", "error", err)
		return err
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer shutdownCancel()

	slog.Info("Shutting down HTTP server gracefully")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
		return err
	}

	slog.Info("HTTP server shutdown complete")
	return nil
}
