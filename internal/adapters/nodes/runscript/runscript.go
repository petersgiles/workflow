package runscript

import (
	"context"

	"local.com/internal/flow"
)

type RunScriptNode struct {
	id     string
	cmd    string
	args   []string
	env    map[string]string
	runner flow.ScriptRunner
}

func New(id string, cmd string, args []string, env map[string]string, r flow.ScriptRunner) *RunScriptNode {
	return &RunScriptNode{id: id, cmd: cmd, args: args, env: env, runner: r}
}

func (n *RunScriptNode) ID() string { return n.id }

func (n *RunScriptNode) Enter(ctx context.Context, in flow.Input) (flow.State, error) {
	return flow.State{Data: map[string]any{}, Input: in}, nil
}

func (n *RunScriptNode) Process(ctx context.Context, s flow.State) (flow.Result, error) {
	stdout, stderr, exit, err := n.runner.Exec(ctx, n.cmd, n.args, n.env)
	data := map[string]any{"stdout": stdout, "stderr": stderr, "exit": exit}
	sig := flow.ControlSignal{}
	if err != nil || exit != 0 {
		sig.Abort = true
	}
	return flow.Result{Data: data, Signal: sig}, err
}

func (n *RunScriptNode) Exit(ctx context.Context, s flow.State) (flow.FinalResult, error) {
	return flow.FinalResult{Data: s.Data}, nil
}
