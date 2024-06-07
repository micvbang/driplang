package driplang_test

import (
	"testing"
	"time"

	"github.com/micvbang/driplang"
	"github.com/micvbang/go-helpy/timey"
	"github.com/stretchr/testify/require"
)

/*
TestEvaluateNeverXThenYFollowedByX verifies that the following logic works as
expected: "there never was 2, then there is 1 followed by 2"

My intuition incorrectly told me that the above logic could be expressed by the
expression `(NOT 2) THEN (1 THEN 2)`, but this is wrong; a counter example is as
follows:

For events [3, 2, 1, 2], `NOT 2` is satisfied by the sequence [3], while `1 THEN
2` is satisfied by the sequence `[2, 1, 2]`.

In order to encode the logic, we have to ensure that the second term of the
`THEN` expression is not satisfied if there is a `2` in the sequence.

The way to correctly express the logic is: `(NOT 2) THEN ((NOT 2 AND 1) THEN
2)`.  For events [3, 2, 1, 2] `NOT 2` will again be satisfied by [] or [3], but
`(NOT 2) AND 1` will not be satisfied by [3, 2] nor [3]. This means that the
whole expression can't be satisfied by the sequence which is exactly what we
want.

		THEN
		/	\
	NOT 2	THEN
			/  \
		  AND	2
		 /   \
	  NOT 2   1
*/
func TestEvaluateNeverXThenYFollowedByX(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
		eventName3 = "event3"
	)

	expr := driplang.Then{
		A: driplang.Not{A: driplang.EventName(eventName2)},
		B: driplang.Then{
			A: driplang.And{
				A: driplang.Not{A: driplang.EventName(eventName2)},
				B: driplang.EventName(eventName1),
			},
			B: driplang.EventName(eventName2),
		},
	}

	tests := map[string]struct {
		expected bool
		events   []driplang.Event
	}{
		"Existing `2` event not satisfied by outer `THEN.B` expression": {
			expected: false,
			events: makeEvents(
				eventName3,
				eventName2,
				eventName1,
				eventName2,
			),
		},
		"`NOT` trivially satisfied by non-existing event": {
			expected: true,
			events: makeEvents(
				eventName1,
				eventName2,
			),
		},
		"`NOT` satisfied by non-2 event": {
			expected: true,
			events: makeEvents(
				eventName3,
				eventName1,
				eventName2,
			),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.Evaluate(expr, test.events))
		})
	}
}

// TestEvaluateSimpleAnd verifies that simple And expressions are only satisfied
// when both sides are satisfied.
func TestEvaluateSimpleAnd(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
	)

	events := []driplang.Event{
		{Name: eventName2},
		{Name: eventName1},
	}

	tests := map[string]struct {
		expected bool
		expr     driplang.Expr
	}{
		"true": {
			expected: true,
			expr: driplang.And{
				A: driplang.EventName(eventName1),
				B: driplang.EventName(eventName2),
			},
		},
		"a false": {
			expected: false,
			expr: driplang.And{
				A: driplang.EventName("does not exist"),
				B: driplang.EventName(eventName1),
			},
		},
		"b false": {
			expected: false,
			expr: driplang.And{
				A: driplang.EventName(eventName1),
				B: driplang.EventName("does not exist"),
			},
		},
		"a+b false": {
			expected: false,
			expr: driplang.And{
				A: driplang.EventName("does not exist"),
				B: driplang.EventName("does not exist"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.Evaluate(test.expr, events))
		})
	}
}

// TestEvaluateSimpleOr verifies that simple Or expressions are satisfied when
// at least one of the sides are satisfied.
func TestEvaluateSimpleOr(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
	)

	events := []driplang.Event{
		{Name: eventName1},
		{Name: eventName2},
	}

	tests := map[string]struct {
		expected bool
		expr     driplang.Expr
	}{
		"a true": {
			expected: true,
			expr: driplang.Or{
				A: driplang.EventName(eventName1),
				B: driplang.EventName("does not exist"),
			},
		},
		"b true": {
			expected: true,
			expr: driplang.Or{
				A: driplang.EventName("does not exist"),
				B: driplang.EventName(eventName1),
			},
		},
		"a+b true": {
			expected: true,
			expr: driplang.Or{
				A: driplang.EventName(eventName1),
				B: driplang.EventName(eventName2),
			},
		},
		"both false": {
			expected: false,
			expr: driplang.Or{
				A: driplang.EventName("does not exist"),
				B: driplang.EventName("also does not exist"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.Evaluate(test.expr, events))
		})
	}
}

