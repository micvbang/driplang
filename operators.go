package driplang

import (
	"fmt"
	"time"

	"github.com/micvbang/go-helpy/stringy"
)

/*
expr 		::= expr AND expr | expr OR expr | NOT expr | expr THEN expr | expr AFTER duration
event_name 	::= [string]
duration    ::= [int]

*/

type Duration time.Duration

type Expr interface {
	Expression() string
}

type EventName string

func (e EventName) Expression() string {
	return fmt.Sprintf("\"%v\"", e)
}

type And struct {
	A Expr `json:"a"`
	B Expr `json:"b"`
}

func (a And) Expression() string {
	return fmt.Sprintf("(%s AND %s)", a.A.Expression(), a.B.Expression())
}

type Or struct {
	A Expr `json:"a"`
	B Expr `json:"b"`
}

func (o Or) Expression() string {
	return fmt.Sprintf("(%s OR %s)", o.A.Expression(), o.B.Expression())
}

type Not struct {
	A Expr `json:"a"`
}

func (n Not) Expression() string {
	return fmt.Sprintf("(NOT %s)", n.A.Expression())
}

type Then struct {
	A Expr `json:"a"`
	B Expr `json:"b"`
}

func (t Then) Expression() string {
	return fmt.Sprintf("(%s THEN %s)", t.A.Expression(), t.B.Expression())
}

type After struct {
	A Expr     `json:"a"`
	D Duration `json:"t"`
}

func (a After) Expression() string {
	return fmt.Sprintf("(%s AFTER %s)", a.A.Expression(), time.Duration(a.D))
}

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

// Names returns a list of all the unique names contained within Expr.
func Names(e Expr) []string {
	return stringy.Unique(getNames(e))
}

func getNames(e Expr) []string {
	switch v := e.(type) {

	case EventName:
		return []string{string(v)}

	case Not:
		return Names(v.A)

	case Or:
		return append(Names(v.A), Names(v.B)...)

	case And:
		return append(Names(v.A), Names(v.B)...)

	case Then:
		return append(Names(v.A), Names(v.B)...)

	case After:
		return Names(v.A)

	default:
		return []string{}
	}
}
