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
	"reflect"
	"strings"

	"go.uber.org/multierr"
)

var (
	ErrorInvalidValidator      = NewValidatorError(errors.New("invalid WRP message type validator"), nil, "")
	ErrorInvalidMsgType        = NewValidatorError(errors.New("invalid WRP message type"), []string{"Type"}, "")
	errorInvalidValidatorError = errors.New("invalid validator error")
)

type ValidatorError struct {
	// Fields are the relevant fields involved in Err.
	Fields []reflect.StructField

	// Err is the cause of the error, e.g. invalid message type.
	// Either Err or Message must be set.
	Err error

	// Message is a validation message in case the validator wants
	// to communicate something beyond the Err cause.
	Message string
}

// Is reports whether any error in e.err's chain matches target.
func (ve ValidatorError) Is(target error) bool {
	return errors.Is(ve.Err, target)
}

// Unwrap returns the ValidatorError's Error
func (ve ValidatorError) Unwrap() error {
	return ve.Err
}

func (ve ValidatorError) Error() string {
	var o strings.Builder
	o.WriteString("Validator error")

	if len(ve.Fields) > 0 {
		o.WriteString(" [")
		for i, f := range ve.Fields {
			if i > 0 {
				o.WriteRune(',')
			}

			o.WriteString(f.Name)
		}

		o.WriteRune(']')
	}

	if ve.Err != nil {
		o.WriteString(" err=")
		o.WriteString(ve.Err.Error())
	}

	if len(ve.Message) > 0 {
		o.WriteString(" msg=")
		o.WriteString(ve.Message)
	}

	return o.String()
}

func NewValidatorError(err error, sf []string, m string) ValidatorError {
	// Either Err or Message must be set.
	if err == nil && len(m) == 0 {
		return ValidatorError{Err: errorInvalidValidatorError}
	}

	var badFields string
	verr := ValidatorError{Err: err, Fields: make([]reflect.StructField, 0, len(sf)), Message: m}
	// Fields must exist.
	for _, f := range sf {
		ft, ok := messageReflectType.FieldByName(f)
		if !ok {
			badFields += fmt.Sprintf(",'%v'", f)
			continue
		}

		verr.Fields = append(verr.Fields, ft)
	}

	if len(badFields) != 0 {
		verr.Err = fmt.Errorf("%v: %w: following fields were not found %v", err, errorInvalidValidatorError, badFields)
	}

	return verr
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
	var err error
	for _, v := range vs {
		if v != nil {
			err = multierr.Append(err, v.Validate(m))
		}
	}

	return err
}

// Add returns a new Validators with the appended Validator list
func (vs Validators) Add(v ...Validator) Validators {
	for _, v := range v {
		if v != nil {
			vs = append(vs, v)
		}
	}

	return vs
}

// AddFunc returns a new Validators with the appended ValidatorFunc list
func (vs Validators) AddFunc(vf ...ValidatorFunc) Validators {
	for _, v := range vf {
		if v != nil {
			vs = append(vs, v)
		}
	}

	return vs
}

// ValidatorFunc is a WRP validator that takes messages and validates them
// against functions.
type ValidatorFunc func(Message) error

// Validate executes its own ValidatorFunc receiver and returns the result.
func (vf ValidatorFunc) Validate(m Message) error { return vf(m) }

// TypeValidator is a WRP validator that validates based on message type
// or using the defaultValidator if message type is unfound.
type TypeValidator struct {
	m                map[MessageType]Validator
	defaultValidator Validator
}

// Validate validates messages based on message type or using the defaultValidator
// if message type is unfound.
func (tv TypeValidator) Validate(msg Message) error {
	vs := tv.m[msg.MessageType()]
	if vs == nil {
		return tv.defaultValidator.Validate(msg)
	}

	return vs.Validate(msg)
}

// NewTypeValidator is a TypeValidator factory.
func NewTypeValidator(m map[MessageType]Validator, defaultValidator Validator) (TypeValidator, error) {
	if m == nil {
		return TypeValidator{}, ErrorInvalidValidator
	}

	if defaultValidator == nil {
		defaultValidator = ValidatorFunc(AlwaysInvalid)
	}

	return TypeValidator{
		m:                m,
		defaultValidator: defaultValidator,
	}, nil
}

// AlwaysInvalid doesn't validate anything about the message and always returns an error.
func AlwaysInvalid(_ Message) error {
	return ErrorInvalidMsgType
}

// AlwaysValid doesn't validate anything about the message and always returns nil.
func AlwaysValid(_ Message) error {
	return nil
}
