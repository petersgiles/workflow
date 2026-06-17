package flow

import (
	"context"
	"time"
)

// Scheduler runs a function on a cron-like schedule. Minimal interface.
type Scheduler interface {
	Start(ctx context.Context) error
	Stop() error
	Schedule(cronExpr string, job func(ctx context.Context)) error
}

// ScriptRunner executes commands.
type ScriptRunner interface {
	Exec(ctx context.Context, cmd string, args []string, env map[string]string) (stdout string, stderr string, exit int, err error)
}

// ExprEvaluator safely evaluates a boolean expression against payload.
type ExprEvaluator interface {
	EvalBool(ctx context.Context, expr string, payload map[string]any) (bool, error)
}

// HTTPClient sends outbound requests (minimal).
type HTTPClient interface {
	PostJSON(ctx context.Context, url string, body any) (status int, respBody []byte, err error)
}

// Console logger
type Console interface {
	Printf(format string, v ...any)
}

// BTDevice represents a discovered Bluetooth device.
type BTDevice struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

// BluetoothScanner scans for nearby Bluetooth devices.
type BluetoothScanner interface {
	Scan(ctx context.Context, timeout time.Duration) ([]BTDevice, error)
}
