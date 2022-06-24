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
	"strconv"

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

// SimpleEventValidators ensures messages are valid based on
// each validator in the list. SimpleEventValidators validates the following:
// UTF8 (all string fields), MessageType is valid, Source, Destination, MessageType is of SimpleEventMessageType.
func SimpleEventValidators() Validators {
	return Validators{SpecValidators()}.AddFunc(SimpleEventTypeValidator)
}

// SimpleResponseRequestValidators ensures messages are valid based on
// each validator in the list. SimpleResponseRequestValidators validates the following:
// UTF8 (all string fields), MessageType is valid, Source, Destination, Spans, MessageType is of
// SimpleRequestResponseMessageType.
func SimpleResponseRequestValidators() Validators {
	return Validators{SpecValidators()}.AddFunc(SimpleResponseRequestTypeValidator, SpansValidator)
}

// SimpleResponseRequestTypeValidator takes messages and validates their Type is of SimpleRequestResponseMessageType.
func SimpleResponseRequestTypeValidator(m Message) error {
	if m.Type != SimpleRequestResponseMessageType {
		return ErrorNotSimpleResponseRequestType
	}

	return nil
}

// TODO Do we want to include SpanParentValidator? SpanParent currently doesn't exist in the Message Struct

// SpansValidator takes messages and validates their Spans.
func SpansValidator(m Message) error {
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

// SimpleEventTypeValidator takes messages and validates their Type is of SimpleEventMessageType.
func SimpleEventTypeValidator(m Message) error {
	if m.Type != SimpleEventMessageType {
		return ErrorNotSimpleEventType
	}

	return nil
}
