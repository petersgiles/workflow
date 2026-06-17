package lookup

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"local.com/internal/flow"
)

type LookupNode struct {
	id      string
	src     string
	field   string
	pattern string
	match   string
	mode    string
	trans   map[string]string
}

func New(id, src, field, pattern, match, mode string, trans map[string]string) *LookupNode {
	if src == "" {
		src = "devices"
	}
	if field == "" {
		field = "name"
	}
	if match == "" {
		match = "exact"
	}
	if mode == "" {
		mode = "first"
	}
	return &LookupNode{id: id, src: src, field: field, pattern: pattern, match: match, mode: mode, trans: trans}
}

func (l *LookupNode) ID() string { return l.id }

func (l *LookupNode) Enter(ctx context.Context, in flow.Input) (flow.State, error) {
	// prepare source value for Process to use from the incoming payload only.
	// Do not fallback to any orchestrator-injected state. If the required
	// source is missing, return an error to indicate a violated precondition.
	srcVal := getSourceValue(in.Payload, l.src)
	if srcVal == nil {
		return flow.State{}, fmt.Errorf("lookup: source '%s' not found in payload", l.src)
	}
	data := map[string]any{"src": srcVal}
	return flow.State{Data: data, Input: in}, nil
}

func getSourceValue(payload map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var cur any = payload
	for _, p := range parts {
		if m, ok := cur.(map[string]any); ok {
			cur = m[p]
		} else {
			return nil
		}
	}
	return cur
}

func matchValue(v any, field, pattern, matchType string) bool {
	// extract candidate string
	var s string
	switch vv := v.(type) {
	case map[string]any:
		if x, ok := vv[field]; ok {
			if ss, ok2 := x.(string); ok2 {
				s = ss
			}
		}
	case map[string]string:
		if ss, ok := vv[field]; ok {
			s = ss
		}
	case string:
		s = vv
	}
	if pattern == "" {
		return false
	}
	switch matchType {
	case "exact":
		return s == pattern
	case "ci_exact":
		return strings.EqualFold(s, pattern)
	case "substring":
		return strings.Contains(s, pattern)
	case "ci_substring":
		return strings.Contains(strings.ToLower(s), strings.ToLower(pattern))
	case "regex":
		ok, _ := regexp.MatchString(pattern, s)
		return ok
	default:
		return s == pattern
	}
}

func (l *LookupNode) Process(ctx context.Context, s flow.State) (flow.Result, error) {
	var srcVal any
	if v, ok := s.Data["src"]; ok {
		srcVal = v
	}
	var foundList []map[string]any

	switch arr := srcVal.(type) {
	case []any:
		for _, e := range arr {
			if matchValue(e, l.field, l.pattern, l.match) {
				// normalize to map[string]any
				switch m := e.(type) {
				case map[string]any:
					foundList = append(foundList, m)
				case map[string]string:
					mm := map[string]any{}
					for k, v := range m {
						mm[k] = v
					}
					foundList = append(foundList, mm)
				}
				if l.mode == "first" {
					break
				}
			}
		}
	case []map[string]any:
		for _, e := range arr {
			if matchValue(e, l.field, l.pattern, l.match) {
				foundList = append(foundList, e)
				if l.mode == "first" {
					break
				}
			}
		}
	case []map[string]string:
		for _, e := range arr {
			if matchValue(e, l.field, l.pattern, l.match) {
				mm := map[string]any{}
				for k, v := range e {
					mm[k] = v
				}
				foundList = append(foundList, mm)
				if l.mode == "first" {
					break
				}
			}
		}
	}

	var out any = nil
	if len(foundList) > 0 {
		if l.mode == "all" {
			out = foundList
		} else {
			out = foundList[0]
		}
	}

	// If nothing found, treat as failure (return error) so the orchestrator
	// follows the configured `fail` transition. On success, return data and
	// let orchestrator follow `success` transition if configured.
	if out == nil {
		return flow.Result{}, fmt.Errorf("lookup: not found pattern '%s' in source '%s'", l.pattern, l.src)
	}

	// success
	return flow.Result{Data: map[string]any{"found": out}, Signal: flow.ControlSignal{}}, nil
}

func (l *LookupNode) Exit(ctx context.Context, s flow.State) (flow.FinalResult, error) {
	return flow.FinalResult{Data: s.Data}, nil
}
