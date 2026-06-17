package bluetooth

import (
	"context"
	"time"

	"local.com/internal/flow"
)

type BluetoothNode struct {
	id      string
	scanner flow.BluetoothScanner
	timeout time.Duration
}

func New(id string, scanner flow.BluetoothScanner, timeout time.Duration) *BluetoothNode {
	return &BluetoothNode{id: id, scanner: scanner, timeout: timeout}
}

func (b *BluetoothNode) ID() string { return b.id }

func (b *BluetoothNode) Enter(ctx context.Context, in flow.Input) (flow.State, error) {
	return flow.State{Data: map[string]any{}, Input: in}, nil
}

func (b *BluetoothNode) Process(ctx context.Context, s flow.State) (flow.Result, error) {
	devs, err := b.scanner.Scan(ctx, b.timeout)
	if err != nil {
		return flow.Result{}, err
	}
	// Convert to slice of maps for YAML/JSON friendliness
	out := make([]map[string]string, 0, len(devs))
	for _, d := range devs {
		out = append(out, map[string]string{"address": d.Address, "name": d.Name})
	}
	data := map[string]any{"devices": out}
	return flow.Result{Data: data, Signal: flow.ControlSignal{}}, nil
}

func (b *BluetoothNode) Exit(ctx context.Context, s flow.State) (flow.FinalResult, error) {
	return flow.FinalResult{Data: s.Data}, nil
}
