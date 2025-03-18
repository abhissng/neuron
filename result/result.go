package result

import (
	"strings"

	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/utils/constant"
)

// Result is a generic interface that can represent either a success or an error.
type Result[T any] interface {
	// IsSuccess returns true if the result is a success, false otherwise.
	IsSuccess() bool
	// IsError returns true if the result is an error, false otherwise.
	IsError() bool
	// Value returns the success value and error value if there is error any.
	Value() (*T, blame.Blame)
	// Error returns the error value.
	Error() blame.Blame
	//Redirect is used only if an api is redirecting to other service/webapp
	Redirect() (string, bool)
	// ToValue returns the success value if the result is a success, nil otherwise.
	ToValue() *T
}

// Success represents a successful result.
type Success[T any] struct {
	Val         *T
	RedirectURL string
}

// NewSuccess creates a new success result.
func NewSuccess[T any](value *T) Result[T] {
	return &Success[T]{Val: value}
}

// IsSuccess implements Result.
func (s Success[T]) IsSuccess() bool {
	return true
}

// Value implements Result.
func (s Success[T]) Value() (*T, blame.Blame) {
	return s.Val, nil
}

// IsError implements Result.
func (s Success[T]) IsError() bool {
	return false
}

// Error implements Result.
func (s Success[T]) Error() blame.Blame {
	return blame.NewBasicBlame("success-cannot-be-error").WithComponent(constant.ErrLibrary)
}

// Redirect implements Result.
func (s Success[T]) Redirect() (string, bool) {
	return s.RedirectURL, strings.TrimSpace(s.RedirectURL) != ""
}

// ToValue returns the success value if the result is a success, nil otherwise.
func (s Success[T]) ToValue() *T {
	return s.Val
}

// Failure represents an error result.
type Failure[T any] struct {
	Val         *T
	Err         blame.Blame
	RedirectURL string
}

// NewError creates a new Failure result.
func NewFailure[T any](err blame.Blame) Result[T] {
	return &Failure[T]{Err: err}
}

func NewFailureWithValue[T any](value *T, err blame.Blame) Result[T] {
	return &Failure[T]{Val: value, Err: err}
}

// IsSuccess implements Result.
func (f Failure[T]) IsSuccess() bool {
	return false
}

// IsError implements Result.
func (f Failure[T]) IsError() bool {
	return true
}

// Value implements Result.
func (f Failure[T]) Value() (*T, blame.Blame) {
	if f.Val == nil {
		return nil, f.Err
	}
	return f.Val, f.Err
}

// Error implements Result.
func (f Failure[T]) Error() blame.Blame {
	return f.Err
}

// Redirect implements Result.
func (f Failure[T]) Redirect() (string, bool) {
	return f.RedirectURL, strings.TrimSpace(f.RedirectURL) != ""
}

// Failure[T] implements ToValue
func (f Failure[T]) ToValue() *T {
	return nil
}

// ToResult cast the value or error to Result
func ToResult[T any](value *T, err blame.Blame) Result[T] {
	if err != nil {
		return NewFailure[T](err)
	}
	return NewSuccess[T](value)
}

// CastFailure attempts to cast the failure to a specific type E and returns a new Result.
func CastFailure[T, E any](r Result[T]) Result[E] {
	if r.IsSuccess() {
		return NewFailure[E](blame.NewBasicBlame("SUCCESS_CANNOT_PRODUCE_ERROR"))
	}
	_, err := r.Value()
	return NewFailure[E](err)
}

// MapError maps the error of a Result to a new Result with a different type.
func MapError[T, R any](r Result[T], mapFn func(error) blame.Blame) Result[R] {
	if r.IsSuccess() {
		return NewFailure[R](blame.NewBasicBlame("SUCCESS_CANNOT_MAP_WITH_ERROR"))
	}
	return NewFailure[R](mapFn(r.Error()))
}

// NewRedirectSuccess creates a new Success result with a redirect URL.
func NewRedirectSuccess[T any](url string) Success[T] {
	return Success[T]{
		RedirectURL: url,
	}
}

// NewFailureWithRedirect creates a new Failure result with a redirect URL.
func NewFailureWithRedirect[T any](err blame.Blame, url string) Failure[T] {
	return Failure[T]{
		Err:         err,
		RedirectURL: url,
	}
}

// CastRedirectFailure casts the failure to a specific type R and returns a new Result.
func CastRedirectFailure[T any, R any](result Result[T], url string) Result[R] {
	_, err := result.Value()
	return NewFailureWithRedirect[R](err, url)
}
