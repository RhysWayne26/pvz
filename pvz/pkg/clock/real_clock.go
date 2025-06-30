package clock

import "time"

var _ Clock = &RealClock{}

type RealClock struct{}

func (c *RealClock) Now() time.Time {
	return time.Now()
}

func (c *RealClock) After(d time.Duration) time.Time {
	return c.Now().Add(d)
}
