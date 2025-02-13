// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"context"
	"errors"
	"reflect"
)

type MsgType interface {
	Type() MessageType
	to() *Message
}

func Is(msg *Message, target any, validators ...Processor) bool {
	if msg == nil || target == nil {
		// This is the simple way to handle nils, otherwise the nil check would
		// possibly say nil != nil because of the typeness of the nils.
		if msg == nil && target == nil {
			return true
		}
		return false
	}

	exact := mtToStruct[msg.Type]
	if exact == nil {
		return false
	}

	var exactPtr any
	var targetType reflect.Type
	if _, ok := target.(*Message); ok {
		goto probably
	}

	exactPtr = reflect.New(reflect.TypeOf(exact)).Interface()

	if reflect.TypeOf(target) == reflect.TypeOf(exactPtr) {
		goto probably
	}

	// At this point we want to see if the target is a pointer to a struct that
	// contains a Message/*Message field or a field of any of the other types in
	// MessageType enum that match what this message is targeting.

	// Check if the target type is a struct or a pointer to a struct
	targetType = reflect.TypeOf(target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}
	if targetType.Kind() != reflect.Struct {
		// We're done, the target is not a struct that can have a valid message
		// embedded in it.
		return false
	}

	// Iterate through the fields of the target struct to see if
	// any of them are a Message/*Message or the exact type or *exact type.
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		if field.Type.Kind() == reflect.Ptr {
			field.Type = field.Type.Elem()
		}
		if field.Type.Kind() == reflect.Struct {
			if Is(msg, reflect.New(field.Type).Interface(), validators...) {
				return true
			}
		}
	}

	return false

probably:
	if err := Processors(validators).ProcessWRP(context.Background(), *msg); err != nil {
		if !errors.Is(err, ErrNotHandled) {
			return false
		}
	}

	return true
}

func As(msg *Message, target any, validators ...Processor) bool {
	return false
}
