// Package common provides functional programming utilities and types
package common

// Option represents a value that may or may not be present
// This is similar to Maybe/Option types in functional languages
type Option[T any] struct {
	value    T
	hasValue bool
}

// Some creates an Option with a value
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, hasValue: true}
}

// None creates an empty Option
func None[T any]() Option[T] {
	var zero T
	return Option[T]{value: zero, hasValue: false}
}

// IsSome returns true if the Option contains a value
func (o Option[T]) IsSome() bool {
	return o.hasValue
}

// IsNone returns true if the Option is empty
func (o Option[T]) IsNone() bool {
	return !o.hasValue
}

// Value returns the contained value. Should only be called if IsSome() is true
func (o Option[T]) Value() T {
	return o.value
}

// ValueOr returns the contained value or the provided default if None
func (o Option[T]) ValueOr(defaultValue T) T {
	if o.hasValue {
		return o.value
	}
	return defaultValue
}

// MapOption applies a function to the value if the Option is Some, otherwise returns None
func MapOption[T, U any](o Option[T], fn func(T) U) Option[U] {
	if o.IsNone() {
		return None[U]()
	}
	return Some(fn(o.value))
}

// BindOption applies a function that returns an Option to the value if the Option is Some
func BindOption[T, U any](o Option[T], fn func(T) Option[U]) Option[U] {
	if o.IsNone() {
		return None[U]()
	}
	return fn(o.value)
}

// FilterOption returns the Option if it's Some and the predicate is true, otherwise returns None
func FilterOption[T any](o Option[T], predicate func(T) bool) Option[T] {
	if o.IsNone() || !predicate(o.value) {
		return None[T]()
	}
	return o
}

// OrElseOption returns the Option if it's Some, otherwise returns the alternative Option
func OrElseOption[T any](o Option[T], alternative Option[T]) Option[T] {
	if o.IsSome() {
		return o
	}
	return alternative
}

// ToResult converts Option to Result with the provided error for None case
func (o Option[T]) ToResult(err error) Result[T] {
	if o.IsSome() {
		return Ok(o.value)
	}
	return Err[T](err)
}

// FromResult converts Result to Option, discarding error information
func FromResult[T any](r Result[T]) Option[T] {
	if r.IsOk() {
		return Some(r.value)
	}
	return None[T]()
}

// ForEachOption executes a side effect function if the Option is Some
func ForEachOption[T any](o Option[T], fn func(T)) {
	if o.IsSome() {
		fn(o.value)
	}
}

// FoldOption extracts the value from Option by applying one of two functions
func FoldOption[T, U any](o Option[T], onSome func(T) U, onNone func() U) U {
	if o.IsSome() {
		return onSome(o.value)
	}
	return onNone()
}
