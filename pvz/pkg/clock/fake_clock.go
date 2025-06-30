package clock

import (
	"time"
)

var _ Clock = &FakeClock{}

var fixed = time.Date(2025, 6, 26, 12, 0, 0, 0, time.UTC)

type FakeClock struct{}

func (c *FakeClock) Now() time.Time {
	return fixed
}

func (c *FakeClock) After(d time.Duration) time.Time {
	return fixed.Add(d)
}
