package driplang

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Marshal marshals an expression to a byte format that can be unmarshalled
// (using Unmarshal) to the same expression.
func Marshal(e Expr) ([]byte, error) {
	return json.Marshal(&e)
}

func (a And) MarshalJSON() ([]byte, error) {
	return marshalABOperator("and", a.A, a.B)
}

func (o Or) MarshalJSON() ([]byte, error) {
	return marshalABOperator("or", o.A, o.B)
}

func (t Then) MarshalJSON() ([]byte, error) {
	return marshalABOperator("then", t.A, t.B)
}

func (n Not) MarshalJSON() ([]byte, error) {
	opa, err := json.Marshal(n.A)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf(`{"operator": "not", "a": %v}`, string(opa))), nil
}

func (e EventName) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"operator": "event_name", "a": "%v"}`, string(e))), nil
}

func (a After) MarshalJSON() ([]byte, error) {
	opa, err := json.Marshal(a.A)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf(`{"operator": "after", "a": %v, "d": "%v"}`, string(opa), a.D)), nil
}

func marshalABOperator(name string, a, b Expr) ([]byte, error) {
	opa, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	opb, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf(`{"operator": "%s", "a": %v, "b": %v}`, name, string(opa), string(opb))), nil
}

// Unmarshal unmarshals an Expr.
func Unmarshal(bs []byte) (Expr, error) {
	m := map[string]interface{}{}
	err := json.Unmarshal(bs, &m)
	if err != nil {
		return nil, err
	}

	return unmarshal(m)
}

// ErrInvalidExpression is returned when attempting to unmarshal something that
// isn't a driplang.Expr.
var ErrInvalidExpression = errors.New("invalid expression")

func unmarshal(m map[string]interface{}) (Expr, error) {
	name, ok := m["operator"].(string)
	if !ok {
		return nil, ErrInvalidExpression
	}

	switch name {
	case "event_name":
		return EventName(m["a"].(string)), nil

	case "not":
		a, err := unmarshal(m["a"].(map[string]interface{}))
		return Not{A: a}, err

	case "and":
		a, err := unmarshal(m["a"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		b, err := unmarshal(m["b"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		return And{A: a, B: b}, nil

	case "or":
		a, err := unmarshal(m["a"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		b, err := unmarshal(m["b"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return Or{A: a, B: b}, nil

	case "then":
		a, err := unmarshal(m["a"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		b, err := unmarshal(m["b"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return Then{A: a, B: b}, nil

	case "after":
		a, err := unmarshal(m["a"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		v, isStr := m["d"].(string)
		d, err := strconv.ParseInt(v, 10, 64)
		if !isStr || err != nil {
			return nil, err
		}
		return After{A: a, D: Duration(d)}, nil

	default:
		return nil, fmt.Errorf("unhandled Expr %v", m)
	}
}
