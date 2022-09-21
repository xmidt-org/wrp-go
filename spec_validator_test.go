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

func TestSpecHelperValidators(t *testing.T) {
	tests := []struct {
		description string
		test        func(*testing.T)
	}{
		{"UTF8Validator", testUTF8Validator},
		{"MessageTypeValidator", testMessageTypeValidator},
		{"SourceValidator", testSourceValidator},
		{"DestinationValidator", testDestinationValidator},
		{"validateLocator", testValidateLocator},
	}

	for _, tc := range tests {
		t.Run(tc.description, tc.test)
	}
}

func TestSpecValidators(t *testing.T) {
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
			description: "Valid spec success",
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
			description: "Invaild spec error",
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
				Spans:                   [][]string{{"1", "2"}, {"3"}},
				IncludeSpans:            &expectedIncludeSpans,
				Path:                    "/some/where/over/the/rainbow",
				Payload:                 []byte{1, 2, 3, 4, 0xff, 0xce},
				ServiceName:             "ServiceName",
				// Not UFT8 URL string
				URL:        "someURL\xed\xbf\xbf.com",
				PartnerIDs: []string{"foo"},
				SessionID:  "sessionID123",
			},
			expectedErr: []error{ErrorInvalidMessageType, ErrorInvalidSource, ErrorInvalidDestination, ErrorInvalidMessageEncoding},
		},
		{
			description: "Invaild spec error, empty message",
			msg:         Message{},
			expectedErr: []error{ErrorInvalidMessageType, ErrorInvalidSource, ErrorInvalidDestination},
		},
		{
			description: "Invaild spec error, nonexistent MessageType",
			msg: Message{
				Type:        lastMessageType + 1,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
			expectedErr: []error{ErrorInvalidMessageType},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := SpecValidators().Validate(tc.msg)
			if tc.expectedErr != nil {
				for _, e := range tc.expectedErr {
					var targetErr ValidatorError

					assert.ErrorAs(e, &targetErr)
					assert.ErrorIs(err, targetErr.Err)
				}

				return
			}

			assert.NoError(err)
		})
	}
}

func ExampleTypeValidator_Validate_specValidators() {
	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[MessageType]Validator{
			// Validates opinionated portions of the spec
			SimpleEventMessageType: SpecValidators(),
			// Only validates Source and nothing else
			SimpleRequestResponseMessageType: ValidatorFunc(SourceValidator),
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
		Type: SimpleEventMessageType,
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
		Spans:                   [][]string{{"1", "2"}, {"3"}},
		IncludeSpans:            &expectedIncludeSpans,
		Path:                    "/some/where/over/the/rainbow",
		// Not UFT8 Payload
		Payload:     []byte{1, 2, 3, 4, 0xff /* \xed\xbf\xbf is invalid */, 0xce},
		ServiceName: "ServiceName",
		// Not UFT8 URL string
		URL:        "someURL\xed\xbf\xbf.com",
		PartnerIDs: []string{"foo"},
		SessionID:  "sessionID123",
	}) // Found error
	foundErrSuccess1 := msgv.Validate(Message{
		Type:        SimpleEventMessageType,
		Source:      "MAC:11:22:33:44:55:66",
		Destination: "MAC:11:22:33:44:55:61",
	}) // Found success
	foundErrSuccess2 := msgv.Validate(Message{
		Type:        SimpleRequestResponseMessageType,
		Source:      "MAC:11:22:33:44:55:66",
		Destination: "invalid:a-BB-44-55",
	}) // Found success
	unfoundErrFailure := msgv.Validate(Message{Type: CreateMessageType}) // Unfound error
	fmt.Println(foundErrFailure == nil, foundErrSuccess1 == nil, foundErrSuccess2 == nil, unfoundErrFailure == nil)
	// Output: false true true false
}

