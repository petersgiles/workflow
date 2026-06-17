package main

import (
	"context"
	"flag"
	"log"
	"time"

	expr "local.com/internal/adapters/expr"
	bluetooth "local.com/internal/adapters/nodes/bluetooth"
	"local.com/internal/adapters/nodes/branch"
	"local.com/internal/adapters/nodes/console"
	"local.com/internal/adapters/nodes/httpclient"
	lookup "local.com/internal/adapters/nodes/lookup"
	"local.com/internal/adapters/nodes/loop"
	"local.com/internal/adapters/nodes/noop"
	runscript "local.com/internal/adapters/nodes/runscript"
	scriptpkg "local.com/internal/adapters/nodes/script"
	"local.com/internal/adapters/nodes/webhook"
	"local.com/internal/flow"
)

type Deps struct {
	Events       flow.Events
	ScriptRunner flow.ScriptRunner
	ExprEval     flow.ExprEvaluator
	HTTPClient   flow.HTTPClient
}

// scriptRunnerAdapter adapts the existing OSRunner to the flow.ScriptRunner interface.
type scriptRunnerAdapter struct{ impl *scriptpkg.OSRunner }

func (s scriptRunnerAdapter) Exec(ctx context.Context, cmd string, args []string, env map[string]string) (string, string, int, error) {
	return s.impl.Exec(ctx, cmd, args, env, 60*time.Second)
}

// expression evaluator lives in package flow (see internal/flow/simple_expr_eval.go)

func main() {
	flowPath := flag.String("flow", "flows/build.yaml", "path to flow yaml")
	flag.Parse()

	ctx := context.Background()
	deps := Deps{
		Events:       flow.LogEvents{},
		ScriptRunner: scriptRunnerAdapter{impl: scriptpkg.NewOSRunner()},
		ExprEval:     expr.SimpleEval{},
		HTTPClient:   httpclient.NewNetClient(),
	}

	fr := flow.NewSimpleFactoryRegistry()
	fr.Register("noop", func(spec flow.NodeSpec, d any) (flow.Node, error) {
		return noop.New(spec.ID), nil
	})

	fr.Register("console", func(spec flow.NodeSpec, d any) (flow.Node, error) {
		return console.New(spec.ID), nil
	})

	fr.Register("run_script", func(spec flow.NodeSpec, d any) (flow.Node, error) {
		deps := d.(Deps)
		cfg := spec.Config
		cmd := ""
		if v, ok := cfg["cmd"]; ok {
			if s, ok := v.(string); ok {
				cmd = s
			}
		}
		var args []string
		if v, ok := cfg["args"]; ok {
			if arr, ok := v.([]any); ok {
				for _, e := range arr {
					if s, ok := e.(string); ok {
						args = append(args, s)
					}
				}
			}
		}
		env := map[string]string{}
		if v, ok := cfg["env"]; ok {
			if m, ok := v.(map[string]any); ok {
				for kk, vv := range m {
					if s, ok := vv.(string); ok {
						env[kk] = s
					}
				}
			}
		}
		return runscript.New(spec.ID, cmd, args, env, deps.ScriptRunner), nil
	})

	fr.Register("branch", func(spec flow.NodeSpec, d any) (flow.Node, error) {
		deps := d.(Deps)
		expr := ""
		if v, ok := spec.Config["expr"]; ok {
			if s, ok := v.(string); ok {
				expr = s
			}
		}
		// optional requires list
		var requires []string
		if v, ok := spec.Config["requires"]; ok {
			if arr, ok2 := v.([]any); ok2 {
				for _, e := range arr {
					if sreq, ok3 := e.(string); ok3 {
						requires = append(requires, sreq)
					}
				}
			}
		}
		return branch.New(spec.ID, expr, deps.ExprEval, spec.Transitions, requires), nil
	})

	fr.Register("loop", func(spec flow.NodeSpec, d any) (flow.Node, error) {
		// items_key: string, subflow: []string
		itemsKey := "items"
		if v, ok := spec.Config["items_key"]; ok {
			if s, ok := v.(string); ok {
				itemsKey = s
			}
		}
		var subflowIds []string
		if v, ok := spec.Config["subflow"]; ok {
			if arr, ok := v.([]any); ok {
				for _, e := range arr {
					if s, ok := e.(string); ok {
						subflowIds = append(subflowIds, s)
					}
				}
			}
		}
		orchStub := func(ctx context.Context, nodeIDs []string, input flow.Input) (flow.FinalResult, error) {
			return flow.FinalResult{}, nil
		}
		return loop.New(spec.ID, itemsKey, subflowIds, orchStub), nil
	})

	fr.Register("webhook", func(spec flow.NodeSpec, d any) (flow.Node, error) {
		deps := d.(Deps)
		url := ""
		if v, ok := spec.Config["url"]; ok {
			if s, ok := v.(string); ok {
				url = s
			}
		}
		return webhook.New(spec.ID, url, deps.HTTPClient), nil
	})

	fr.Register("lookup", func(spec flow.NodeSpec, d any) (flow.Node, error) {
		src := "devices"
		if v, ok := spec.Config["source"]; ok {
			if s, ok := v.(string); ok {
				src = s
			}
		}
		field := "name"
		if v, ok := spec.Config["field"]; ok {
			if s, ok := v.(string); ok {
				field = s
			}
		}
		pattern := ""
		if v, ok := spec.Config["pattern"]; ok {
			if s, ok := v.(string); ok {
				pattern = s
			}
		}
		match := "exact"
		if v, ok := spec.Config["match"]; ok {
			if s, ok := v.(string); ok {
				match = s
			}
		}
		mode := "first"
		if v, ok := spec.Config["mode"]; ok {
			if s, ok := v.(string); ok {
				mode = s
			}
		}
		return lookup.New(spec.ID, src, field, pattern, match, mode, spec.Transitions), nil
	})

	fr.Register("bluetooth", func(spec flow.NodeSpec, d any) (flow.Node, error) {
		timeout := 5 * time.Second
		if v, ok := spec.Config["timeout_seconds"]; ok {
			if n, ok := v.(int); ok {
				timeout = time.Duration(n) * time.Second
			} else if f, ok := v.(float64); ok {
				timeout = time.Duration(int(f)) * time.Second
			}
		}
		// use a concrete adapter scanner
		scanner := bluetooth.NewAdapterScanner()
		return bluetooth.New(spec.ID, scanner, timeout), nil
	})

	spec, err := flow.LoadFlowSpec(*flowPath)
	if err != nil {
		log.Fatal(err)
	}

	flow.WriteMermaid(spec, "flow.mmd")

	rt, order, err := flow.BuildRuntime(spec, fr, deps)
	if err != nil {
		log.Fatal(err)
	}

	orch := flow.NewOrchestrator(rt, deps.Events)
	_, err = orch.Run(ctx, order, flow.Input{Payload: map[string]any{"from": "trigger"}})
	if err != nil {
		log.Fatal(err)
	}
}