// TestEvaluateSimpleNot verifies that simple Not expressions are satisfied only
// when the underlying value isn't.
func TestEvaluateSimpleNot(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
	)

	events := []driplang.Event{
		{Name: eventName1},
		{Name: eventName1},
		{Name: eventName2},
		{Name: eventName1},
	}

	tests := map[string]struct {
		expected bool
		expr     driplang.Expr
	}{
		"true": {
			expected: true,
			expr: driplang.Not{
				A: driplang.EventName("does not exist"),
			},
		},
		"false": {
			expected: false,
			expr: driplang.Not{
				A: driplang.EventName(eventName1),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.Evaluate(test.expr, events))
		})
	}
}

// TestEvaluateSimpleAfter that After expressions only are satisfied when what
// they follow is within the correct time frame, e.g. "there was 1 and then at
// least 10 hours passed before 2"
// Additionally, it's verified that After expressions can only be satisfied if
// they're placed as Then.B.
func TestEvaluateSimpleAfter(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
	)

	now := time.Now()
	events := []driplang.Event{
		{Name: eventName1, Time: timey.AddHours(now, -20)},
		{Name: eventName1, Time: timey.AddHours(now, -10)},
		{Name: eventName2, Time: timey.AddHours(now, -5)},
		{Name: eventName1, Time: timey.AddHours(now, 0)},
	}

	tests := map[string]struct {
		expected bool
		expr     driplang.Expr
	}{
		"1 AFTER 10h": {
			// NOTE: a naked After expression (one that isn't the right side of
			// a THEN expression) should never be true since we don't have a
			// point in time to relate After.A, making After.D unused. The query
			// simply does not make sense.
			// It would probably be nicer to simply rework the language such
			// that this construct isn't possible. One suggestion I got was to
			// merge After and Then, making the duration an optional part of
			// Then.
			expected: false,
			expr: driplang.After{
				A: driplang.EventName(eventName1),
				D: driplang.Duration(10 * time.Hour),
			},
		},
		"1 THEN (2 AFTER 10h)": {
			expected: true,
			expr: driplang.Then{
				A: driplang.EventName(eventName1),
				B: driplang.After{
					A: driplang.EventName(eventName2),
					D: driplang.Duration(10 * time.Hour),
				},
			},
		},
		"1 THEN (2 AFTER 20h)": {
			expected: false,
			expr: driplang.Then{
				A: driplang.EventName(eventName1),
				B: driplang.After{
					A: driplang.EventName(eventName2),
					D: driplang.Duration(20 * time.Hour),
				},
			},
		},
		"1 THEN (1 AFTER 21h)": {
			expected: false,
			expr: driplang.Then{
				A: driplang.EventName(eventName1),
				B: driplang.After{
					A: driplang.EventName(eventName1),
					D: driplang.Duration(21 * time.Hour),
				},
			},
		},
		"2 THEN (2 AFTER 0s)": {
			expected: false,
			expr: driplang.Then{
				A: driplang.EventName(eventName2),
				B: driplang.After{
					A: driplang.EventName(eventName2),
					D: driplang.Duration(0),
				},
			},
		},
		"2 THEN (1 AFTER 5h)": {
			expected: true,
			expr: driplang.Then{
				A: driplang.EventName(eventName2),
				B: driplang.After{
					A: driplang.EventName(eventName1),
					D: driplang.Duration(5 * time.Hour),
				},
			},
		},
		"2 THEN (1 AFTER 5h+1ns)": {
			expected: false,
			expr: driplang.Then{
				A: driplang.EventName(eventName2),
				B: driplang.After{
					A: driplang.EventName(eventName1),
					D: driplang.Duration(5*time.Hour + 1*time.Nanosecond),
				},
			},
		},
		"2 THEN (2 AFTER -1h)": {
			// Verifies that the same event can't be consumed multiple times,
			// even if the Duration "works".
			expected: false,
			expr: driplang.Then{
				A: driplang.EventName(eventName2),
				B: driplang.After{
					A: driplang.EventName(eventName2),
					D: driplang.Duration(-1 * time.Hour),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.Evaluate(test.expr, events))
		})
	}
}

