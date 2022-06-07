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

	"go.uber.org/multierr"
)

var (
	ErrInvalidTypeValidator = errors.New("invalid TypeValidator")
	ErrInvalidValidator     = errors.New("invalid WRP message type validator")
	ErrInvalidMsgType       = errors.New("invalid WRP message type")
)

// AlwaysInvalid doesn't validate anything about the message and always returns an error.
var AlwaysInvalid ValidatorFunc = func(m Message) error { return ErrInvalidMsgType }

// AlwaysValid doesn't validate anything about the message and always returns nil.
var AlwaysValid ValidatorFunc = func(msg Message) error { return nil }

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
	var err error
	for _, v := range vs {
		if v != nil {
			err = multierr.Append(err, v.Validate(m))
		}
	}

	return err
}

// ValidatorFunc is a WRP validator that takes messages and validates them
// against functions.
type ValidatorFunc func(Message) error

// Validate executes its own ValidatorFunc receiver and returns the result.
func (vf ValidatorFunc) Validate(m Message) error {
	return vf(m)
}

// TypeValidator is a WRP validator that validates based on message type
// or using the defaultValidators if message type is unfound.
type TypeValidator struct {
	m                 map[MessageType]Validator
	defaultValidators Validator
	isbad             bool
}

// Validate validates messages based on message type or using the defaultValidators
// if message type is unfound.
func (m TypeValidator) Validate(msg Message) error {
	if m.isbad {
		return ErrInvalidTypeValidator
	}

	vs := m.m[msg.MessageType()]
	if vs == nil {
		return m.defaultValidators.Validate(msg)
	}

	return vs.Validate(msg)
}

// IsBad returns a boolean indicating whether the TypeValidator receiver is valid
func (m TypeValidator) IsBad() bool {
	return m.isbad
}

// NewTypeValidator is a TypeValidator factory.
func NewTypeValidator(m map[MessageType]Validator, defaultValidators ...Validator) (TypeValidator, error) {
	if m == nil {
		return TypeValidator{isbad: true}, ErrInvalidValidator
	}

	if defaultValidators == nil {
		defaultValidators = Validators{AlwaysInvalid}
	}

	return TypeValidator{
		m:                 m,
		defaultValidators: Validators(defaultValidators),
	}, nil
}
