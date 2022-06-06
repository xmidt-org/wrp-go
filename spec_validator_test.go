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
	"github.com/stretchr/testify/require"
)

func testUTF8Validator(t *testing.T) {
	/*
		"\x85"  - 5 name value pairs
			"\xa8""msg_type"         : "\x03" // 3
			"\xa4""dest"             : "\xac""\xed\xbf\xbft-address"
			"\xa7""payload"          : "\xc4""\x03" - len 3
											 "123"
			"\xa6""source"           : "\xae""source-address"
			"\xb0""transaction_uuid" : "\xd9\x24""c07ee5e1-70be-444c-a156-097c767ad8aa"
	*/
	invalid := []byte{
		0x85,
		0xa8, 'm', 's', 'g', '_', 't', 'y', 'p', 'e', 0x03,
		0xa4, 'd', 'e', 's', 't', 0xac /* \xed\xbf\xbf is invalid */, 0xed, 0xbf, 0xbf, 't', '-', 'a', 'd', 'd', 'r', 'e', 's', 's',
		0xa7, 'p', 'a', 'y', 'l', 'o', 'a', 'd', 0xc4, 0x03, '1', '2', '3',
		0xa6, 's', 'o', 'u', 'r', 'c', 'e', 0xae, 's', 'o', 'u', 'r', 'c', 'e', '-', 'a', 'd', 'd', 'r', 'e', 's', 's',
		0xb0, 't', 'r', 'a', 'n', 's', 'a', 'c', 't', 'i', 'o', 'n', '_', 'u', 'u', 'i', 'd', 0xd9, 0x24, 'c', '0', '7', 'e', 'e', '5', 'e', '1', '-', '7', '0', 'b', 'e', '-', '4', '4', '4', 'c', '-', 'a', '1', '5', '6', '-', '0', '9', '7', 'c', '7', '6', '7', 'a', 'd', '8', 'a', 'a',
	}
	decoder := NewDecoderBytes(invalid, Msgpack)
	msg := new(Message)
	err := decoder.Decode(msg)
	require.NoError(t, err)
	tests := []struct {
		description string
		value       Message
		expectedErr []error
	}{
		// Success case
		{
			description: "UTF8 success",
			value:       Message{Source: "MAC:11:22:33:44:55:66"},
			expectedErr: nil,
		},
		{
			description: "Not UTF8 error",
			value:       *msg,
			expectedErr: []error{ErrorInvalidMessageEncoding},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := UTF8Validator(tc.value)
			if tc.expectedErr != nil {
				for _, e := range tc.expectedErr {
					assert.ErrorIs(err, e)
				}
				return
			}

			assert.NoError(err)
		})
	}
}

func testMessageTypeValidator(t *testing.T) {
	tests := []struct {
		description string
		value       Message
		expectedErr error
	}{
		// Success case
		{
			description: "AuthorizationMessageType success",
			value:       Message{Type: AuthorizationMessageType},
			expectedErr: nil,
		},
		{
			description: "SimpleRequestResponseMessageType success",
			value:       Message{Type: SimpleRequestResponseMessageType},
			expectedErr: nil,
		},
		{
			description: "SimpleEventMessageType success",
			value:       Message{Type: SimpleEventMessageType},
			expectedErr: nil,
		},
		{
			description: "CreateMessageType success",
			value:       Message{Type: CreateMessageType},
			expectedErr: nil,
		},
		{
			description: "RetrieveMessageType success",
			value:       Message{Type: RetrieveMessageType},
			expectedErr: nil,
		},
		{
			description: "UpdateMessageType success",
			value:       Message{Type: UpdateMessageType},
			expectedErr: nil,
		},
		{
			description: "DeleteMessageType success",
			value:       Message{Type: DeleteMessageType},
			expectedErr: nil,
		},
		{
			description: "ServiceRegistrationMessageType success",
			value:       Message{Type: ServiceRegistrationMessageType},
			expectedErr: nil,
		},
		{
			description: "ServiceAliveMessageType success",
			value:       Message{Type: ServiceAliveMessageType},
			expectedErr: nil,
		},
		{
			description: "UnknownMessageType success",
			value:       Message{Type: UnknownMessageType},
			expectedErr: nil,
		},
		// Failure case
		{
			description: "Invalid0MessageType error",
			value:       Message{Type: Invalid0MessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Invalid0MessageType error",
			value:       Message{Type: Invalid0MessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Invalid1MessageType error",
			value:       Message{Type: Invalid1MessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "lastMessageType error",
			value:       Message{Type: lastMessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Non-existing negative MessageType error",
			value:       Message{Type: -10},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Non-existing positive MessageType error",
			value:       Message{Type: lastMessageType + 1},
			expectedErr: ErrorInvalidMessageType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := MessageTypeValidator(tc.value)
			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
				return
			}

			assert.NoError(err)
		})
	}
}

func testSourceValidator(t *testing.T) {
	tests := []struct {
		description string
		value       Message
		expectedErr error
	}{
		// Success case
		{
			description: "Source success",
			value:       Message{Source: "MAC:11:22:33:44:55:66"},
			expectedErr: nil,
		},
		// Failures
		{
			description: "Source error",
			value:       Message{Source: "invalid:a-BB-44-55"},
			expectedErr: ErrorInvalidSource,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := SourceValidator(tc.value)
			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
				return
			}

			assert.NoError(err)
		})
	}
}

