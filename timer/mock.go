package timer

import "time"

type MockTimer struct {
	t time.Time
}

func NewMockTimer(ts string) *MockTimer {
	return &MockTimer{*ParseStringToTime(ts)}
}

func (t *MockTimer) Now() time.Time {
	return t.t
}

func (t *MockTimer) SetTime(nt time.Time) {
	t.t = nt
}

func ParseStringToTime(ts string) *time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.000Z", ts)
	return &t
}
