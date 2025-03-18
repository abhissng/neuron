package result

// TaskResult wraps a Result with TaskID and BatchID.
type TaskResult[T any] struct {
	TaskID  int
	BatchID int
	Output  Result[T]
}

// NewTaskResult creates a new TaskResult.
func NewTaskResult[T any](taskID int, batchID int, output Result[T]) TaskResult[T] {
	return TaskResult[T]{
		TaskID:  taskID,
		BatchID: batchID,
		Output:  output,
	}
}

// Task represents a task to be executed by the worker pool.
type Task[T any] struct {
	ID      int
	BatchID int
	Input   T
}

// TaskProcessor defines an interface for processing tasks.
type TaskProcessor[T any, U any] interface {
	Process(input T) Result[U]
}
