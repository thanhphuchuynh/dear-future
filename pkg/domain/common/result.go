// Package common provides functional programming utilities and types
package common

import "fmt"

// Result represents a computation that can either succeed with a value or fail with an error
// This is similar to Either/Result types in functional languages
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a successful Result with the given value
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, err: nil}
}

// Err creates a failed Result with the given error
func Err[T any](err error) Result[T] {
	var zero T
	return Result[T]{value: zero, err: err}
}

// IsOk returns true if the Result represents a successful computation
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr returns true if the Result represents a failed computation
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Value returns the success value. Should only be called if IsOk() is true
func (r Result[T]) Value() T {
	return r.value
}

// Error returns the error. Should only be called if IsErr() is true
func (r Result[T]) Error() error {
	return r.err
}

// Unwrap returns both the value and error, similar to Go's common pattern
func (r Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

// Map applies a function to the value if the Result is Ok, otherwise returns the error
// This is a functor operation in functional programming
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.IsErr() {
		return Err[U](r.err)
	}
	return Ok(fn(r.value))
}

// Bind applies a function that returns a Result to the value if the Result is Ok
// This is a monadic bind operation (also known as flatMap)
func Bind[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.IsErr() {
		return Err[U](r.err)
	}
	return fn(r.value)
}

// MapErr applies a function to the error if the Result is Err, otherwise returns the value unchanged
func MapErr[T any](r Result[T], fn func(error) error) Result[T] {
	if r.IsOk() {
		return r
	}
	return Err[T](fn(r.err))
}

// OrElse returns the Result if it's Ok, otherwise returns the alternative Result
func OrElse[T any](r Result[T], alternative Result[T]) Result[T] {
	if r.IsOk() {
		return r
	}
	return alternative
}

// Filter returns the Result if it's Ok and the predicate is true, otherwise returns an error
func Filter[T any](r Result[T], predicate func(T) bool, errorMsg string) Result[T] {
	if r.IsErr() {
		return r
	}
	if predicate(r.value) {
		return r
	}
	return Err[T](fmt.Errorf(errorMsg))
}

// Fold extracts the value from Result by applying one of two functions
func Fold[T, U any](r Result[T], onSuccess func(T) U, onError func(error) U) U {
	if r.IsOk() {
		return onSuccess(r.value)
	}
	return onError(r.err)
}

// ForEach executes a side effect function if the Result is Ok
func ForEach[T any](r Result[T], fn func(T)) {
	if r.IsOk() {
		fn(r.value)
	}
}

// Collect converts a slice of Results into a Result of slice
// Fails fast - if any Result is an error, returns that error
func Collect[T any](results []Result[T]) Result[[]T] {
	values := make([]T, len(results))
	for i, result := range results {
		if result.IsErr() {
			return Err[[]T](result.err)
		}
		values[i] = result.value
	}
	return Ok(values)
}
