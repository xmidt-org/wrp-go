// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"context"
	"errors"
	"reflect"
)

func Is(msg *Message, target any, validators ...Processor) bool {
	if msg == nil || target == nil {
		// This is the simple way to handle nils, otherwise the nil check would
		// possibly say nil != nil because of the typeness of the nils.
		if msg == nil && target == nil {
			return true
		}
		return false
	}

	// Check if the target type is a struct or a pointer to a struct
	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}
	if targetType.Kind() != reflect.Struct {
		// We're done, the target is not a struct that can have a valid message
		// embedded in it.
		return false
	}

	exact := mtToStruct[msg.Type]
	if exact == nil {
		return false
	}

	var exactPtr any
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

func findTarget(target any, typ MessageType) (any, reflect.StructField, bool) {
	// Only structs are allowed to be targets.
	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}
	if targetType.Kind() != reflect.Struct {
		return nil, reflect.StructField{}, false
	}

	exact := mtToStruct[typ]
	if exact == nil {
		return nil, reflect.StructField{}, false
	}

	if reflect.TypeOf(target) == reflect.TypeOf(exact) {
		return target, reflect.StructField{}, true
	}

	return nil, reflect.StructField{}, false
}

// Message -> target
// target -> Message
// target -> Message -> msg (if not Message)
// methods:
//
//	To(*Message) error -- called on a target if present to allow converting to a message
//	From(*Message) error -- called on a target if present to allow converting from a message
func As(msg, target any, validators ...Processor) bool {
	if msg == nil || target == nil {
		return false
	}

	// Check if the target type is a struct or a pointer to a struct
	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}
	if targetType.Kind() != reflect.Struct {
		// We're done, the target is not a struct that can have a valid message
		// embedded in it.
		return false
	}

	exact := mtToStruct[msg.Type]
	if exact == nil {
		return false
	}

	var exactPtr any
	if m, ok := target.(*Message); ok {
		if err := Validate(msg, validators...); err != nil {
			return false
		}
		m.from(msg)
		return true
	}

	exactPtr = reflect.New(reflect.TypeOf(exact)).Interface()
	if reflect.TypeOf(target) == reflect.TypeOf(exactPtr) {
		goto probably
	}

	// At this point we want to see if the target is a pointer to a struct that
	// contains a Message/*Message field or a field of any of the other types in
	// MessageType enum that match what this message is targeting.

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
}
