package driplang

import (
	"github.com/micvbang/go-helpy/stringy"
)

// ContainsOperator returns true if the operator `op` is part of the expression
// `e` (or any of its subexpressions).
func ContainsOperator(e Expr, op Expr) bool {
	switch v := e.(type) {
	case EventName:
		return IsOperator(v, op)

	case Or:
		return IsOperator(v, op) || ContainsOperator(v.A, op) || ContainsOperator(v.B, op)

	case And:
		return IsOperator(v, op) || ContainsOperator(v.A, op) || ContainsOperator(v.B, op)

	case Then:
		return IsOperator(v, op) || ContainsOperator(v.A, op) || ContainsOperator(v.B, op)

	case After:
		return IsOperator(v, op) || ContainsOperator(v.A, op)

	case Not:
		return IsOperator(v, op) || ContainsOperator(v.A, op)

	default:
		return false
	}
}

// IsOperator returns true if the root expression of `e` is the same type as
// `op`.
func IsOperator(e Expr, op Expr) bool {
	switch op.(type) {
	case EventName:
		_, ok := e.(EventName)
		return ok

	case Or:
		_, ok := e.(Or)
		return ok

	case And:
		_, ok := e.(And)
		return ok

	case Then:
		_, ok := e.(Then)
		return ok

	case After:
		_, ok := e.(After)
		return ok

	case Not:
		_, ok := e.(Not)
		return ok

	default:
		return false
	}
}

// GetNames returns a list of all the unique names contained within Expr.
func GetNames(e Expr) []string {
	return stringy.Unique(getNames(e))
}

func getNames(e Expr) []string {
	switch v := e.(type) {

	case EventName:
		return []string{string(v)}

	case Not:
		return GetNames(v.A)

	case Or:
		return append(GetNames(v.A), GetNames(v.B)...)

	case And:
		return append(GetNames(v.A), GetNames(v.B)...)

	case Then:
		return append(GetNames(v.A), GetNames(v.B)...)

	case After:
		return GetNames(v.A)

	default:
		return []string{}
	}
}
