package widgets

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
)

type Anim struct {
	Next      func(int) (int, time.Duration)
	animating bool
	next      time.Time
	start     int
	end       int
	v, w      int // current and next value
}

func (a *Anim) Animating() bool {
	return a.animating
}

// Start starts the animation, ending when the value crosses the end value.
func (a *Anim) Start(start, end int) {
	a.animating = true
	a.next = time.Time{}
	a.start = start
	a.end = end
	a.v = start
	a.w = start
}

// Value returns the current value, without animating.
func (a *Anim) Value() int {
	return a.v
}

// Animate returns the value and moves the animation forward.
func (a *Anim) Animate(gtx layout.Context) int {
	v := a.v
	switch asc := a.start < a.end; {
	case !a.animating:
		return a.v
	case asc && a.v >= a.end || !asc && a.v <= a.end:
		// Animation done.
		a.animating = false
		a.next = time.Time{}
		a.v = a.end
		op.InvalidateOp{}.Add(gtx.Ops)
		return a.end
	case a.next.IsZero():
		// First animation.
		a.next = gtx.Now
		nv, d := a.Next(a.v)
		a.next = a.next.Add(d)
		a.w = nv
	case a.next.Before(gtx.Now) || a.next.Equal(gtx.Now):
		// Next animation.
		nv, d := a.Next(a.w)
		a.next = a.next.Add(d)
		a.v, a.w = a.w, nv
		v = a.v
	default:
		// Current animation.
	}
	op.InvalidateOp{At: a.next}.Add(gtx.Ops)
	return v
}