// TestEvaluateAfterNotAllHappensAfter verifies that an After expression only is
// satisfied if ALL events that satisfy sub-expressions happen AFTER After.D.
func TestEvaluateAfterNotAllHappensAfter(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
		eventName3 = "event3"
	)

	now := time.Now()
	expr := driplang.Then{
		A: driplang.EventName(eventName1),
		B: driplang.After{
			A: driplang.And{
				A: driplang.EventName(eventName2),
				B: driplang.EventName(eventName3),
			},
			D: driplang.Duration(10 * time.Hour),
		},
	}

	tests := map[string]struct {
		expected bool
		events   []driplang.Event
	}{
		"not all within duration constraint": {
			expected: false,
			events: []driplang.Event{
				{Name: eventName1, Time: timey.AddHours(now, -100)},
				// NOTE: eventName2 happens BEFORE the After.D constraint
				{Name: eventName2, Time: timey.AddHours(now, -99)},
				{Name: eventName3, Time: timey.AddHours(now, -1)},
			},
		},
		"all within duration constraint": {
			expected: true,
			events: []driplang.Event{
				{Name: eventName1, Time: timey.AddHours(now, -100)},
				// NOTE: eventName2 happens AFTER the After.D constraint
				{Name: eventName2, Time: timey.AddHours(now, -1)},
				{Name: eventName3, Time: timey.AddHours(now, -1)},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.Evaluate(expr, test.events))
		})
	}
}

// TestEvaluateAfterTimeOfUnrelatedEventIsUsed verifies that only time from
// events used to satisfy After.A are used when validating the After.D
// constraint
func TestEvaluateAfterTimeOfUnrelatedEventIsUsed(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
		eventName3 = "event3"
	)

	now := time.Now()
	events := []driplang.Event{
		{Name: eventName1, Time: timey.AddHours(now, -100)},

		// NOTE: this event should _NOT_ cause the evaluation of the expression
		// to become false
		{Name: eventName1, Time: timey.AddHours(now, -5)},

		{Name: eventName2, Time: timey.AddHours(now, -1)},
		{Name: eventName3, Time: timey.AddHours(now, -1)},
	}
	expr := driplang.Then{
		A: driplang.EventName(eventName1),
		B: driplang.After{
			A: driplang.And{
				A: driplang.EventName(eventName2),
				B: driplang.EventName(eventName3),
			},
			D: driplang.Duration(10 * time.Hour),
		},
	}

	require.True(t, driplang.Evaluate(expr, events))
}

// TestThenBExprBoundary verifies that a bounds check in the evaluation of Then
// is run, avoiding an out of bounds panic.
// The panic was caused by an index of -1 being returned in Then, which was
// incorrectly being used to index into a slice to gain the Time of the latest
// Then.A event.
func TestThenBExprBoundary(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
	)

	expr := driplang.Then{
		A: driplang.Not{A: driplang.EventName(eventName1)},
		B: driplang.EventName(eventName2),
	}
	require.Equal(t, true, driplang.Evaluate(expr, makeEvents(eventName2)))
}

// TestAfterSatisfiedByLaterEvent verifies that the After operator is satisfied
// only after both After.A AND After.D are satisfied, and allows events that
// satisfy only After.A to arrive before an event that satisfies both After.A
// and After.D. See test case "second event" for an example.
func TestAfterSatisfiedByLaterEvent(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
	)

	expr := driplang.Then{
		A: driplang.EventName(eventName1),
		B: driplang.After{
			A: driplang.EventName(eventName2),
			D: driplang.Duration(5 * time.Hour),
		},
	}

	now := time.Now()
	tests := map[string]struct {
		expected bool
		events   []driplang.Event
	}{
		"first event": {
			expected: true,
			events: []driplang.Event{
				{Name: eventName1, Time: timey.AddHours(now, -20)},
				{Name: eventName2, Time: timey.AddHours(now, 0)},
			},
		},
		"second event": {
			expected: true,
			events: []driplang.Event{
				{Name: eventName1, Time: timey.AddHours(now, -20)},
				{Name: eventName2, Time: timey.AddHours(now, -19)},
				{Name: eventName2, Time: timey.AddHours(now, -1)},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.Evaluate(expr, test.events))
		})
	}
}

