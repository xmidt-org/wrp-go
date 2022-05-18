/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
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

package wrp

import (
	"errors"
	"fmt"
	"reflect"
	"unicode/utf8"
)

var (
	ErrNotUTF8        = errors.New("field contains non-utf-8 characters")
	ErrUnexpectedKind = errors.New("A struct or non-nil pointer to struct is required")
)

// UTF8 takes any struct verifies that it contains UTF-8 strings.
func UTF8(v interface{}) error {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr && !value.IsNil() {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return fmt.Errorf("%w: %s", ErrUnexpectedKind, value.Kind())
	}

	for i := 0; i < value.NumField(); i++ {
		ft := value.Type().Field(i)
		if len(ft.PkgPath) > 0 || ft.Anonymous {
			continue // skip embedded or unexported fields
		}

		f := value.Field(i)
		if !f.CanInterface() {
			continue // this should never happen, but ... you never know
		}

		if s, ok := f.Interface().(string); ok {
			if !utf8.ValidString(s) {
				return fmt.Errorf("%w: '%s:%v'", ErrNotUTF8, ft.Name, s)
			}
			fmt.Println(s)
		}
	}

	return nil
}
