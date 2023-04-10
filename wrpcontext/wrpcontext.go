/**
 * Copyright 2023 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package wrpcontext

import (
	"context"
	"reflect"
)

type contextKey struct{}

// Set provides a standard way to add a wrp message to a context.Context. This supports not only wrp.Message
// but also all the other message types, such as wrp.SimpleRequestResponse
func Set(ctx context.Context, msg any) context.Context {
	return context.WithValue(ctx, contextKey{}, msg)
}

// Get a message from a context and return it as type T
func Get[T any](ctx context.Context) (dest T, ok bool) {
	src := ctx.Value(contextKey{})
	if src == nil {
		return
	}

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
