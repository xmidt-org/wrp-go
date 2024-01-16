// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"errors"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/wrp-go/v3"
	"go.uber.org/multierr"
)

var (
	ErrorInvalidValidator      = NewValidatorError(errors.New("invalid WRP message type validator"), "", nil)
	ErrorInvalidMsgType        = NewValidatorError(errors.New("invalid WRP message type"), "", []string{"Type"})
	ErrorInvalidValidatorError = errors.New("empty ValidatorError 'Err' and 'Message'")
)

type ValidatorError struct {
	// Err is the cause of the error, e.g. invalid message type.
	// Either Err or Message must be set and nonempty
	Err error

	// Message is a validation message in case the validator wants
	// to communicate something beyond the Err cause.
	Message string

	// Fields are the relevant fields involved in Err.
	Fields []string
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

			o.WriteString(f)
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

// NewValidatorError is a ValidatorError factory and will panic if
// both 'err' and 'm' are empty or nil
func NewValidatorError(err error, m string, f []string) ValidatorError {
	if (err == nil || len(err.Error()) == 0) && len(m) == 0 {
		panic(ErrorInvalidValidatorError)
	}

	return ValidatorError{err, m, f}
}

// Validator is a WRP validator that allows access to the Validate function.
type Validator interface {
	Validate(wrp.Message, prometheus.Labels) error
}

// Validators is a WRP validator that ensures messages are valid based on
// message type and each validator in the list.
type Validators []Validator

// Validate runs messages through each validator in the validators list.
// It returns as soon as the message is considered invalid, otherwise returns nil if valid.
func (vs Validators) Validate(m wrp.Message, ls prometheus.Labels) error {
	var err error
	for _, v := range vs {
		if v != nil {
			err = multierr.Append(err, v.ValidateWithMetrics(m, ls))
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
type ValidatorFunc func(wrp.Message, prometheus.Labels) error

// Validate executes its own ValidatorFunc receiver and returns the result.
func (vf ValidatorFunc) Validate(m wrp.Message, ls prometheus.Labels) error { return vf(m, ls) }

// TypeValidator is a WRP validator that validates based on message type
// or using the defaultValidator if message type is unfound.
type TypeValidator struct {
	m                map[MessageType]Validator
	defaultValidator Validator
}

// Validate validates messages based on message type or using the defaultValidator
// if message type is unfound.
func (tv TypeValidator) Validate(m wrp.Message, ls prometheus.Labels) error {
	vs := tv.m[m.MessageType()]
	if vs == nil {
		return tv.defaultValidator.Validate(m, ls)
	}

	return vs.Validate(m, ls)
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