func testUTF8Validator(t *testing.T) {
	var (
		expectedStatus                  int64 = 3471
		expectedRequestDeliveryResponse int64 = 34
		expectedIncludeSpans            bool  = true
	)

	tests := []struct {
		description string
		msg         Message
		expectedErr error
	}{
		// Success case
		{
			description: "UTF8 success",
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
		{
			description: "Not UTF8 error",
			msg: Message{
				Type:   SimpleRequestResponseMessageType,
				Source: "dns:external.com",
				// Not UFT8 Destination string
				Destination:             "MAC:\xed\xbf\xbf",
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
			expectedErr: ErrorInvalidMessageEncoding,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := UTF8Validator(tc.msg)
			if expectedErr := tc.expectedErr; expectedErr != nil {
				var targetErr ValidatorError

				assert.ErrorAs(expectedErr, &targetErr)
				assert.ErrorIs(err, targetErr.Err)
				return
			}

			assert.NoError(err)
		})
	}
}

func testMessageTypeValidator(t *testing.T) {
	tests := []struct {
		description string
		msg         Message
		expectedErr error
	}{
		// Success case
		{
			description: "AuthorizationMessageType success",
			msg:         Message{Type: AuthorizationMessageType},
		},
		{
			description: "SimpleRequestResponseMessageType success",
			msg:         Message{Type: SimpleRequestResponseMessageType},
		},
		{
			description: "SimpleEventMessageType success",
			msg:         Message{Type: SimpleEventMessageType},
		},
		{
			description: "CreateMessageType success",
			msg:         Message{Type: CreateMessageType},
		},
		{
			description: "RetrieveMessageType success",
			msg:         Message{Type: RetrieveMessageType},
		},
		{
			description: "UpdateMessageType success",
			msg:         Message{Type: UpdateMessageType},
		},
		{
			description: "DeleteMessageType success",
			msg:         Message{Type: DeleteMessageType},
		},
		{
			description: "ServiceRegistrationMessageType success",
			msg:         Message{Type: ServiceRegistrationMessageType},
		},
		{
			description: "ServiceAliveMessageType success",
			msg:         Message{Type: ServiceAliveMessageType},
		},
		{
			description: "UnknownMessageType success",
			msg:         Message{Type: UnknownMessageType},
		},
		// Failure case
		{
			description: "Invalid0MessageType error",
			msg:         Message{Type: Invalid0MessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Invalid0MessageType error",
			msg:         Message{Type: Invalid0MessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Invalid1MessageType error",
			msg:         Message{Type: Invalid1MessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "lastMessageType error",
			msg:         Message{Type: lastMessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Nonexistent negative MessageType error",
			msg:         Message{Type: -10},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Nonexistent positive MessageType error",
			msg:         Message{Type: lastMessageType + 1},
			expectedErr: ErrorInvalidMessageType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := MessageTypeValidator(tc.msg)
			if expectedErr := tc.expectedErr; expectedErr != nil {
				var targetErr ValidatorError

				assert.ErrorAs(expectedErr, &targetErr)
				assert.ErrorIs(err, targetErr.Err)
				return
			}

			assert.NoError(err)
		})
	}
}

func testSourceValidator(t *testing.T) {
	// SourceValidator is mainly a wrapper for validateLocator.
	// This test mainly ensures that SourceValidator returns nil for non errors
	// and wraps errors with ErrorInvalidSource.
	// testValidateLocator covers the actual spectrum of test cases.

	tests := []struct {
		description string
		msg         Message
		expectedErr error
	}{
		// Success case
		{
			description: "Source success",
			msg:         Message{Source: "MAC:11:22:33:44:55:66"},
		},
		// Failures
		{
			description: "Source error",
			msg:         Message{Source: "invalid:a-BB-44-55"},
			expectedErr: ErrorInvalidSource,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := SourceValidator(tc.msg)
			if expectedErr := tc.expectedErr; expectedErr != nil {
				var targetErr ValidatorError

				assert.ErrorAs(expectedErr, &targetErr)
				assert.ErrorIs(err, targetErr.Err)
				return
			}

			assert.NoError(err)
		})
	}
}

func testDestinationValidator(t *testing.T) {
	// DestinationValidator is mainly a wrapper for validateLocator.
	// This test mainly ensures that DestinationValidator returns nil for non errors
	// and wraps errors with ErrorInvalidDestination.
	// testValidateLocator covers the actual spectrum of test cases.

	tests := []struct {
		description string
		msg         Message
		expectedErr error
	}{
		// Success case
		{
			description: "Destination success",
			msg:         Message{Destination: "MAC:11:22:33:44:55:66"},
		},
		// Failures
		{
			description: "Destination error",
			msg:         Message{Destination: "invalid:a-BB-44-55"},
			expectedErr: ErrorInvalidDestination,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := DestinationValidator(tc.msg)
			if expectedErr := tc.expectedErr; expectedErr != nil {
				var targetErr ValidatorError

				assert.ErrorAs(expectedErr, &targetErr)
				assert.ErrorIs(err, targetErr.Err)
				return
			}

			assert.NoError(err)
		})
	}
}

func testValidateLocator(t *testing.T) {
	tests := []struct {
		description string
		value       string
		expectedErr error
	}{
		// mac success case
		{
			description: "Mac ID success, ':' delimiter",
			value:       "MAC:11:22:33:44:55:66",
			expectedErr: nil,
		},
		{
			description: "Mac ID success, no delimiter",
			value:       "MAC:11aaBB445566",
			expectedErr: nil,
		},
		{
			description: "Mac ID success, '-' delimiter",
			value:       "mac:11-aa-BB-44-55-66",
			expectedErr: nil,
		},
		{
			description: "Mac ID success, ',' delimiter",
			value:       "mac:11,aa,BB,44,55,66",
			expectedErr: nil,
		},
		{
			description: "Mac service success",
			value:       "mac:11,aa,BB,44,55,66/parodus/tag/test0",
			expectedErr: nil,
		},
		// Mac failure case
		{
			description: "Mac ID error, invalid mac ID character",
			value:       "MAC:invalid45566",
			expectedErr: errorInvalidCharacter,
		},
		{
			description: "Mac ID error, invalid mac ID length",
			value:       "mac:11-aa-BB-44-55",
			expectedErr: errorInvalidMacLength,
		},
		{
			description: "Mac ID error, no ID",
			value:       "mac:",
			expectedErr: errorEmptyAuthority,
		},
		// Serial success case
		{
			description: "Serial ID success",
			value:       "serial:anything Goes!",
			expectedErr: nil,
		},
		// Serial failure case
		{
			description: "Invalid serial ID error, no ID",
			value:       "serial:",
			expectedErr: errorEmptyAuthority,
		},
		// UUID success case
		// The variant specified in RFC4122
		{
			description: "UUID RFC4122 variant ID success",
			value:       "uuid:f47ac10b-58cc-0372-8567-0e02b2c3d479",
			expectedErr: nil,
		},
		{
			description: "UUID RFC4122 variant ID success, with Microsoft encoding",
			value:       "uuid:{f47ac10b-58cc-0372-8567-0e02b2c3d479}",
			expectedErr: nil,
		},
		// Reserved, NCS backward compatibility.
		{
			description: "UUID Reserved variant ID success, with URN lower case ",
			value:       "UUID:urn:uuid:f47ac10b-58cc-4372-0567-0e02b2c3d479",
			expectedErr: nil,
		},
		{
			description: "UUID Reserved variant ID success, with URN upper case",
			value:       "UUID:URN:UUID:f47ac10b-58cc-4372-0567-0e02b2c3d479",
			expectedErr: nil,
		},
		{
			description: "UUID Reserved variant ID success, without URN",
			value:       "UUID:f47ac10b-58cc-4372-0567-0e02b2c3d479",
			expectedErr: nil,
		},
		// Reserved, Microsoft Corporation backward compatibility.
		{
			description: "UUID Microsoft variant ID success",
			value:       "uuid:f47ac10b-58cc-4372-c567-0e02b2c3d479",
			expectedErr: nil,
		},
		// Reserved for future definition.
		{
			description: "UUID Future variant ID success",
			value:       "uuid:f47ac10b-58cc-4372-e567-0e02b2c3d479",
			expectedErr: nil,
		},
		// UUID failure case
		{
			description: "Invalid UUID ID error",
			value:       "uuid:invalid45566",
			expectedErr: errorInvalidUUID,
		},
		{
			description: "Invalid UUID ID error, with URN",
			value:       "uuid:URN:UUID:invalid45566",
			expectedErr: errorInvalidUUID,
		},
		{
			description: "Invalid UUID ID error, with Microsoft encoding",
			value:       "uuid:{invalid45566}",
			expectedErr: errorInvalidUUID,
		},
		{
			description: "Invalid UUID ID error, no ID",
			value:       "uuid:",
			expectedErr: errorEmptyAuthority,
		},
		// Event success case
		{
			description: "Event ID success",
			value:       "event:anything Goes!",
			expectedErr: nil,
		},
		// Event failure case
		{
			description: "Invalid event ID error, no ID",
			value:       "event:",
			expectedErr: errorEmptyAuthority,
		},
		// DNS success case
		{
			description: "DNS ID success",
			value:       "dns:anything Goes!",
			expectedErr: nil,
		},
		// DNS failure case
		{
			description: "Invalid DNS ID error, no ID",
			value:       "dns:",
			expectedErr: errorEmptyAuthority,
		},
		// Scheme failure case
		{
			description: "Invalid scheme error",
			value:       "invalid:a-BB-44-55",
			expectedErr: errorInvalidLocatorPattern,
		},
		{
			description: "Invalid scheme error, empty string",
			value:       "",
			expectedErr: errorInvalidLocatorPattern,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := validateLocator(tc.value)
			if expectedErr := tc.expectedErr; expectedErr != nil {
				assert.ErrorIs(err, expectedErr)
				return
			}

			assert.NoError(err)
		})
	}
}
