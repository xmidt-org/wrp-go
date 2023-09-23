// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"errors"
	"fmt"
	"reflect"
	"unicode/utf8"
)

var (
	ErrNotUTF8        = errors.New("field contains non-utf-8 characters")
	ErrUnexpectedKind = errors.New("a struct or non-nil pointer to struct is required")
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
		}
	}

	return nil
}
