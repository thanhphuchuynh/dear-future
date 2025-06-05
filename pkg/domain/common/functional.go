// Package common provides functional programming utilities and types
package common

// Pipe composes two functions that return Results
func Pipe[T, U, V any](
	input Result[T],
	fn func(T) Result[U],
) Result[U] {
	return Bind(input, fn)
}

// Pipe2 composes three functions that return Results
func Pipe2[T, U, V, W any](
	input Result[T],
	fn1 func(T) Result[U],
	fn2 func(U) Result[V],
) Result[V] {
	return Bind(Bind(input, fn1), fn2)
}

// Pipe3 composes four functions that return Results
func Pipe3[T, U, V, W, X any](
	input Result[T],
	fn1 func(T) Result[U],
	fn2 func(U) Result[V],
	fn3 func(V) Result[W],
) Result[W] {
	return Bind(Bind(Bind(input, fn1), fn2), fn3)
}

// Compose creates a new function by composing two functions
func Compose[T, U, V any](f func(U) V, g func(T) U) func(T) V {
	return func(x T) V {
		return f(g(x))
	}
}

// Curry converts a function with two parameters into a curried function
func Curry[T, U, V any](fn func(T, U) V) func(T) func(U) V {
	return func(t T) func(U) V {
		return func(u U) V {
			return fn(t, u)
		}
	}
}

// MapSlice applies a function to each element of a slice
func MapSlice[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, item := range slice {
		result[i] = fn(item)
	}
	return result
}

// FilterSlice filters a slice based on a predicate function
func FilterSlice[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// ReduceSlice reduces a slice to a single value using an accumulator function
func ReduceSlice[T, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, item := range slice {
		result = fn(result, item)
	}
	return result
}

// FindSlice finds the first element that satisfies the predicate
func FindSlice[T any](slice []T, predicate func(T) bool) Option[T] {
	for _, item := range slice {
		if predicate(item) {
			return Some(item)
		}
	}
	return None[T]()
}

// AllSlice returns true if all elements satisfy the predicate
func AllSlice[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// AnySlice returns true if any element satisfies the predicate
func AnySlice[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if predicate(item) {
			return true
		}
	}
	return false
}

// PartitionSlice splits a slice into two slices based on a predicate
func PartitionSlice[T any](slice []T, predicate func(T) bool) ([]T, []T) {
	var trueSlice, falseSlice []T
	for _, item := range slice {
		if predicate(item) {
			trueSlice = append(trueSlice, item)
		} else {
			falseSlice = append(falseSlice, item)
		}
	}
	return trueSlice, falseSlice
}

// GroupBySlice groups elements by a key function
func GroupBySlice[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range slice {
		key := keyFn(item)
		result[key] = append(result[key], item)
	}
	return result
}

// UniqueSlice returns a slice with unique elements based on a key function
func UniqueSlice[T any, K comparable](slice []T, keyFn func(T) K) []T {
	seen := make(map[K]bool)
	var result []T
	for _, item := range slice {
		key := keyFn(item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	return result
}

// TakeSlice returns the first n elements of a slice
func TakeSlice[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}
	if n >= len(slice) {
		return slice
	}
	return slice[:n]
}

// DropSlice returns a slice with the first n elements removed
func DropSlice[T any](slice []T, n int) []T {
	if n <= 0 {
		return slice
	}
	if n >= len(slice) {
		return []T{}
	}
	return slice[n:]
}

// ChunkSlice splits a slice into chunks of the specified size
func ChunkSlice[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}

	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// ZipSlice combines two slices into a slice of pairs
func ZipSlice[T, U any](slice1 []T, slice2 []U) []struct {
	First  T
	Second U
} {
	minLen := len(slice1)
	if len(slice2) < minLen {
		minLen = len(slice2)
	}

	result := make([]struct {
		First  T
		Second U
	}, minLen)
	for i := 0; i < minLen; i++ {
		result[i] = struct {
			First  T
			Second U
		}{slice1[i], slice2[i]}
	}
	return result
}

// Identity returns the input value unchanged (useful for function composition)
func Identity[T any](x T) T {
	return x
}

// Constant returns a function that always returns the same value
func Constant[T, U any](value T) func(U) T {
	return func(U) T {
		return value
	}
}
