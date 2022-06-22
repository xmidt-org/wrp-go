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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleMessageTypesValidatorErrors(t *testing.T) {
	tests := []struct {
		description  string
		validatorErr ValidatorError
	}{
		// Success case
		{
			description:  "ErrorNotSimpleResponseRequestType",
			validatorErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description:  "ErrorNotSimpleEventType",
			validatorErr: ErrorNotSimpleEventType,
		},
		{
			description:  "ErrorInvalidSpanLength",
			validatorErr: ErrorInvalidSpanLength,
		},
		{
			description:  "ErrorInvalidSpanFormat",
			validatorErr: ErrorInvalidSpanFormat,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			assert.NotErrorIs(tc.validatorErr, errorInvalidValidatorError)
		})
	}
}

func TestSimpleMessageTypesHelperValidators(t *testing.T) {
	tests := []struct {
		description string
		test        func(*testing.T)
	}{
		{"SpansValidator", testSpansValidator},
		{"SimpleResponseRequestTypeValidator", testSimpleResponseRequestTypeValidator},
		{"SimpleEventTypeValidator", testSimpleEventTypeValidator},
	}

	for _, tc := range tests {
		t.Run(tc.description, tc.test)
	}
}

func TestSimpleEventValidators(t *testing.T) {
	var (
		expectedStatus                  int64 = 3471
		expectedRequestDeliveryResponse int64 = 34
		expectedIncludeSpans            bool  = true
	)

	tests := []struct {
		description string
		msg         Message
		expectedErr []error
	}{
		// Success case
		{
			description: "Valid simple event message success",
			msg: Message{
				Type:                    SimpleEventMessageType,
				Source:                  "dns:external.com",
				Destination:             "MAC:11:22:33:44:55:66",
				TransactionUUID:         "DEADBEEF",
				ContentType:             "ContentType",
				Accept:                  "Accept",
				Status:                  &expectedStatus,
				RequestDeliveryResponse: &expectedRequestDeliveryResponse,
				Headers:                 []string{"Header1", "Header2"},
				Metadata:                map[string]string{"name": "value"},
				Spans:                   [][]string{{"1", "2"}, {"3"}},
				IncludeSpans:            &expectedIncludeSpans,
				Path:                    "/some/where/over/the/rainbow",
				Payload:                 []byte{1, 2, 3, 4, 0xff, 0xce},
				ServiceName:             "ServiceName",
				URL:                     "someURL.com",
				PartnerIDs:              []string{"foo"},
				SessionID:               "sessionID123",
			},
		},
		// Failure case
		{
			description: "Invaild simple event message error",
			msg: Message{
				Type: Invalid0MessageType,
				// Missing scheme
				Source: "external.com",
				// Invalid Mac
				Destination:     "MAC:+++BB-44-55",
				TransactionUUID: "DEADBEEF",
				ContentType:     "ContentType",
				Headers:         []string{"Header1", "Header2"},
				Metadata:        map[string]string{"name": "value"},
				Path:            "/some/where/over/the/rainbow",
				Payload:         []byte{1, 2, 3, 4, 0xff, 0xce},
				ServiceName:     "ServiceName",
				// Not UFT8 URL string
				URL:        "someURL\xed\xbf\xbf.com",
				PartnerIDs: []string{"foo"},
				SessionID:  "sessionID123",
			},
			expectedErr: []error{ErrorInvalidMessageType, ErrorInvalidSource, ErrorInvalidDestination, ErrorInvalidMessageEncoding, ErrorNotSimpleEventType},
		},
		{
			description: "Invaild simple event message error, empty message",
			msg:         Message{},
			expectedErr: []error{ErrorInvalidMessageType, ErrorInvalidSource, ErrorInvalidDestination, ErrorNotSimpleEventType},
		},
		{
			description: "Invaild simple event message error, non SimpleEventMessageType",
			msg: Message{
				Type:        CreateMessageType,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
			expectedErr: []error{ErrorNotSimpleEventType},
		},
		{
			description: "Invaild simple event message error, nonexistent MessageType",
			msg: Message{
				Type:        lastMessageType + 1,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
			expectedErr: []error{ErrorInvalidMessageType, ErrorNotSimpleEventType},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := SimpleEventValidators().Validate(tc.msg)
			if tc.expectedErr != nil {
				for _, e := range tc.expectedErr {
					if ve, ok := e.(ValidatorError); ok {
						e = ve.Err
					}

					assert.ErrorIs(err, e)
				}

				return
			}

			assert.NoError(err)
		})
	}
}

func TestSimpleResponseRequestValidators(t *testing.T) {
	var (
		expectedStatus                  int64 = 3471
		expectedRequestDeliveryResponse int64 = 34
		expectedIncludeSpans            bool  = true
	)

	tests := []struct {
		description string
		msg         Message
		expectedErr []error
	}{
		// Success case
		{
			description: "Valid simple request response message success",
			msg: Message{
				Type:                    SimpleRequestResponseMessageType,
				Source:                  "dns:external.com",
				Destination:             "MAC:11:22:33:44:55:66",
				TransactionUUID:         "DEADBEEF",
				ContentType:             "ContentType",
				Accept:                  "Accept",
				Status:                  &expectedStatus,
				RequestDeliveryResponse: &expectedRequestDeliveryResponse,
				Headers:                 []string{"Header1", "Header2"},
				Metadata:                map[string]string{"name": "value"},
				Spans:                   [][]string{{"parent", "name", "1234", "1234", "1234"}},
				IncludeSpans:            &expectedIncludeSpans,
				Path:                    "/some/where/over/the/rainbow",
				Payload:                 []byte{1, 2, 3, 4, 0xff, 0xce},
				ServiceName:             "ServiceName",
				URL:                     "someURL.com",
				PartnerIDs:              []string{"foo"},
				SessionID:               "sessionID123",
			},
		},
		// Failure case
		{
			description: "Invaild simple request response message error",
			msg: Message{
				Type: Invalid0MessageType,
				// Missing scheme
				Source: "external.com",
				// Invalid Mac
				Destination:             "MAC:+++BB-44-55",
				TransactionUUID:         "DEADBEEF",
				ContentType:             "ContentType",
				Accept:                  "Accept",
				Status:                  &expectedStatus,
				RequestDeliveryResponse: &expectedRequestDeliveryResponse,
				Headers:                 []string{"Header1", "Header2"},
				Metadata:                map[string]string{"name": "value"},
				Spans: [][]string{
					// // Invalid length
					{},
					// Invalid length
					{"3"},
					// Invalid 'start time', 'duration' and 'status' components
					{"parent", "name", "not start time", "not duration", "not status"},
					// Invalid 'parent' and 'name' components
					{"1234", "1234", "1234", "1234", "1234"},
				},
				IncludeSpans: &expectedIncludeSpans,
				Path:         "/some/where/over/the/rainbow",
				Payload:      []byte{1, 2, 3, 4, 0xff, 0xce},
				ServiceName:  "ServiceName",
				// Not UFT8 URL string
				URL:        "someURL\xed\xbf\xbf.com",
				PartnerIDs: []string{"foo"},
				SessionID:  "sessionID123",
			},
			expectedErr: []error{ErrorInvalidMessageType, ErrorInvalidSource, ErrorInvalidDestination, ErrorInvalidMessageEncoding, ErrorNotSimpleResponseRequestType, ErrorInvalidSpanLength, ErrorInvalidSpanFormat},
		},
		{
			description: "Invaild simple request response message error, empty message",
			msg:         Message{},
			expectedErr: []error{ErrorInvalidMessageType, ErrorInvalidSource, ErrorInvalidDestination, ErrorNotSimpleResponseRequestType},
		},
		{
			description: "Invaild simple request response message error, non SimpleEventMessageType",
			msg: Message{
				Type:        CreateMessageType,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
			expectedErr: []error{ErrorNotSimpleResponseRequestType},
		},
		{
			description: "Invaild simple request response message error, nonexistent MessageType",
			msg: Message{
				Type:        lastMessageType + 1,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
			expectedErr: []error{ErrorInvalidMessageType, ErrorNotSimpleResponseRequestType},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := SimpleResponseRequestValidators().Validate(tc.msg)
			if tc.expectedErr != nil {
				for _, e := range tc.expectedErr {
					if ve, ok := e.(ValidatorError); ok {
						e = ve.Err
					}

					assert.ErrorIs(err, e)
				}

				return
			}

			assert.NoError(err)
		})
	}
}

func ExampleTypeValidator_Validate_simpleTypesValidators() {
	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[MessageType]Validator{
			SimpleEventMessageType:           SimpleEventValidators(),
			SimpleRequestResponseMessageType: SimpleResponseRequestValidators(),
		},
		// Validates unfound msg types
		ValidatorFunc(AlwaysInvalid))
	if err != nil {
		return
	}

	var (
		expectedStatus                  int64 = 3471
		expectedRequestDeliveryResponse int64 = 34
		expectedIncludeSpans            bool  = true
	)
	foundErrFailure := msgv.Validate(Message{
		Type: SimpleRequestResponseMessageType,
		// Missing scheme
		Source: "external.com",
		// Invalid Mac
		Destination:             "MAC:+++BB-44-55",
		TransactionUUID:         "DEADBEEF",
		ContentType:             "ContentType",
		Accept:                  "Accept",
		Status:                  &expectedStatus,
		RequestDeliveryResponse: &expectedRequestDeliveryResponse,
		Headers:                 []string{"Header1", "Header2"},
		Metadata:                map[string]string{"name": "value"},
		Spans: [][]string{
			// // Invalid length
			{},
			// Invalid length
			{"3"},
			// Invalid 'start time', 'duration' and 'status' components
			{"parent", "name", "not start time", "not duration", "not status"},
			// Invalid 'parent' and 'name' components
			{"1234", "1234", "1234", "1234", "1234"},
		},
		IncludeSpans: &expectedIncludeSpans,
		Path:         "/some/where/over/the/rainbow",
		Payload:      []byte{1, 2, 3, 4, 0xff, 0xce},
		ServiceName:  "ServiceName",
		// Not UFT8 URL string
		URL:        "someURL\xed\xbf\xbf.com",
		PartnerIDs: []string{"foo"},
		SessionID:  "sessionID123",
	}) // Found error
	foundErrSuccess1 := msgv.Validate(Message{
		Type:        SimpleRequestResponseMessageType,
		Source:      "MAC:11:22:33:44:55:66",
		Destination: "MAC:11:22:33:44:55:61",
	}) // Found success
	foundErrSuccess2 := msgv.Validate(Message{
		Type:   SimpleEventMessageType,
		Source: "MAC:11:22:33:44:55:66",
		// Invalid Destination
		Destination: "invalid:a-BB-44-55",
	}) // Found error
	unfoundErrFailure := msgv.Validate(Message{Type: CreateMessageType}) // Unfound error
	fmt.Println(foundErrFailure == nil, foundErrSuccess1 == nil, foundErrSuccess2 == nil, unfoundErrFailure == nil)
	// Output: false true false false
}

func testSpansValidator(t *testing.T) {
	tests := []struct {
		description string
		msg         Message
		expectedErr error
	}{
		// Success case
		{
			description: "Valid spans",
			msg: Message{
				Spans: [][]string{{"parent", "name", "1234", "1234", "1234"}},
			},
		},
		// Failure case
		{
			description: "Invaild spans error, empty span",
			msg: Message{
				Spans: [][]string{
					{},
				},
			},
			expectedErr: ErrorInvalidSpanLength,
		},
		{
			description: "Invaild spans error, span length too large",
			msg: Message{
				Spans: [][]string{
					{"3", "3", "3", "3", "3", "3", "3"},
				},
			},
			expectedErr: ErrorInvalidSpanLength,
		},
		{
			description: "Invaild spans error, bad 'start time' 'duration' and 'status' components",
			msg: Message{
				Spans: [][]string{
					{"parent", "name", "not start time", "not duration", "not status"},
				},
			},
			expectedErr: ErrorInvalidSpanFormat,
		},
		{
			description: "Invaild spans error, bad 'parent' and 'name' components",
			msg: Message{
				Spans: [][]string{
					{"", "", "1234", "1234", "1234"},
				},
			},
			expectedErr: ErrorInvalidSpanFormat,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := SpansValidator(tc.msg)
			if expectedErr := tc.expectedErr; expectedErr != nil {
				if ve, ok := expectedErr.(ValidatorError); ok {
					expectedErr = ve.Err
				}

				assert.ErrorIs(err, expectedErr)
				return
			}

			assert.NoError(err)
		})
	}
}

