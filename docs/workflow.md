# Workflow

Below is concise documentation explaining the parts you have, how they map to a text-file (n8n-style) flow format, and a minimal plan for the additional nodes you listed.

``` bash
❯ tree .
.
├── docs
│   └── workflow.md
├── go.mod
├── go.sum
├── internal
│   ├── adapters
│   │   └── nodes
│   │       └── noop
│   │           └── noop.go
│   └── flow
│       ├── factory.go
│       ├── loader.go
│       ├── node.go
│       ├── orchestrator.go
│       ├── ports.go
│       ├── spec.go
│       └── types.go
├── local.com
└── main.go
```


1) What each part does
- ports.go (domain-owned contracts)
  - Declares Node and Events interfaces and any small client ports you add later. Domain code depends only on these interfaces.
- types.go (DTOs)
  - Defines Input, State, Result, ControlSignal, FinalResult — the stable data exchanged between orchestrator and nodes.
- node.go
  - Small domain file (errors, helpers). Nodes implement the Node interface from ports.go.
- orchestrator.go
  - The executor: takes an ordered list of node IDs, resolves Node implementations via a Registry, manages lifecycle calls (Enter → Process → Exit), handles ControlSignal (Next, Retry, Abort, Delay), and emits Events.
- adapters/nodes/noop/noop.go
  - Example adapter implementing Node. Adapters contain side effects and depend only on client ports injected via constructors (noop has none).

2) memRegistry usage
- memRegistry is a simple in-memory registry mapping node IDs to Node instances. It’s suitable for:
  - quick wiring in main for tests and small runs
  - loading a flow instance in memory (nodes already constructed)
- For text-driven flows, treat the registry as the runtime resolver: you will parse a flow file into a list/graph of node instances (with config), construct adapter instances (using composition wiring), and populate the registry with those instances keyed by node IDs.

3) Text-file flow format (suggested minimal YAML)
- Represent nodes as a list and edges (for graph flows) with optional configuration per node:
  - Example (linear):
    id: flow-1
    nodes:
      - id: start
        type: http_trigger
        config:
          method: POST
          path: /start
      - id: run-script
        type: run_script
        config:
          cmd: ./testsuite.sh
      - id: notify
        type: webhook
        config:
          url: https://hooks.example.com/build
- Example (graph with explicit transitions):
    nodes:
      - id: check
        type: branch
        config:
          expr: "payload.status == 'ok'"
        transitions:
          true: run
          false: fail
      - id: run
        type: run_script
      - id: fail
        type: console

4) How parsing -> runtime mapping works (high level)
- Composition root holds:
  - Node factory map: map[type]string -> constructor func(config map[string]any) (flow.Node, error)
  - System clients (HTTPClient, Scheduler, ScriptRunner, Store) that adapters can receive
- Flow loader:
  - Read YAML/JSON file into a FlowSpec struct (list of node specs and transitions).
  - For each NodeSpec, call factory for its type with its config + injected clients → Node instance.
  - Populate runtime registry with nodeID -> Node instance.
  - Build execution order or graph (list of node IDs or adjacency map) from transitions or default linear ordering.
- Execution:
  - For simple linear runs, pass ordered nodeIDs to orchestrator.Run.
  - For graph runs, extend orchestrator to evaluate transitions (Result.Signal.Next or transition mapping in spec) and follow edges.

5) Extending orchestrator for graph/triggered flows
- Keep the same Node lifecycle but change orchestrator.Run to:
  - Accept a start node ID (or trigger event) and an adjacency map.
  - Execute nodes based on Result.Signal.Next or the spec’s transitions mapping.
  - Support concurrency by launching independent branches as goroutines and aggregating FinalResult (use sync.WaitGroup, contexts).
  - Provide a RunUntilIdle or RunWithTimeout mode for long-running trigger-based flows.

6) Suggested node implementations (brief responsibilities & ports they need)
- Schedule Trigger
  - Purpose: start a flow on a schedule (cron). Runs outside orchestrator as a trigger service that calls orchestrator with start node.
  - Needs: Scheduler port (cron-like), Events.
- HTTP Trigger
  - Purpose: expose HTTP endpoint to start a flow with incoming payload.
  - Needs: HTTP server handler adapter that calls orchestrator.Run for the flow, Events.
- Run Script
  - Purpose: execute shell commands or scripts and return outputs.
  - Needs: ScriptRunner port (Exec(ctx, cmd, env) -> stdout, stderr, exitCode, error).
- Branch (if / switch)
  - Purpose: evaluate an expression against current payload and route to Next node accordingly.
  - Design: either a special Node that returns Result.Signal.Next, or orchestrator uses node spec transitions mapping.
  - Needs: optionally an expression evaluator port (for safety/sandboxing).
- Loop
  - Purpose: iterate over array (CSV/JSON/YAML) and run a child subflow per item.
  - Needs: ability to spawn sub-orchestrators (or orchestrator exposes a RunSubflow(ctx, nodeIDs, input)).
- Console
  - Purpose: write to stdout/stderr, used for debugging.
  - Needs: Logger/Console port.
- Webhook
  - Purpose: send HTTP request to external service.
  - Needs: HTTPClient port.

7) Important design details and constraints
- Keep Node implementations deterministic and side-effects explicit via client ports; orchestrator manages retries/timeouts.
- Control flow: prefer Result.Signal (Next, Retry, Abort) over returning errors to encode non-fatal control decisions.
- Config-driven wiring: node config from file determines which adapter implementation to construct and with which parameters.
- Security: sandbox script execution; validate/whitelist expression language for Branch nodes.
- Testing: provide mock clients and test flows by loading text files and running orchestrator with deterministic events.

8) Minimal next steps you can implement now
- Create FlowSpec structs and a loader to parse YAML into NodeSpec (id, type, config, transitions).
- Implement a factory registry for node types (map[string]func(cfg map[string]any, clients ...any) (flow.Node, error)).
- Extend memRegistry to accept node configs and keep node instances.
- Implement one new adapter: run_script using os/exec behind a ScriptRunner port; add a Console node.
- Add an HTTP trigger adapter that calls orchestrator.Run with parsed payload.
