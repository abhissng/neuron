package schedule

import (
	"fmt"
	"sync"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/helpers"
)

const (
	DefaultName       = "Schedule"
	SchedulerStarted  = "Schedule started"
	DefaultTimeFormat = "02-01-2006 03:04:05 PM"
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
	startAt        *time.Time
	timeZone       string
	processor      ScheduleProcessor
	log            *log.Log
	StopChannel    chan struct{}
	isDebugEnabled bool
	stopOnce       sync.Once
}

// NewSchedule creates a new Schedule with functional options.
func NewSchedule(processor ScheduleProcessor, opts ...Option) *Schedule {
	s := &Schedule{
		StopChannel:    make(chan struct{}),
		log:            log.NewBasicLogger(helpers.IsProdEnvironment(), true),
		processor:      processor,
		name:           DefaultName,
		timeZone:       DefaultTimeZone,
		isDebugEnabled: false,
	}

	// Apply functional options (these can use defaults above, like timeZone).
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

	// Validate interval
	if s.interval <= 0 {
		s.log.Error("Schedule interval must be > 0")
		return
	}

	// Ensure timezone can be used
	location, err := time.LoadLocation(s.timeZone)
	if err != nil {
		// if invalid timezone, fallback to local and log
		s.log.Warn(fmt.Sprintf("invalid timezone '%s', falling back to Local", s.timeZone))
		location = time.Local
	}

	// If startAt is set but has no location or different location, normalize it to schedule timezone.
	if s.startAt != nil {
		// Normalize pointer value into schedule timezone:
		start := s.startAt.In(location)

		// If start time already passed for today, roll to next day
		now := time.Now().In(location)
		if now.After(start) || now.Equal(start) {
			start = start.Add(24 * time.Hour)
		}
		s.startAt = &start

		s.log.Info("Scheduled first run at", log.Any(s.name, s.startAt.Format(DefaultTimeFormat)))

		// Wait until startAt, then run first job and continue with interval loop.
		go func() {
			// Compute wait duration
			now := time.Now().In(location)
			wait := s.startAt.Sub(now)
			timer := time.NewTimer(wait)
			defer timer.Stop()

			select {
			case <-timer.C:
				// First run
				s.log.Info("Executing first scheduled run", log.Any(s.name, s.startAt.Format(DefaultTimeFormat)))
				s.processor.Start()

				// After first run, continue with interval loop (duration should count from first run)
				s.startIntervalLoop()
			case <-s.StopChannel:
				// stop requested before first run
				return
			}
		}()

		return
	}

	// No startAt -> start immediately with interval loop
	s.startIntervalLoop()
}

// startIntervalLoop runs the ticker loop that executes processor.Start on every tick.
func (s *Schedule) startIntervalLoop() {
	location, err := time.LoadLocation(s.timeZone)
	if err != nil {
		location = time.Local
	}

	// Setup end time (counts from the moment this function is called)
	var endTime time.Time
	if s.duration > 0 {
		endTime = time.Now().In(location).Add(s.duration)
	}

	ticker := time.NewTicker(s.interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Check duration expiry
				if s.duration > 0 && time.Now().In(location).After(endTime) {
					s.log.Info("Schedule duration completed, stopping schedule", log.Any(s.name, "duration_expired"))
					// Use Stop to close StopChannel safely
					s.Stop()
					return
				}

				now := time.Now().In(location)
				next := now.Add(s.interval).Format(DefaultTimeFormat)
				nextMessage := fmt.Sprintf("Next schedule will start in (%s) %s", s.timeZone, next)
				s.log.Info(SchedulerStarted, log.Any(s.name, nextMessage))

				// Execute the processor
				s.processor.Start()

			case <-s.StopChannel:
				// Graceful stop requested
				s.log.Info("Stop signal received, stopping ticker", log.Any(s.name, "stopped"))
				return
			}
		}
	}()
}

// Stop gracefully shuts down the schedule.
func (s *Schedule) Stop() {
	s.stopOnce.Do(func() { close(s.StopChannel) })
}

// IsDebugEnabled returns if the debug is enabled.
func (s *Schedule) IsDebugEnabled() bool {
	return s.isDebugEnabled
}

/*
Usage example:

	package main

	import (
		"fmt"
		"time"

		"your_module_path/schedule"
	)

	type myProcessor struct{}
	func (m *myProcessor) Start() {
		fmt.Println("job running at", time.Now())
	}

	func main() {
		proc := &myProcessor{}

		// Example: start at 01:00 (schedule timezone Asia/Kolkata), then run every 2 hours.
		s := schedule.NewSchedule(proc,
			schedule.WithName("NightlyJob"),
			schedule.WithTimeZone("Asia/Kolkata"),
			schedule.WithInterval(2*time.Hour),
			schedule.WithStartAtTime(1, 0),   // first run at 01:00 in Asia/Kolkata
			schedule.WithDuration(24*time.Hour), // run for next 24 hours (optional)
		)

		s.Run()

		// Let it run for a while for demo (in real apps block on signal or use context)
		time.Sleep(6 * time.Hour)
		s.Stop()
	}
*/
