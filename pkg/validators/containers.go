package validators

import (
	"strconv"
	"strings"
)

// Container helpers used by generated code. They mirror the semantics of
// go-playground/validator's isOneOf/isUnique but take concrete types.

// IsOneOf mirrors isOneOf for string values.
func IsOneOf(v string, vals []string) bool {
	for _, val := range vals {
		if val == v {
			return true
		}
	}
	return false
}

// IsOneOfFold mirrors isOneOfCI.
func IsOneOfFold(v string, vals []string) bool {
	for _, val := range vals {
		if strings.EqualFold(val, v) {
			return true
		}
	}
	return false
}

// Unique mirrors isUnique on slices and arrays of comparable values.
func Unique[T comparable](in []T) bool {
	seen := make(map[T]struct{}, len(in))
	for _, v := range in {
		seen[v] = struct{}{}
	}
	return len(in) == len(seen)
}

// UniquePtr mirrors isUnique for slices of pointers: values are compared after
// dereferencing, nil pointers all collapse onto the same (zero) entry.
func UniquePtr[T comparable](in []*T) bool {
	seen := make(map[T]struct{}, len(in))
	for _, v := range in {
		var zero T
		if v != nil {
			zero = *v
		}
		seen[zero] = struct{}{}
	}
	return len(in) == len(seen)
}

// UniqueBy deduplicates by a key extracted from each element; elements whose
// key function returns false are ignored, mirroring validator's invalid-key
// handling.
func UniqueBy[T any, K comparable](in []T, key func(T) (K, bool)) bool {
	seen := make(map[K]struct{}, len(in))
	count := 0
	for _, v := range in {
		k, ok := key(v)
		if !ok {
			continue
		}
		count++
		seen[k] = struct{}{}
	}
	return count == len(seen)
}

// UniqueMap mirrors isUnique on maps: the values must be unique.
func UniqueMap[K comparable, V comparable](in map[K]V) bool {
	seen := make(map[V]struct{}, len(in))
	for _, v := range in {
		seen[v] = struct{}{}
	}
	return len(in) == len(seen)
}

// UniqueMapPtr mirrors isUnique on maps with pointer values.
func UniqueMapPtr[K comparable, V comparable](in map[K]*V) bool {
	seen := make(map[V]struct{}, len(in))
	for _, v := range in {
		var zero V
		if v != nil {
			zero = *v
		}
		seen[zero] = struct{}{}
	}
	return len(in) == len(seen)
}

// IsBoolean mirrors isBoolean for string fields: the value must parse as a
// boolean ("1", "t", "TRUE", "false", ... - anything strconv.ParseBool accepts).
func IsBoolean(s string) bool {
	_, err := strconv.ParseBool(s)
	return err == nil
}
