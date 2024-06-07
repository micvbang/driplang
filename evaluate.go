package driplang

import (
	"time"
)

var minTime = time.Time{}

// Evaluate checks if Expr is satisfied by the given slice of Events
// Assumes that events are sorted by Event.Time.
func Evaluate(e Expr, events []Event) bool {
	_, satisfied, _ := evaluate(e, events, minTime)
	return satisfied
}

func EvaluateWithIndex(e Expr, events []Event) (int, bool) {
	i, satisfied, _ := evaluate(e, events, minTime)
	return i, satisfied
}

func evaluate(e Expr, evs []Event, mustBeAfter time.Time) (evsIndex int, satisfied, timeAfter bool) {
	switch v := e.(type) {
	case EventName:
		name := string(v)

		// Check for satisfied + timeAfter
		for i, ev := range evs {
			if name == ev.Name && ev.Time.Sub(mustBeAfter) >= 0 {
				return i, true, true
			}
		}

		// Check for satisfied
		for i, ev := range evs {
			if name == ev.Name {
				return i, true, false
			}
		}

		// If event doesn't exist, we still have to report if the time is
		// after. This is important for NOT expressions where an event is
		// expected to not be present (and we therefore can't compare its'
		// arrival time)
		return -1, false, time.Now().After(mustBeAfter)

	case Or:
		ai, a, aAfter := evaluate(v.A, evs, mustBeAfter)
		bi, b, bAfter := evaluate(v.B, evs, mustBeAfter)
		if a && b {
			// Neither index will be < 0, use the minimum one
			return min(ai, bi), true, aAfter || bAfter
		}

		// One index is < 0, use the maximum one
		return max(ai, bi), a || b, aAfter || bAfter

	case And:
		ai, a, aAfter := evaluate(v.A, evs, mustBeAfter)
		bi, b, bAfter := evaluate(v.B, evs, mustBeAfter)
		if a && b {
			// Both indices >= 0, use maximum one
			return max(ai, bi), true, aAfter && bAfter
		}
		// One index is false, use neither one
		return -1, false, false

	case Not:
		ai, a, aAfter := evaluate(v.A, evs, mustBeAfter)
		if a {
			// Invert a
			return ai, false, aAfter
		}
		return -1, true, aAfter

	case Then:
		for i := len(evs); i > 0; i-- {
			ai, a, aAfter := evaluate(v.A, evs[:i], mustBeAfter)
			if !a {
				continue
			}

			bMustBeAfter := mustBeAfter
			if ai >= 0 && ai < len(evs) {
				bMustBeAfter = evs[ai].Time
			}

			bi, b, bAfter := evaluate(v.B, evs[ai+1:], bMustBeAfter)
			if a && b {
				return ai + bi + 1, true, aAfter && bAfter
			}
		}
		return -1, false, false

	case After:
		if mustBeAfter == minTime {
			// if t is minTime then `After` has not been in a `Then.B` clause,
			// which is a requirement; otherwise we don't have a point in time
			// to compare v.A to.
			return -1, false, false
		}

		ai, a, aAfter := evaluate(v.A, evs, mustBeAfter.Add(time.Duration(v.D)))
		if a {
			return ai, true && aAfter, aAfter
		}

		return -1, false, false

	default:
		return -1, false, false
	}
}
