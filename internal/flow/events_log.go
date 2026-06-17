package flow

import (
	"context"
	"log"
)

type LogEvents struct{}

func (LogEvents) Info(ctx context.Context, nodeID string, payload map[string]any) {
	log.Printf("[info] node=%s payload=%v", nodeID, payload)
}
func (LogEvents) Success(ctx context.Context, nodeID string, payload map[string]any) {
	log.Printf("[success] node=%s payload=%v", nodeID, payload)
}
func (LogEvents) Error(ctx context.Context, nodeID string, err error, payload map[string]any) {
	log.Printf("[error] node=%s err=%v payload=%v", nodeID, err, payload)
}