func testDestinationValidator(t *testing.T) {
	tests := []struct {
		description string
		value       Message
		expectedErr error
	}{
		// Success case
		{
			description: "Destination success",
			value:       Message{Destination: "MAC:11:22:33:44:55:66"},
			expectedErr: nil,
		},
		// Failures
		{
			description: "Destination error",
			value:       Message{Destination: "invalid:a-BB-44-55"},
			expectedErr: ErrorInvalidDestination,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := DestinationValidator(tc.value)
			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
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
		shouldErr   bool
	}{
		// mac success case
		{
			description: "Mac ID ':' delimiter success",
			value:       "MAC:11:22:33:44:55:66",
			shouldErr:   false,
		},
		{
			description: "Mac ID no delimiter success",
			value:       "MAC:11aaBB445566",
			shouldErr:   false,
		},
		{
			description: "Mac ID '-' delimiter success",
			value:       "mac:11-aa-BB-44-55-66",
			shouldErr:   false,
		},
		{
			description: "Mac ID ',' delimiter success",
			value:       "mac:11,aa,BB,44,55,66",
			shouldErr:   false,
		},
		{
			description: "Mac with service success",
			value:       "mac:11,aa,BB,44,55,66/parodus/tag/test0",
			shouldErr:   false,
		},
		// Mac failure case
		{
			description: "Invalid mac ID character error",
			value:       "MAC:invalid45566",
			shouldErr:   true,
		},
		{
			description: "Invalid mac ID length error",
			value:       "mac:11-aa-BB-44-55",
			shouldErr:   true,
		},
		// Serial success case
		{
			description: "Serial ID success",
			value:       "serial:anything Goes!",
			shouldErr:   false,
		},
		// UUID success case
		{
			description: "UUID RFC4122 variant ID success", // The variant specified in RFC4122
			value:       "uuid:f47ac10b-58cc-0372-8567-0e02b2c3d479",
			shouldErr:   false,
		},
		{
			description: "UUID RFC4122 variant with Microsoft encoding ID success", // The variant specified in RFC4122
			value:       "uuid:{f47ac10b-58cc-0372-8567-0e02b2c3d479}",
			shouldErr:   false,
		},
		{
			description: "UUID Reserved variant ID #1 success", // Reserved, NCS backward compatibility.
			value:       "UUID:urn:uuid:f47ac10b-58cc-4372-0567-0e02b2c3d479",
			shouldErr:   false,
		},
		{
			description: "UUID Reserved variant ID #2 success", // Reserved, NCS backward compatibility.
			value:       "UUID:URN:UUID:f47ac10b-58cc-4372-0567-0e02b2c3d479",
			shouldErr:   false,
		},
		{
			description: "UUID Reserved variant ID #3 success", // Reserved, NCS backward compatibility.
			value:       "UUID:f47ac10b-58cc-4372-0567-0e02b2c3d479",
			shouldErr:   false,
		},
		{
			description: "UUID Microsoft variant ID success", // Reserved, Microsoft Corporation backward compatibility.
			value:       "uuid:f47ac10b-58cc-4372-c567-0e02b2c3d479",
			shouldErr:   false,
		},
		{
			description: "UUID Future variant ID success", // Reserved for future definition.
			value:       "uuid:f47ac10b-58cc-4372-e567-0e02b2c3d479",
			shouldErr:   false,
		},
		// UUID failure case
		{
			description: "Invalid UUID ID #1 error",
			value:       "uuid:invalid45566",
			shouldErr:   true,
		},
		{
			description: "Invalid UUID ID #2 error",
			value:       "uuid:URN:UUID:invalid45566",
			shouldErr:   true,
		},
		{
			description: "Invalid UUID ID #3 error",
			value:       "uuid:{invalid45566}",
			shouldErr:   true,
		},
		// Event success case
		{
			description: "Event ID success",
			value:       "event:anything Goes!",
			shouldErr:   false,
		},
		// DNS success case
		{
			description: "DNS ID success",
			value:       "dns:anything Goes!",
			shouldErr:   false,
		},
		// Scheme failure case
		{
			description: "Invalid scheme error",
			value:       "invalid:a-BB-44-55",
			shouldErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := validateLocator(tc.value)
			if tc.shouldErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
		})
	}
}

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
	tests := []struct {
		description string
		value       Message
		expectedErr []error
	}{
		// Success case
		{
			description: "Valid specs success",
			value: Message{
				Type:        SimpleEventMessageType,
				Source:      "MAC:11:22:33:44:55:66",
				Destination: "MAC:11:22:33:44:55:66",
			},
			expectedErr: nil,
		},
		// Failure cases
		{
			description: "Invaild specs error",
			value: Message{
				Type:        Invalid0MessageType,
				Source:      "invalid:a-BB-44-55",
				Destination: "invalid:a-BB-44-55",
			},
			expectedErr: []error{ErrorInvalidMessageType, ErrorInvalidSource, ErrorInvalidDestination},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := SpecValidators().Validate(tc.value)
			if tc.expectedErr != nil {
				for _, e := range tc.expectedErr {
					assert.ErrorIs(err, e)
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
			SimpleRequestResponseMessageType: SourceValidator,
		},
		// Validates unfound msg types
		AlwaysInvalid)
	if err != nil {
		return
	}

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
	foundErrFailure := msgv.Validate(Message{
		Type:        Invalid0MessageType,
		Source:      "invalid:a-BB-44-55",
		Destination: "invalid:a-BB-44-55",
	}) // Found error
	unfoundErrFailure := msgv.Validate(Message{Type: CreateMessageType}) // Unfound error
	fmt.Println(foundErrSuccess1 == nil, foundErrSuccess2 == nil, foundErrFailure == nil, unfoundErrFailure == nil)
	// Output: true true false false
}
