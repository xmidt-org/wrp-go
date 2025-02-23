// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"context"
	"errors"
	"reflect"
)

// Validate performs a set of validations on a message.  If no validators are
// provided, the default set of standard WRP validators is used.  If the
// NoStandardValidation() processor is provided, no standard validation is
// performed.  After standard validation (if applicable) is performed, any
// additional validators are executed in the order they are provided.  If any
// validator returns an error excluding ErrNotHandled, the iteration stops and
// the error is returned.  If a validator return ErrNotHandled, then the
// validation is considered successful.  Any combination of nil errors and
// ErrNotHandled is considered a successful validation.  All other errors are
// considered validation failures and the first encountered error is returned.
func Validate(msg Union, validators ...Processor) error {
	if msg == nil || reflect.ValueOf(msg).IsNil() {
		return ErrMessageIsInvalid
	}

	if m, ok := msg.(*Message); ok {
		return validate(m, validators...)
	}

	return msg.To(&Message{}, validators...)
}

// SkipStandardValidation returns true if the provided processors contain
// the NoStandardValidation() provided processor.
func SkipStandardValidation(p []Processor) bool {
	for _, v := range p {
		if _, ok := v.(noStandardValidation); ok {
			return true
		}
	}
	return false
}

func validate(msg *Message, validators ...Processor) error {
	defaults := []Processor{
		StandardValidator(),
	}
	if SkipStandardValidation(validators) {
		defaults = nil
	}

	validators = append(defaults, validators...)

	err := Processors(validators).ProcessWRP(context.Background(), *msg)
	if err == nil || errors.Is(err, ErrNotHandled) {
		return nil
	}

	return err
}
