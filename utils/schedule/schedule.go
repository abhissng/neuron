package schedule

import (
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/helpers"
)

const (
	DefaultName       = "Schedule"
	SchedulerStarted  = "Schedule started"
	DefaultTimeFormat = "01-02-2006 03:04:05 PM"
	DefaultTimeZone   = "Asia/Kolkata"
)

// ScheduleProcessor defines an interface for types that can be scheduled.
type ScheduleProcessor interface {
	Start()
}

// Schedule is responsible for scheduling and executing a function at regular intervals.
type Schedule struct {
	name           string
	interval       time.Duration
	duration       time.Duration
	timeZone       string
	processor      ScheduleProcessor
	log            *log.Log
	StopChannel    chan struct{}
	isDebugEnabled bool
}

// NewSchedule creates a new Schedule with functional options.
func NewSchedule(processor ScheduleProcessor, opts ...Option) *Schedule {
	s := &Schedule{
		StopChannel:    make(chan struct{}),
		log:            log.NewBasicLogger(helpers.IsProdEnvironment()),
		processor:      processor,
		name:           DefaultName,
		timeZone:       DefaultTimeZone,
		isDebugEnabled: false,
	}

	// Apply functional options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Run starts the schedule.
func (s *Schedule) Run() {
	if s.processor == nil {
		s.log.Error("Schedule processor is nil")
		return
	}

	ticker := time.NewTicker(s.interval)
	location, _ := time.LoadLocation(s.timeZone)

	var endTime time.Time
	if s.duration > 0 { // Only set endTime if duration is specified
		endTime = time.Now().In(location).Add(s.duration)
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				if s.duration > 0 && time.Now().After(endTime) {
					ticker.Stop()
					close(s.StopChannel)
					return
				}

				now := time.Now()
				l, _ := time.LoadLocation(s.timeZone)
				t := now.In(l)
				nextMessage := `Next schedule will start in (` + s.timeZone + `) ` + t.Add(s.interval).Format(DefaultTimeFormat)
				s.log.Info(SchedulerStarted, log.Any(s.name, nextMessage))
				s.processor.Start()
			case <-s.StopChannel:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop gracefully shuts down the schedule.
func (s *Schedule) Stop() {
	select {
	case <-s.StopChannel: // Prevent closing twice
	default:
		close(s.StopChannel)
	}
}

// IsDebugEnabled returns if the debug is enabled.
func (s *Schedule) IsDebugEnabled() bool {
	return s.isDebugEnabled

}
