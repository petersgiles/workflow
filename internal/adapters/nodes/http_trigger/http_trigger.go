package http_trigger

import (
	"context"
	"encoding/json"
	"net/http"

	"local.com/internal/flow"
)

// HandlerAdapter holds orchestrator runner callback to start flows.
type HandlerAdapter struct {
	startNode string
	orchRun   func(ctx context.Context, startNode string, input flow.Input)
}

func New(startNode string, orchRun func(ctx context.Context, startNode string, input flow.Input)) *HandlerAdapter {
	return &HandlerAdapter{startNode: startNode, orchRun: orchRun}
}

// ServeHTTP implements http.Handler.
func (h *HandlerAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	_ = json.NewDecoder(r.Body).Decode(&payload) // ignore error for brevity
	go h.orchRun(r.Context(), h.startNode, flow.Input{Payload: payload})
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"started"}`))
}
