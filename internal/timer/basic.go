package timer

import "time"

type BasicTimer struct{}

func NewBasicTimer() *BasicTimer {
	return &BasicTimer{}
}

func (t *BasicTimer) Now() time.Time {
	return time.Now()
}