// TestAfterNot verifies that After is satisfied when a certain event _has not
// happened_ in a certain timeframe, e.g. "1 happened and then 2 did not happen fro X time"
func TestAfterNot(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
	)

	expr := driplang.Then{
		A: driplang.And{
			A: driplang.EventName(eventName1),
			B: driplang.Not{
				A: driplang.EventName(eventName2),
			},
		},
		B: driplang.After{
			A: driplang.Not{
				A: driplang.EventName(eventName2),
			},
			D: driplang.Duration(5 * time.Hour),
		},
	}

	now := time.Now()
	tests := map[string]struct {
		expected bool
		events   []driplang.Event
	}{
		"2 exists": {
			expected: false,
			events: []driplang.Event{
				{Name: eventName1, Time: timey.AddHours(now, -20)},
				{Name: eventName1, Time: timey.AddHours(now, -10)},
				{Name: eventName2, Time: timey.AddHours(now, -1)},
			},
		},
		"2 not exists": {
			expected: true,
			events: []driplang.Event{
				{Name: eventName1, Time: timey.AddHours(now, -20)},
				{Name: eventName1, Time: timey.AddHours(now, -10)},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.Evaluate(expr, test.events))
		})
	}
}

func TestEvaluateSimpleNested(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
	)

	events := []driplang.Event{
		{Name: eventName1},
		{Name: eventName1},
		{Name: eventName2},
		{Name: eventName1},
	}

	tests := map[string]struct {
		expected bool
		expr     driplang.Expr
	}{
		"and not, false": {
			expected: false,
			expr: driplang.And{
				A: driplang.EventName(eventName1),
				B: driplang.Not{
					A: driplang.EventName(eventName2),
				},
			},
		},
		"and not, true": {
			expected: true,
			expr: driplang.And{
				A: driplang.EventName(eventName1),
				B: driplang.Not{
					A: driplang.EventName("does not exist"),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.Evaluate(test.expr, events))
		})
	}
}

func TestEvaluateThen(t *testing.T) {
	const (
		eventName1 = "event1"
		eventName2 = "event2"
		eventName3 = "event3"
	)

	events := makeEvents(
		eventName3,
		eventName2,
		eventName1,
		eventName2,
	)

	tests := map[string]struct {
		expected bool
		expr     driplang.Expr
	}{
		"nested, full sequence, true": {
			expected: true,
			expr: driplang.Then{
				A: driplang.EventName(eventName3),
				B: driplang.Then{
					A: driplang.EventName(eventName2),
					B: driplang.Then{
						A: driplang.EventName(eventName1),
						B: driplang.EventName(eventName2),
					},
				},
			},
		},
		"partial, start of sequence": {
			expected: true,
			expr: driplang.Then{
				A: driplang.EventName(eventName3),
				B: driplang.EventName(eventName2),
			},
		},
		"partial, mid of sequence": {
			expected: true,
			expr: driplang.Then{
				A: driplang.EventName(eventName2),
				B: driplang.EventName(eventName1),
			},
		},
		"partial, end of sequence": {
			expected: true,
			expr: driplang.Then{
				A: driplang.EventName(eventName1),
				B: driplang.EventName(eventName2),
			},
		},
		"nested, not satisfied": {
			expected: false,
			expr: driplang.Then{
				A: driplang.EventName(eventName1),
				B: driplang.Then{
					A: driplang.EventName(eventName2),
					B: driplang.EventName(eventName2),
				},
			},
		},
		"nested with not, not satisfied": {
			expected: false,
			expr: driplang.Then{
				A: driplang.EventName(eventName2),
				B: driplang.Then{
					A: driplang.Not{
						A: driplang.EventName(eventName1),
					},
					B: driplang.EventName(eventName2),
				},
			},
		},
		"nested with not, satisfied": {
			expected: true,
			expr: driplang.Then{
				A: driplang.EventName(eventName2),
				B: driplang.Then{
					A: driplang.Not{
						A: driplang.EventName(eventName3),
					},
					B: driplang.EventName(eventName2),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, driplang.Evaluate(test.expr, events))
		})
	}
}

func makeEvents(names ...string) []driplang.Event {
	events := make([]driplang.Event, len(names))
	for i, n := range names {
		events[i].Name = n
	}
	return events
}
