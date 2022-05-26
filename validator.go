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
)

var (
	ErrInvalidTypeValidator = errors.New("invalid WRP message type validator")
	ErrInvalidMsgType       = errors.New("invalid WRP message type")
)

// AlwaysInvalid doesn't validate anything about the message and always returns an error.
var AlwaysInvalid ValidatorFunc = func(m Message) error {
	return ErrInvalidMsgType
}

// Validator is a WRP validator that allows access to the Validate function.
type Validator interface {
	Validate(m Message) error
}

// Validators is a WRP validator that ensures messages are valid based on
// message type and each validator in the list.
type Validators []Validator

// Validate runs messages through each validator in the validators list.
// It returns as soon as the message is considered invalid, otherwise returns nil if valid.
func (vs Validators) Validate(m Message) error {
	for _, v := range vs {
		err := v.Validate(m)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidatorFunc is a WRP validator that takes messages and validates them
// against functions.
type ValidatorFunc func(Message) error

// Validate executes its own ValidatorFunc receiver and returns the result.
func (vf ValidatorFunc) Validate(m Message) error {
	return vf(m)
}

// TypeValidator is a WRP validator that validates based on message type
// or using the defaultValidator if message type is unknown.
type TypeValidator struct {
	m                map[MessageType]Validators
	defaultValidator Validator
}

// Validate validates messages based on message type or using the defaultValidator
// if message type is unknown.
func (m TypeValidator) Validate(msg Message) error {
	vs := m.m[msg.MessageType()]
	if vs == nil {
		return m.defaultValidator.Validate(msg)
	}

	return vs.Validate(msg)
}

// NewTypeValidator is a TypeValidator factory.
func NewTypeValidator(m map[MessageType]Validators, defaultValidator Validator) (TypeValidator, error) {
	if m == nil {
		return TypeValidator{}, ErrInvalidTypeValidator
	}

	for _, v := range m {
		if v == nil || len(v) == 0 {
			return TypeValidator{}, ErrInvalidTypeValidator
		}
	}

	if defaultValidator == nil {
		defaultValidator = AlwaysInvalid
	}

	return TypeValidator{
		m:                m,
		defaultValidator: defaultValidator,
	}, nil
}
