package schedule

import (
	"context"
	"fmt"
	"time"

	"local.com/internal/flow"
)

// SimpleCron is a trivial scheduler using time.Ticker for demo only (not full cron).
type SimpleCron struct {
	entries []struct {
		interval time.Duration
		job      func(ctx context.Context)
	}
	running bool
	cancel  context.CancelFunc
}

func NewSimpleCron() *SimpleCron { return &SimpleCron{} }

func (c *SimpleCron) Schedule(expr string, job func(ctx context.Context)) error {
	// for demo interpret expr as seconds: "5s", "10s"
	d, err := time.ParseDuration(expr)
	if err != nil {
		return fmt.Errorf("invalid duration expr (use Go duration like 5s): %w", err)
	}
	c.entries = append(c.entries, struct {
		interval time.Duration
		job      func(ctx context.Context)
	}{interval: d, job: job})
	return nil
}

func (c *SimpleCron) Start(ctx context.Context) error {
	if c.running {
		return nil
	}
	c.running = true
	var ctx2 context.Context
	ctx2, c.cancel = context.WithCancel(ctx)
	for _, e := range c.entries {
		go func(ent struct {
			interval time.Duration
			job      func(ctx context.Context)
		}) {
			t := time.NewTicker(ent.interval)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					ent.job(ctx2)
				case <-ctx2.Done():
					return
				}
			}
		}(e)
	}
	return nil
}

func (c *SimpleCron) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.running = false
	return nil
}

// ScheduleTrigger node — not run inside orchestrator: it calls orchestrator.Run when triggered.
type ScheduleTrigger struct {
	id        string
	cronExpr  string
	scheduler flow.Scheduler
	orch      func(ctx context.Context, startNode string, input flow.Input)
	startNode string
}

func New(id string, cronExpr string, scheduler flow.Scheduler, orch func(ctx context.Context, startNode string, input flow.Input), startNode string) *ScheduleTrigger {
	return &ScheduleTrigger{id: id, cronExpr: cronExpr, scheduler: scheduler, orch: orch, startNode: startNode}
}

func (s *ScheduleTrigger) Start(ctx context.Context) error {
	return s.scheduler.Schedule(s.cronExpr, func(ctx context.Context) {
		// call orchestrator with empty payload
		s.orch(ctx, s.startNode, flow.Input{Payload: map[string]any{"triggered_by": s.id}})
	})
}
