package expr

import (
	"context"
	"fmt"
)

// Simple evaluator supports checking equality for payload fields with syntax "payload.key == value"
type SimpleExpr struct{}

func NewSimpleExpr() *SimpleExpr { return &SimpleExpr{} }

func (e *SimpleExpr) EvalBool(ctx context.Context, ex string, payload map[string]any) (bool, error) {
	// Example: "payload.status == ok" (no quotes)
	var key, op, val string
	_, err := fmt.Sscanf(ex, "payload.%s %s %s", &key, &op, &val)
	if err != nil {
		return false, err
	}
	// remove potential punctuation from key/val
	_ = key
	_ = val
	if op != "==" {
		return false, fmt.Errorf("unsupported op %s", op)
	}
	v, ok := payload[key]
	return ok && fmt.Sprint(v) == val, nil
}
