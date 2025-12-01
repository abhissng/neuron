package schedule

import (
	"time"

	"github.com/abhissng/neuron/adapters/log"
)

// Option is a functional option type for configuring the Schedule.
type Option func(*Schedule)

// WithName sets the name of the scheduler.
func WithName(name string) Option {
	return func(s *Schedule) {
		s.name = name
	}
}

// WithInterval sets the interval between executions.
func WithInterval(interval time.Duration) Option {
	return func(s *Schedule) {
		s.interval = interval
	}
}

// WithDuration sets the total duration for which the schedule will run.
func WithDuration(duration time.Duration) Option {
	return func(s *Schedule) {
		s.duration = duration
	}
}

// WithTimeZone sets the schedule's timezone.
func WithTimeZone(timeZone string) Option {
	return func(s *Schedule) {
		s.timeZone = timeZone
	}
}

// WithLogger sets the logger.
func WithLogger(log *log.Log) Option {
	return func(s *Schedule) {
		s.log = log
	}
}

// WithDebugEnabled sets the Debug Mode.
func WithDebugEnabled(enabled bool) Option {
	return func(s *Schedule) {
		s.isDebugEnabled = enabled
	}
}

// WithStartAt sets an exact start time (time.Time should include location).
// If time.Time has no location, it will be used as-is (UTC).
func WithStartAt(t time.Time) Option {
	return func(s *Schedule) {
		tt := t
		s.startAt = &tt
	}
}

// WithStartAtTime sets start time using hour and minute in the schedule's timezone.
// Example: WithStartAtTime(1, 0) => starts at 01:00 in s.timeZone (today or next day if already passed).
func WithStartAtTime(hour, minute int) Option {
	return func(s *Schedule) {
		loc, err := time.LoadLocation(s.timeZone)
		if err != nil {
			// fallback to local if timezone not found
			loc = time.Local
		}
		now := time.Now().In(loc)
		start := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		s.startAt = &start
	}
}
