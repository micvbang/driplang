package driplang

import (
	"fmt"
	"time"
)

/*
expr 		::= expr AND expr | expr OR expr | NOT expr | expr THEN expr | expr AFTER duration
event_name 	::= [string]
duration    ::= [int]

*/

type Expr interface {
	Representation() string
}

type Duration time.Duration

type EventName string

func (e EventName) Representation() string {
	return fmt.Sprintf("\"%v\"", e)
}

type And struct {
	A Expr `json:"a"`
	B Expr `json:"b"`
}

func (a And) Representation() string {
	return fmt.Sprintf("(%s AND %s)", a.A.Representation(), a.B.Representation())
}

type Or struct {
	A Expr `json:"a"`
	B Expr `json:"b"`
}

func (o Or) Representation() string {
	return fmt.Sprintf("(%s OR %s)", o.A.Representation(), o.B.Representation())
}

type Not struct {
	A Expr `json:"a"`
}

func (n Not) Representation() string {
	return fmt.Sprintf("(NOT %s)", n.A.Representation())
}

type Then struct {
	A Expr `json:"a"`
	B Expr `json:"b"`
}

func (t Then) Representation() string {
	return fmt.Sprintf("(%s THEN %s)", t.A.Representation(), t.B.Representation())
}

type After struct {
	A Expr     `json:"a"`
	D Duration `json:"t"`
}

func (a After) Representation() string {
	return fmt.Sprintf("(%s AFTER %s)", a.A.Representation(), time.Duration(a.D))
}
