// Mirrors github.com/go-playground/validator/v10 (util.go, extractTypeInternal).
// Copyright (c) 2015 Dean Karn. MIT License — see LICENSE.

package validators

import "reflect"

// DynamicKind reports the kind the reference would see after unwrapping every
// pointer and interface that is not nil.
//
// It exists for one case: an interface field holds a value whose type is only
// known at run time, so the Kind of a validation error on such a field cannot be
// written out at generation time. Generated code calls it while building the
// error, never while checking a value, so it costs nothing on the passing path.
func DynamicKind(v interface{}) reflect.Kind {
	rv := reflect.ValueOf(v)
	for {
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface:
			if rv.IsNil() {
				return rv.Kind()
			}
			rv = rv.Elem()
		default:
			return rv.Kind()
		}
	}
}
