// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone"
	"github.com/xmidt-org/wrp-go/v3"
	"go.uber.org/multierr"
)

var (
	ErrorNotSimpleResponseRequestType = NewValidatorError(errors.New("not simple response request message type"), "", []string{"Type"})
	ErrorNotSimpleEventType           = NewValidatorError(errors.New("not simple event message type"), "", []string{"Type"})
	ErrorInvalidSpanLength            = NewValidatorError(errors.New("invalid span length"), "", []string{"Spans"})
	ErrorInvalidSpanFormat            = NewValidatorError(errors.New("invalid span format"), "", []string{"Spans"})
)

// spanFormat is a simple map of allowed span format.
var spanFormat = map[int]string{
	// parent is the root parent for the spans below to link to
	0: "parent",
	// name is the name of the operation
	1: "name",
	// start time of the operation.
	2: "start time",
	// duration is how long the operation took.
	3: "duration",
	// status of the operation
	4: "status",
}

// SimpleEvent ensures messages are valid based on
// each validator in the list. SimpleEvent validates the following:
// UTF8 (all string fields), MessageType is valid, Source, Destination, MessageType is of SimpleEventMessageType.
func SimpleEvent(tf *touchstone.Factory, labelNames ...string) (Validators, error) {
	var errs error
	sv, err := SpecWithMetrics(tf, labelNames...)
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	stv, err := NewSimpleEventTypeWithMetric(tf, labelNames...)
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	return sv.AddFunc(stv), errs
}

// SimpleResponseRequest ensures messages are valid based on
// each validator in the list. SimpleResponseRequest validates the following:
// UTF8 (all string fields), MessageType is valid, Source, Destination, Spans, MessageType is of
// SimpleRequestResponseMessageType.
func SimpleResponseRequest(tf *touchstone.Factory, labelNames ...string) (Validators, error) {
	var errs error
	sv, err := SpecWithMetrics(tf, labelNames...)
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	stv, err := NewSimpleResponseRequestTypeWithMetric(tf, labelNames...)
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	spv, err := NewSpansWithMetric(tf, labelNames...)
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	return sv.AddFunc(stv, spv), errs
}

// NewSimpleResponseRequestTypeWithMetric returns a SimpleResponseRequestType validator with a metric middleware.
func NewSimpleResponseRequestTypeWithMetric(tf *touchstone.Factory, labelNames ...string) (ValidatorFunc, error) {
	m, err := newSimpleRequestResponseMessageTypeErrorTotal(tf, labelNames...)

	return func(msg wrp.Message, ls prometheus.Labels) error {
		err := SimpleResponseRequestType(msg)
		if err != nil {
			m.With(ls).Add(1.0)
		}

		return err
	}, err
}

// NewSimpleEventTypeWithMetric returns a SimpleEventType validator with a metric middleware.
func NewSimpleEventTypeWithMetric(tf *touchstone.Factory, labelNames ...string) (ValidatorFunc, error) {
	m, err := newSimpleEventTypeErrorTotal(tf, labelNames...)

	return func(msg wrp.Message, ls prometheus.Labels) error {
		err := SimpleEventType(msg)
		if err != nil {
			m.With(ls).Add(1.0)
		}

		return err
	}, err
}

// NewSpansWithMetric returns a Spans validator with a metric middleware.
func NewSpansWithMetric(tf *touchstone.Factory, labelNames ...string) (ValidatorFunc, error) {
	m, err := newSpansErrorTotal(tf, labelNames...)

	return func(msg wrp.Message, ls prometheus.Labels) error {
		err := Spans(msg)
		if err != nil {
			m.With(ls).Add(1.0)
		}

		return err
	}, err
}

// SimpleResponseRequestType takes messages and validates their Type is of SimpleRequestResponseMessageType.
func SimpleResponseRequestType(m wrp.Message) error {
	if m.Type != wrp.SimpleRequestResponseMessageType {
		return ErrorNotSimpleResponseRequestType
	}

	return nil
}

// SimpleEventType takes messages and validates their Type is of SimpleEventMessageType.
func SimpleEventType(m wrp.Message) error {
	if m.Type != wrp.SimpleEventMessageType {
		return ErrorNotSimpleEventType
	}

	return nil
}

// TODO Do we want to include SpanParentValidator? SpanParent currently doesn't exist in the Message Struct

// Spans takes messages and validates their Spans.
func Spans(m wrp.Message) error {
	var err error
	// Spans consist of individual Span(s), arrays of timing values.
	for _, s := range m.Spans {
		if len(s) != len(spanFormat) {
			err = multierr.Append(err, ErrorInvalidSpanLength)
			continue
		}

		for i, j := range spanFormat {
			switch j {
			// Any nonempty string is valid
			case "parent", "name":
				if len(s[i]) == 0 {
					err = multierr.Append(err, fmt.Errorf("%w %v: invalid %v component '%v'", ErrorInvalidSpanFormat, s, j, s[i]))
				}
			// Must be an integer
			case "start time", "duration", "status":
				if _, atoiErr := strconv.Atoi(s[i]); atoiErr != nil {
					err = multierr.Append(err, fmt.Errorf("%w %v: invalid %v component '%v': %v", ErrorInvalidSpanFormat, s, j, s[i], atoiErr))
				}
			}
		}
	}

	return err
}
