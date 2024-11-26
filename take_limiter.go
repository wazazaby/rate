package rate

import "time"

type TakeLimiter struct {
	tick <-chan time.Time
}

func NewTakeLimiter(rps uint) TakeLimiter {
	return TakeLimiter{
		tick: time.Tick(time.Second / time.Duration(rps)),
	}
}

func (l TakeLimiter) Take() time.Time {
	return <-l.tick
}

func (l TakeLimiter) TryTake() (time.Time, bool) {
	select {
	case t := <-l.tick:
		return t, true
	default:
		return time.Time{}, false
	}
}
