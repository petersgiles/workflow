package finddevice

import (
	"context"

	"local.com/internal/flow"
)

type FindDeviceNode struct {
	id     string
	target string
	trans  map[string]string
}

func New(id string, target string, trans map[string]string) *FindDeviceNode {
	return &FindDeviceNode{id: id, target: target, trans: trans}
}

func (f *FindDeviceNode) ID() string { return f.id }

func (f *FindDeviceNode) Enter(ctx context.Context, in flow.Input) (flow.State, error) {
	return flow.State{Data: map[string]any{}, Input: in}, nil
}

func (f *FindDeviceNode) Process(ctx context.Context, s flow.State) (flow.Result, error) {
	var found any
	if devicesI, ok := s.Input.Payload["devices"]; ok {
		switch devs := devicesI.(type) {
		case []any:
			for _, di := range devs {
				if m, ok := di.(map[string]any); ok {
					if name, ok := m["name"].(string); ok && name == f.target {
						found = m
						break
					}
				}
				if m2, ok := di.(map[string]string); ok {
					if name, ok := m2["name"]; ok && name == f.target {
						// convert to map[string]any
						mm := map[string]any{"name": m2["name"], "address": m2["address"]}
						found = mm
						break
					}
				}
			}
		case []map[string]string:
			for _, m := range devs {
				if name, ok := m["name"]; ok && name == f.target {
					mm := map[string]any{"name": m["name"], "address": m["address"]}
					found = mm
					break
				}
			}
		}
	}

	sig := flow.ControlSignal{}
	data := map[string]any{"found": found}
	// optionally set Next via transitions: "found" or "not_found"
	if found != nil {
		if nxt, ok := f.trans["found"]; ok {
			sig.Next = nxt
		}
	} else {
		if nxt, ok := f.trans["not_found"]; ok {
			sig.Next = nxt
		}
	}
	return flow.Result{Data: data, Signal: sig}, nil
}

func (f *FindDeviceNode) Exit(ctx context.Context, s flow.State) (flow.FinalResult, error) {
	return flow.FinalResult{Data: s.Data}, nil
}
