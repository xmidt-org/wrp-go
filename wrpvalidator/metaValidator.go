// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone"
	"github.com/xmidt-org/wrp-go/v3"
)

type Metadata struct {
	Level   validatorLevel `json:"level"`
	Type    validatorType  `json:"type"`
	Disable bool           `json:"disable"`
}

type ValidatorWithMetadata struct {
	meta      Metadata
	validator Validator
}

var (
	ErrValidatorUnmarshalling = errors.New("unmarshalling  error")
	ErrValidatorAddMetric     = errors.New("add metric middleware error")
	ErrValidatorInvalidConfig = errors.New("invalid configuration error")
)

func (v ValidatorWithMetadata) Level() validatorLevel {
	return v.meta.Level
}

func (v ValidatorWithMetadata) Type() validatorType {
	return v.meta.Type
}

func (v ValidatorWithMetadata) Disabled() bool {
	return v.meta.Disable
}

// UnmarshalJSON returns the ValidatorConfig's enum value
func (v *ValidatorWithMetadata) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	if err := json.Unmarshal(b, &v.meta); err != nil {
		return fmt.Errorf("json unmarshal: %w: %s", ErrValidatorUnmarshalling, err)
	}

	var val func(wrp.Message) error
	switch v.meta.Type {
	case AlwaysInvalidType:
		val = AlwaysInvalid
	case AlwaysValidType:
		val = AlwaysValid
	case UTF8Type:
		val = UTF8
	case MessageTypeType:
		val = MessageType
	case SourceType:
		val = Source
	case DestinationType:
		val = Destination
	case SimpleResponseRequestTypeType:
		val = SimpleResponseRequestType
	case SimpleEventTypeType:
		val = SimpleEventType
	case SpansType:
		val = Spans
	default:
		return fmt.Errorf("validator `%s`: wrp validator selection: %w: %s", v.meta.Type, ErrValidatorUnmarshalling, errValidatorTypeInvalid)
	}

	v.validator = NewValidatorWithoutMetric(val)

	if !v.IsValid() {
		return fmt.Errorf("validator `%s`: invalid configuration: %w", v.meta.Type, ErrValidatorInvalidConfig)
	}

	return nil
}

func (v *ValidatorWithMetadata) AddMetric(tf *touchstone.Factory, labelNames ...string) error {
	if !v.IsValid() {
		return fmt.Errorf("validator `%s`: invalid configuration: %w", v.meta.Type, ErrValidatorInvalidConfig)
	} else if v.meta.Disable {
		return nil
	}

	var (
		err error
		val Validator
	)
	switch v.meta.Type {
	case AlwaysInvalidType:
		val, err = NewAlwaysInvalidWithMetric(tf, labelNames...)
	case AlwaysValidType:
		val, err = NewAlwaysValidWithMetric(tf, labelNames...)
	case UTF8Type:
		val, err = NewUTF8WithMetric(tf, labelNames...)
	case MessageTypeType:
		val, err = NewMessageTypeWithMetric(tf, labelNames...)
	case SourceType:
		val, err = NewSourceWithMetric(tf, labelNames...)
	case DestinationType:
		val, err = NewDestinationWithMetric(tf, labelNames...)
	case SimpleResponseRequestTypeType:
		val, err = NewSimpleResponseRequestTypeWithMetric(tf, labelNames...)
	case SimpleEventTypeType:
		val, err = NewSimpleEventTypeWithMetric(tf, labelNames...)
	case SpansType:
		val, err = NewSpansWithMetric(tf, labelNames...)
		// no default is needed since v.IsValid() takes care of this case
	}

	if err != nil {
		return fmt.Errorf("validator `%s`: wrp validator middleware modifier: %w: %s", v.meta.Type, ErrValidatorAddMetric, err)
	}

	v.validator = val

	return nil
}

// Validate executes its own ValidatorFunc receiver and returns the result.
func (v ValidatorWithMetadata) Validate(m wrp.Message, ls prometheus.Labels) error {
	if !v.IsValid() {
		return fmt.Errorf("validator `%s`: invalid configuration: %w", v.meta.Type, ErrValidatorInvalidConfig)
	} else if v.meta.Disable {
		return nil
	} else if err := v.validator.Validate(m, ls); err != nil {
		return fmt.Errorf("validator `%s`: %w", v.meta.Type, err)
	}

	return nil
}

func (v ValidatorWithMetadata) IsValid() bool {
	return v.meta.Type.IsValid() && v.meta.Level.IsValid() && v.validator != nil
}

func (v ValidatorWithMetadata) IsEmpty() bool {
	return v.meta.Type.IsEmpty() && v.meta.Level.IsEmpty() && v.validator == nil
}
