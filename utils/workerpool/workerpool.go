package workerpool

import (
	"fmt"

	"github.com/abhissng/neuron/result"
)

// start initializes the worker pool and starts the workers.
func (wp *WorkerPool[T, U]) start() {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go func(workerID int) {
			defer wp.wg.Done()
			for task := range wp.taskQueue {
				// Process the task using the provided processor
				res := wp.processor.Process(task.Input)
				// Wrap the result with TaskID and BatchID
				taskResult := result.NewTaskResult(task.ID, task.BatchID, res)
				wp.resultQueue <- taskResult

				// Track the number of completed tasks for this batch
				wp.mu.Lock()
				wp.batchTasks[task.BatchID]++
				if wp.batchTasks[task.BatchID] == wp.batchSize {
					// Signal batch completion
					wp.batchDone <- task.BatchID
					delete(wp.batchTasks, task.BatchID) // Clean up the batch
				}
				wp.mu.Unlock()
			}
		}(i + 1)
	}
}

// SubmitBatch submits a batch of tasks to the worker pool and returns the batch ID.
func (wp *WorkerPool[T, U]) SubmitBatch(tasks []result.Task[T]) int {
	wp.mu.Lock()
	batchID := wp.batchCounter
	wp.batchCounter++
	wp.mu.Unlock()

	for i := range tasks {
		tasks[i].BatchID = batchID
		wp.Submit(tasks[i])
	}

	return batchID
}

// Submit adds a task to the worker pool.
func (wp *WorkerPool[T, U]) Submit(task result.Task[T]) {
	wp.taskQueue <- task
}

// Results returns a channel for collecting task results.
func (wp *WorkerPool[T, U]) Results() <-chan result.TaskResult[U] {
	return wp.resultQueue
}

// BatchDone returns a channel to signal batch completion (with batch ID).
func (wp *WorkerPool[T, U]) BatchDone() <-chan int {
	return wp.batchDone
}

// Shutdown gracefully shuts down the worker pool.
func (wp *WorkerPool[T, U]) Shutdown() {
	close(wp.taskQueue)   // Close the task queue to stop accepting new tasks
	wp.wg.Wait()          // Wait for all workers to finish
	close(wp.resultQueue) // Close the result queue
	close(wp.batchDone)   // Close the batch completion channel
	_ = wp.log.Sync()     // Close the logger
}

// GetBatchCount returns the number of batches created so far.
func (wp *WorkerPool[T, U]) GetBatchCount() int {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.batchCounter - 1 // Subtract 1 because batchCounter starts at 1
}

// generateTaskID generates a unique task ID.
func (wp *WorkerPool[T, U]) generateTaskID() int {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	taskID := wp.taskCounter
	wp.taskCounter++
	return taskID
}

// createBatch divides tasks into batches of the specified size.
func (wp *WorkerPool[T, U]) createBatch(tasks []T) [][]T {
	var batches [][]T
	for i := 0; i < len(tasks); i += wp.batchSize {
		end := i + wp.batchSize
		if end > len(tasks) {
			end = len(tasks)
		}
		batches = append(batches, tasks[i:end])
	}
	return batches
}

// ExecuteBatch divides tasks into batches and submits them to the worker pool.
func (wp *WorkerPool[T, U]) Execute(tasks []T) {
	// Divide tasks into batches
	batches := wp.createBatch(tasks)

	// Submit each batch to the worker pool
	for _, batch := range batches {
		tasks := make([]result.Task[T], len(batch))
		for i, input := range batch {
			tasks[i] = result.Task[T]{
				ID:    wp.generateTaskID(), // Generate a unique task ID
				Input: input,
			}
		}
		batchID := wp.SubmitBatch(tasks)
		wp.log.Info(fmt.Sprintf("Submitted Batch %d with %d tasks\n", batchID, len(tasks)))
	}
}
