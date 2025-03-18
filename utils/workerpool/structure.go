package workerpool

import (
	"math"
	"sync"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/helpers"
)

// WorkerPool represents a generic worker pool for concurrent task execution.
type WorkerPool[T any, U any] struct {
	// ctx          *context.ServiceContext
	numWorkers   int                       // Number of workers
	taskQueue    chan result.Task[T]       // Channel for tasks
	resultQueue  chan result.TaskResult[U] // Channel for results
	batchSize    int                       // Number of tasks in a batch
	batchDone    chan int                  // Signal when a batch is completed (sends batch ID)
	wg           sync.WaitGroup            // WaitGroup to wait for all workers to finish
	mu           sync.Mutex                // Mutex to protect shared state
	batchCounter int                       // Counter to generate unique batch IDs
	taskCounter  int                       // Counter to generate unique task IDs
	batchTasks   map[int]int               // Track the number of completed tasks per batch
	processor    TaskProcessor[T, U]       // Task processor
	log          *log.Log                  // Use log.Log

}

// NewWorkerPool creates a new WorkerPool with the provided options.
func NewWorkerPool[T any, U any](processor TaskProcessor[T, U], options ...Option[T, U]) *WorkerPool[T, U] {
	// Default configuration
	wp := &WorkerPool[T, U]{
		numWorkers:   5,                                               // Default number of workers
		taskQueue:    make(chan result.Task[T], 100),                  // Default task queue size
		resultQueue:  make(chan result.TaskResult[U], 100),            // Default result queue size
		batchSize:    5,                                               // Default batch size
		batchDone:    make(chan int),                                  // Channel to signal batch completion (with batch ID)
		batchCounter: 1,                                               // Start batch IDs from 1
		taskCounter:  1,                                               // Start task IDs from 1
		batchTasks:   make(map[int]int),                               // Initialize batch task tracker
		processor:    processor,                                       // Task processor
		log:          log.NewBasicLogger(helpers.IsProdEnvironment()), // Basic Logger (by default production)
	}

	// Apply options to override defaults
	for _, option := range options {
		option(wp)
	}

	// Start the worker pool
	wp.start()
	return wp
}

// TaskProcessor defines an interface for processing tasks.
type TaskProcessor[T any, U any] interface {
	Process(input T) result.Result[U]
}

// Option is a function type for configuring the WorkerPool.
type Option[T any, U any] func(*WorkerPool[T, U])

// WithNumWorkers sets the number of workers in the pool.
func WithNumWorkers[T any, U any](numWorkers int) Option[T, U] {
	return func(wp *WorkerPool[T, U]) {
		wp.numWorkers = numWorkers
	}
}

// WithTaskQueueSize sets the size of the task queue.
func WithTaskQueueSize[T any, U any](size int) Option[T, U] {
	return func(wp *WorkerPool[T, U]) {
		wp.taskQueue = make(chan result.Task[T], size)
	}
}

// WithResultQueueSize sets the size of the result queue.
func WithResultQueueSize[T any, U any](size int) Option[T, U] {
	return func(wp *WorkerPool[T, U]) {
		wp.resultQueue = make(chan result.TaskResult[U], size)
	}
}

// WithBatchSize sets the size of a batch.
func WithBatchSize[T any, U any](batchSize int) Option[T, U] {
	return func(wp *WorkerPool[T, U]) {
		wp.batchSize = batchSize
	}
}

// WithLogger sets logger for worker pool.
func WithLogger[T any, U any](isProd bool) Option[T, U] {
	return func(wp *WorkerPool[T, U]) {
		wp.log = log.NewBasicLogger(isProd)
	}
}

// WorkerPoolConfig holds dynamically calculated worker pool settings.
type WorkerPoolConfig struct {
	NumWorkers      int
	TaskQueueSize   int
	ResultQueueSize int
	BatchSize       int
}

// CalculateOptimalWorkerPoolConfig computes optimal worker pool settings based on task count.
func CalculateOptimalWorkerPoolConfig(numTasks int) WorkerPoolConfig {
	if numTasks <= 0 {
		return WorkerPoolConfig{
			NumWorkers:      1,
			TaskQueueSize:   0,
			ResultQueueSize: 0,
			BatchSize:       1,
		}
	}

	// Determine workers based on sqrt(numTasks), limited between 1 and 16
	numWorkers := int(math.Max(1, math.Min(16, math.Sqrt(float64(numTasks)))))

	// Task queue and result queue should be equal to numTasks
	taskQueueSize := numTasks
	resultQueueSize := numTasks

	// Batch size should evenly distribute tasks across workers
	batchSize := int(math.Max(1, math.Ceil(float64(numTasks)/float64(numWorkers))))

	return WorkerPoolConfig{
		NumWorkers:      numWorkers,
		TaskQueueSize:   taskQueueSize,
		ResultQueueSize: resultQueueSize,
		BatchSize:       batchSize,
	}
}
