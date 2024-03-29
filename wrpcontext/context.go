// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpcontext

import (
	"context"
	"reflect"
)

// Get a message from a context and return it as type T
func get[T any](ctx context.Context, src any) (dest T, ok bool) {
	// if src and dest are the exact same type
	if dest, ok = src.(T); ok {
		return
	}

	// if src is a pointer to the same type as the value of dest
	var srcptr *T
	if srcptr, ok = src.(*T); ok {
		if srcptr == nil {
			ok = false
		} else {
			dest = *srcptr
		}
		return
	}

	// if src is a value, and dest is a pointer to the same type as src
	srcValue := reflect.ValueOf(src)
	destType := reflect.TypeOf((*T)(nil)).Elem()
	ok = (srcValue.Kind() != reflect.Ptr) && (destType.Kind() == reflect.Ptr) && srcValue.Type().ConvertibleTo(destType.Elem())

	if ok {
		// use reflect to create a pointer to T's element type, which will allocate memory
		// then we copy src to that memory
		destValue := reflect.New(destType.Elem())
		destValue.Elem().Set(srcValue)
		dest = destValue.Interface().(T)
	}

	return
}
