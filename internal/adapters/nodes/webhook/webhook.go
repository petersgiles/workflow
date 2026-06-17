package webhook

import (
	"context"

	"local.com/internal/flow"
)

type WebhookNode struct {
	id     string
	url    string
	client flow.HTTPClient
}

func New(id, url string, client flow.HTTPClient) *WebhookNode {
	return &WebhookNode{id: id, url: url, client: client}
}

func (n *WebhookNode) ID() string { return n.id }

func (n *WebhookNode) Enter(ctx context.Context, in flow.Input) (flow.State, error) {
	return flow.State{Data: map[string]any{}, Input: in}, nil
}

func (n *WebhookNode) Process(ctx context.Context, s flow.State) (flow.Result, error) {
	status, body, err := n.client.PostJSON(ctx, n.url, s.Input.Payload)
	data := map[string]any{"status": status, "response": string(body)}
	sig := flow.ControlSignal{}
	if err != nil || status >= 400 {
		sig.Abort = true
	}
	return flow.Result{Data: data, Signal: sig}, err
}

func (n *WebhookNode) Exit(ctx context.Context, s flow.State) (flow.FinalResult, error) {
	return flow.FinalResult{Data: s.Data}, nil
}
