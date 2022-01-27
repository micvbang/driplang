package driplang_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	driplang "gitlab.com/micvbang/event-dripper/internal/eventtriggering/driplang"
)

// TestMarshalEqualUnmarshal verifies that the result of Marshalling and
// Unmarshalling an expression returns the same expression.
func TestMarshalEqualUnmarshal(t *testing.T) {
	tests := map[string]struct {
		expr driplang.Expr
	}{
		"and": {
			expr: driplang.And{
				A: driplang.EventName("a"),
				B: driplang.EventName("b"),
			},
		},
		"or": {
			expr: driplang.Or{
				A: driplang.EventName("a"),
				B: driplang.EventName("b"),
			},
		},
		"then": {
			expr: driplang.Then{
				A: driplang.EventName("a"),
				B: driplang.EventName("b"),
			},
		},
		"not": {
			expr: driplang.Not{
				A: driplang.EventName("a"),
			},
		},
		"event_name": {
			expr: driplang.EventName("a"),
		},
		"after": {
			expr: driplang.After{
				A: driplang.EventName("a"),
				D: driplang.Duration(42133742),
			},
		},
		"deeply nested": {
			expr: driplang.Then{
				A: driplang.And{
					A: driplang.After{
						A: driplang.Not{
							A: driplang.EventName("1"),
						},
						D: driplang.Duration(10 * time.Hour),
					},
					B: driplang.Or{
						A: driplang.EventName("2"),
						B: driplang.After{
							A: driplang.Not{
								A: driplang.EventName("3"),
							},
							D: driplang.Duration(1 * time.Millisecond),
						},
					},
				},
				B: driplang.And{
					A: driplang.And{
						A: driplang.Not{
							A: driplang.EventName("4"),
						},
						B: driplang.Or{
							A: driplang.EventName("5"),
							B: driplang.Not{
								A: driplang.EventName("6"),
							},
						},
					},
					B: driplang.Or{
						A: driplang.EventName("7"),
						B: driplang.Not{
							A: driplang.EventName("8"),
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bs, err := driplang.Marshal(test.expr)
			require.NoError(t, err, "failed to marshal")

			gotExpr, err := driplang.Unmarshal(bs)
			require.NoError(t, err, "failed to unmarshal")

			require.Equal(t, test.expr, gotExpr)
		})
	}
}

func TestUnmarshalInvalidExpression(t *testing.T) {
	tests := map[string]struct {
		bs  []byte
		err error
	}{
		"invalid 1": {
			bs:  jsonMarshal(t, make(map[int]byte)),
			err: driplang.ErrInvalidExpression,
		},
		"invalid 2": {
			bs:  jsonMarshal(t, map[string]interface{}{"no_operation": 1}),
			err: driplang.ErrInvalidExpression,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := driplang.Unmarshal(test.bs)
			require.Equal(t, test.err, err)
		})
	}
}

func jsonMarshal(t *testing.T, v interface{}) []byte {
	mv, err := json.Marshal(v)

	require.NoError(t, err)
	return mv
}
