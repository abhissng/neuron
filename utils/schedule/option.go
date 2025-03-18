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
func WithDebugEnabled() Option {
	return func(s *Schedule) {
		s.isDebugEnabled = true
	}
}
