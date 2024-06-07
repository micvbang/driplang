package driplang_test

import (
	"testing"
	"time"

	"github.com/micvbang/driplang"
	"github.com/micvbang/go-helpy/stringy"
	"github.com/stretchr/testify/require"
)

// TestNames verifies that Names returns all EventNames used in an
// expression.
func TestNames(t *testing.T) {
	const (
		name1 = "1"
		name2 = "2"
		name3 = "3"
		name4 = "4"
	)
	expr := driplang.Then{
		A: driplang.EventName(name1),
		B: driplang.And{
			A: driplang.After{
				A: driplang.EventName(name2),
				D: driplang.Duration(0),
			},
			B: driplang.Or{
				A: driplang.EventName(name4),
				B: driplang.Not{
					A: driplang.EventName(name3),
				},
			},
		},
	}

	names := stringy.MakeSet(driplang.Names(expr)...)

	require.Equal(t, 4, len(names))
	require.True(t, names.Contains(name1))
	require.True(t, names.Contains(name2))
	require.True(t, names.Contains(name3))
	require.True(t, names.Contains(name4))
	require.False(t, names.Contains("not in set"))
}

func TestIsOperatorSimple(t *testing.T) {
	tests := map[string]struct {
		expected bool
		expr     driplang.Expr
		op       driplang.Expr
	}{
		"eventname": {
			expected: true,
			expr:     driplang.EventName("a"),
			op:       driplang.EventName(""),
		},
		"or": {
			expected: true,
			expr: driplang.Or{
				A: driplang.EventName("a"),
				B: driplang.EventName("b"),
			},
			op: driplang.Or{},
		},
		"and": {
			expected: true,
			expr: driplang.And{
				A: driplang.EventName("a"),
				B: driplang.EventName("b"),
			},
			op: driplang.And{},
		},
		"then": {
			expected: true,
			expr: driplang.Then{
				A: driplang.EventName("a"),
				B: driplang.EventName("b"),
			},
			op: driplang.Then{},
		},
		"after": {
			expected: true,
			expr: driplang.After{
				A: driplang.EventName("a"),
				D: driplang.Duration(5 * time.Minute),
			},
			op: driplang.After{},
		},
		"not": {
			expected: true,
			expr: driplang.Not{
				A: driplang.EventName("a"),
			},
			op: driplang.Not{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.IsOperator(test.expr, test.op))
		})
	}
}

func TestIsOperatorSimpleFalse(t *testing.T) {
	tests := map[string]struct {
		expected bool
		expr     driplang.Expr
		op       driplang.Expr
	}{
		"eventname": {
			expected: false,
			expr:     driplang.EventName("a"),
			op:       driplang.And{},
		},
		"or": {
			expected: false,
			expr: driplang.Or{
				A: driplang.EventName("a"),
				B: driplang.EventName("b"),
			},
			op: driplang.Then{},
		},
		"and": {
			expected: false,
			expr: driplang.And{
				A: driplang.EventName("a"),
				B: driplang.EventName("b"),
			},
			op: driplang.Or{},
		},
		"then": {
			expected: false,
			expr: driplang.Then{
				A: driplang.EventName("a"),
				B: driplang.EventName("b"),
			},
			op: driplang.After{},
		},
		"after": {
			expected: false,
			expr: driplang.After{
				A: driplang.EventName("a"),
				D: driplang.Duration(5 * time.Minute),
			},
			op: driplang.Not{},
		},
		"not": {
			expected: false,
			expr: driplang.Not{
				A: driplang.EventName("a"),
			},
			op: driplang.EventName(""),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.IsOperator(test.expr, test.op))
		})
	}
}

func TestContainsOperator(t *testing.T) {
	tests := map[string]struct {
		expected bool
		expr     driplang.Expr
		op       driplang.Expr
	}{
		"eventname": {
			expected: true,
			expr: driplang.And{
				A: driplang.EventName("a"),
				B: driplang.EventName("b"),
			},
			op: driplang.EventName("op"),
		},
		"or": {
			expected: true,
			expr: driplang.Then{
				A: driplang.EventName("b"),
				B: driplang.Or{
					A: driplang.EventName("a"),
					B: driplang.EventName("b"),
				},
			},
			op: driplang.Or{},
		},
		"and": {
			expected: true,
			expr: driplang.After{
				A: driplang.And{
					A: driplang.EventName("a"),
					B: driplang.EventName("b"),
				},
				D: driplang.Duration(42 * time.Second),
			},
			op: driplang.And{},
		},
		"then": {
			expected: true,
			expr: driplang.Not{
				A: driplang.Then{
					A: driplang.EventName("a"),
					B: driplang.EventName("b"),
				},
			},
			op: driplang.Then{},
		},
		"after": {
			expected: true,
			expr: driplang.Or{
				A: driplang.After{
					A: driplang.EventName("a"),
					D: driplang.Duration(5 * time.Minute),
				},
				B: driplang.EventName("b"),
			},
			op: driplang.After{},
		},
		"not": {
			expected: true,
			expr: driplang.Or{
				A: driplang.EventName("a"),
				B: driplang.Not{
					A: driplang.EventName("a"),
				},
			},
			op: driplang.Not{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.ContainsOperator(test.expr, test.op))
		})
	}
}

func TestContainsOperatorFalse(t *testing.T) {
	tests := map[string]struct {
		expected bool
		expr     driplang.Expr
		op       driplang.Expr
	}{
		"or": {
			expected: false,
			expr: driplang.Then{
				A: driplang.EventName("b"),
				B: driplang.And{
					A: driplang.EventName("a"),
					B: driplang.EventName("b"),
				},
			},
			op: driplang.Or{},
		},
		"and": {
			expected: false,
			expr: driplang.After{
				A: driplang.Or{
					A: driplang.EventName("a"),
					B: driplang.EventName("b"),
				},
				D: driplang.Duration(42 * time.Second),
			},
			op: driplang.And{},
		},
		"then": {
			expected: false,
			expr: driplang.Not{
				A: driplang.And{
					A: driplang.EventName("a"),
					B: driplang.EventName("b"),
				},
			},
			op: driplang.Then{},
		},
		"after": {
			expected: false,
			expr: driplang.Or{
				A: driplang.And{
					A: driplang.EventName("a"),
					B: driplang.EventName("b"),
				},
				B: driplang.EventName("b"),
			},
			op: driplang.After{},
		},
		"not": {
			expected: false,
			expr: driplang.Or{
				A: driplang.EventName("a"),
				B: driplang.EventName("b"),
			},
			op: driplang.Not{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.ContainsOperator(test.expr, test.op))
		})
	}
}