func testSimpleEventTypeValidator(t *testing.T) {
	tests := []struct {
		description string
		msg         Message
		expectedErr error
	}{
		// Success case
		{
			description: "SimpleEventMessageType success",
			msg:         Message{Type: SimpleEventMessageType},
		},
		// Failure case
		{
			description: "Invalid0MessageType error",
			msg:         Message{Type: Invalid0MessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "SimpleRequestResponseMessageType error",
			msg:         Message{Type: SimpleRequestResponseMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "CreateMessageType error",
			msg:         Message{Type: CreateMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "RetrieveMessageType error",
			msg:         Message{Type: RetrieveMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "UpdateMessageType error",
			msg:         Message{Type: UpdateMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "DeleteMessageType error",
			msg:         Message{Type: DeleteMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "ServiceRegistrationMessageType error",
			msg:         Message{Type: ServiceRegistrationMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "ServiceAliveMessageType error",
			msg:         Message{Type: ServiceAliveMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "UnknownMessageType error",
			msg:         Message{Type: UnknownMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "AuthorizationMessageType error",
			msg:         Message{Type: AuthorizationMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "Invalid0MessageType error",
			msg:         Message{Type: Invalid0MessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "Invalid1MessageType error",
			msg:         Message{Type: Invalid1MessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "lastMessageType error",
			msg:         Message{Type: lastMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "Nonexistent negative MessageType error",
			msg:         Message{Type: -10},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "Nonexistent positive MessageType error",
			msg:         Message{Type: lastMessageType + 1},
			expectedErr: ErrorNotSimpleEventType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := SimpleEventTypeValidator(tc.msg)
			if expectedErr := tc.expectedErr; expectedErr != nil {
				if ve, ok := expectedErr.(ValidatorError); ok {
					expectedErr = ve.Err
				}

				assert.ErrorIs(err, expectedErr)
				return
			}

			assert.NoError(err)
		})
	}
}

func testSimpleResponseRequestTypeValidator(t *testing.T) {
	tests := []struct {
		description string
		msg         Message
		expectedErr error
	}{
		// Success case
		{
			description: "SimpleRequestResponseMessageType success",
			msg:         Message{Type: SimpleRequestResponseMessageType},
		},
		// Failure case
		{
			description: "Invalid0MessageType error",
			msg:         Message{Type: Invalid0MessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "SimpleEventMessageType error",
			msg:         Message{Type: SimpleEventMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "CreateMessageType error",
			msg:         Message{Type: CreateMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "RetrieveMessageType error",
			msg:         Message{Type: RetrieveMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "UpdateMessageType error",
			msg:         Message{Type: UpdateMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "DeleteMessageType error",
			msg:         Message{Type: DeleteMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "ServiceRegistrationMessageType error",
			msg:         Message{Type: ServiceRegistrationMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "ServiceAliveMessageType error",
			msg:         Message{Type: ServiceAliveMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "UnknownMessageType error",
			msg:         Message{Type: UnknownMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "AuthorizationMessageType error",
			msg:         Message{Type: AuthorizationMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "Invalid0MessageType error",
			msg:         Message{Type: Invalid0MessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "Invalid1MessageType error",
			msg:         Message{Type: Invalid1MessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "lastMessageType error",
			msg:         Message{Type: lastMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "Nonexistent negative MessageType error",
			msg:         Message{Type: -10},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "Nonexistent positive MessageType error",
			msg:         Message{Type: lastMessageType + 1},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := SimpleResponseRequestTypeValidator(tc.msg)
			if expectedErr := tc.expectedErr; expectedErr != nil {
				if ve, ok := expectedErr.(ValidatorError); ok {
					expectedErr = ve.Err
				}

				assert.ErrorIs(err, expectedErr)
				return
			}

			assert.NoError(err)
		})
	}
}
