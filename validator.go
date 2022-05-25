/**
 *  Copyright (c) 2022  Comcast Cable Communications Management, LLC
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
)

var (
	ErrInvalidMsgTypeValidator = errors.New("invalid WRP message type validator")
	ErrInvalidMsgType          = errors.New("invalid WRP message type")
)

// Validator is a WRP validator that allows access to the Validate function.
type Validator interface {
	Validate(m Message) error
}

// Validators is a WRP validator that ensures messages are valid based on
// message type and each validator in the list.
type Validators []Validator

// ValidatorFunc is a WRP validator that takes messages and validates them
// against functions.
type ValidatorFunc func(Message) error

// Validate executes its own ValidatorFunc receiver and returns the result.
func (vf ValidatorFunc) Validate(m Message) error {
	return vf(m)
}

// MsgTypeValidator is a WRP validator that validates based on message type
// or using the defaultValidator if message type is unknown
type MsgTypeValidator struct {
	m                map[MessageType]Validators
	defaultValidator Validator
}

// Validate validates messages based on message type or using the defaultValidator
// if message type is unknown
func (m MsgTypeValidator) Validate(msg Message) error {
	vs, ok := m.m[msg.MessageType()]
	if !ok {
		return m.defaultValidator.Validate(msg)
	}

	for _, v := range vs {
		err := v.Validate(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewMsgTypeValidator returns a MsgTypeValidator
func NewMsgTypeValidator(m map[MessageType]Validators, defaultValidator Validator) (MsgTypeValidator, error) {
	if m == nil {
		return MsgTypeValidator{}, fmt.Errorf("%w: %v", ErrInvalidMsgTypeValidator, m)
	}

	if defaultValidator == nil {
		defaultValidator = alwaysInvalidMsg()
	}

	return MsgTypeValidator{
		m:                m,
		defaultValidator: defaultValidator,
	}, nil
}

// AlwaysInvalid doesn't validate anything about the message and always returns an error.
func alwaysInvalidMsg() ValidatorFunc {
	return func(m Message) error {
		return fmt.Errorf("%w: %v", ErrInvalidMsgType, m.MessageType().String())
	}
}
