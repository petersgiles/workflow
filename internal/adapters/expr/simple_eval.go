package expr

import (
	"context"
	"strconv"
	"strings"
)

// SimpleEval is a minimal, unsafe expression evaluator intended for
// development and tests. Replace with a proper parser/evaluator for
// production use.
type SimpleEval struct{}

func (SimpleEval) EvalBool(ctx context.Context, expr string, payload map[string]any) (bool, error) {
	e := strings.TrimSpace(expr)
	if e == "true" {
		return true, nil
	}
	if e == "false" {
		return false, nil
	}
	// operators
	ops := []string{"==", "!=", ">=", "<=", ">", "<"}
	for _, op := range ops {
		if idx := strings.Index(e, op); idx != -1 {
			left := strings.TrimSpace(e[:idx])
			right := strings.TrimSpace(e[idx+len(op):])
			lv, ltyp := evalOperand(left, payload)
			rv, rtyp := evalOperand(right, payload)
			// numeric comparison if both numeric
			if ltyp == "number" && rtyp == "number" {
				lf := lv.(float64)
				rf := rv.(float64)
				switch op {
				case "==":
					return lf == rf, nil
				case "!=":
					return lf != rf, nil
				case ">":
					return lf > rf, nil
				case "<":
					return lf < rf, nil
				case ">=":
					return lf >= rf, nil
				case "<=":
					return lf <= rf, nil
				}
			}
			// boolean compare
			if ltyp == "bool" && rtyp == "bool" {
				lb := lv.(bool)
				rb := rv.(bool)
				switch op {
				case "==":
					return lb == rb, nil
				case "!=":
					return lb != rb, nil
				}
			}
			// string compare
			ls := toString(lv)
			rs := toString(rv)
			switch op {
			case "==":
				return ls == rs, nil
			case "!=":
				return ls != rs, nil
			}
			return false, nil
		}
	}
	return false, nil
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func evalOperand(tok string, payload map[string]any) (any, string) {
	tok = strings.TrimSpace(tok)
	if tok == "true" {
		return true, "bool"
	}
	if tok == "false" {
		return false, "bool"
	}
	// quoted string
	if len(tok) >= 2 && ((tok[0] == '"' && tok[len(tok)-1] == '"') || (tok[0] == '\'' && tok[len(tok)-1] == '\'')) {
		return tok[1 : len(tok)-1], "string"
	}
	// payload lookup e.g., payload.status
	if strings.HasPrefix(tok, "payload.") && payload != nil {
		path := strings.TrimPrefix(tok, "payload.")
		parts := strings.Split(path, ".")
		var cur any = payload
		for _, p := range parts {
			if m, ok := cur.(map[string]any); ok {
				cur = m[p]
			} else {
				cur = nil
				break
			}
		}
		if cur != nil {
			switch v := cur.(type) {
			case string:
				return v, "string"
			case bool:
				return v, "bool"
			case int:
				return float64(v), "number"
			case float64:
				return v, "number"
			default:
				return v, "string"
			}
		}
	}
	// numeric?
	if f, err := strconv.ParseFloat(tok, 64); err == nil {
		return f, "number"
	}
	// bareword -> string
	return tok, "string"
}
