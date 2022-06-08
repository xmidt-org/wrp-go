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
	"go.uber.org/multierr"
)

func TestValidators(t *testing.T) {
	tests := []struct {
		description string
		vs          Validators
		msg         Message
		expectedErr []error
	}{
		// Success case
		{
			description: "Empty Validators success",
			vs:          Validators{},
			msg:         Message{Type: SimpleEventMessageType},
		},
		// Failure case
		{
			description: "Mix Validators error",
			vs:          Validators{AlwaysValid, nil, AlwaysInvalid, Validators{AlwaysValid, nil, AlwaysInvalid}},
			msg:         Message{Type: SimpleEventMessageType},
			expectedErr: []error{ErrInvalidMsgType, ErrInvalidMsgType},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := tc.vs.Validate(tc.msg)
			if tc.expectedErr != nil {
				assert.Equal(multierr.Errors(err), tc.expectedErr)
				for _, e := range tc.expectedErr {
					assert.ErrorIs(err, e)
				}
				return
			}

			assert.NoError(err)
		})
	}
}

func TestHelperValidators(t *testing.T) {
	tests := []struct {
		description string
		test        func(*testing.T)
	}{
		{"AlwaysInvalid", testAlwaysInvalid},
		{"AlwaysValid", testAlwaysValid},
	}

	for _, tc := range tests {
		t.Run(tc.description, tc.test)
	}
}

func TestTypeValidator(t *testing.T) {
	tests := []struct {
		description string
		test        func(*testing.T)
	}{
		{"Validate", testTypeValidatorValidate},
		{"Factory", testTypeValidatorFactory},
	}

	for _, tc := range tests {
		t.Run(tc.description, tc.test)
	}
}

func ExampleNewTypeValidator() {
	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[MessageType]Validator{SimpleEventMessageType: AlwaysValid},
		// Validates unfound msg types
		AlwaysInvalid)
	fmt.Printf("%v %T", err == nil, msgv)
	// Output: true wrp.TypeValidator
}

func ExampleTypeValidator_Validate() {
	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[MessageType]Validator{SimpleEventMessageType: AlwaysValid},
		// Validates unfound msg types
		AlwaysInvalid)
	if err != nil {
		return
	}

	foundErr := msgv.Validate(Message{Type: SimpleEventMessageType}) // Found success
	unfoundErr := msgv.Validate(Message{Type: CreateMessageType})    // Unfound error
	fmt.Println(foundErr == nil, unfoundErr == nil)
	// Output: true false
}

func testTypeValidatorValidate(t *testing.T) {
	tests := []struct {
		description      string
		m                map[MessageType]Validator
		defaultValidator Validator
		msg              Message
		expectedErr      error
	}{
		// Success case
		{
			description: "Found success",
			m: map[MessageType]Validator{
				SimpleEventMessageType: AlwaysValid,
			},
			msg: Message{Type: SimpleEventMessageType},
		},
		{
			description: "Unfound success",
			m: map[MessageType]Validator{
				SimpleEventMessageType: AlwaysInvalid,
			},
			defaultValidator: AlwaysValid,
			msg:              Message{Type: CreateMessageType},
		},
		{
			description: "Unfound success, nil list of default Validators",
			m: map[MessageType]Validator{
				SimpleEventMessageType: AlwaysInvalid,
			},
			defaultValidator: Validators{nil},
			msg:              Message{Type: CreateMessageType},
		},
		{
			description: "Unfound success, empty map of default Validators",
			m: map[MessageType]Validator{
				SimpleEventMessageType: AlwaysInvalid,
			},
			defaultValidator: Validators{},
			msg:              Message{Type: CreateMessageType},
		},
		// Failure case
		{
			description: "Found error",
			m: map[MessageType]Validator{
				SimpleEventMessageType: AlwaysInvalid,
			},
			defaultValidator: AlwaysValid,
			msg:              Message{Type: SimpleEventMessageType},
			expectedErr:      ErrInvalidMsgType,
		},
		{
			description: "Found error, nil Validator",
			m: map[MessageType]Validator{
				SimpleEventMessageType: nil,
			},
			msg:         Message{Type: SimpleEventMessageType},
			expectedErr: ErrInvalidMsgType,
		},
		{
			description: "Unfound error",
			m: map[MessageType]Validator{
				SimpleEventMessageType: AlwaysValid,
			},
			msg:         Message{Type: CreateMessageType},
			expectedErr: ErrInvalidMsgType,
		},
		{
			description: "Unfound error, nil default Validators",
			m: map[MessageType]Validator{
				SimpleEventMessageType: AlwaysInvalid,
			},
			defaultValidator: nil,
			msg:              Message{Type: CreateMessageType},
			expectedErr:      ErrInvalidMsgType,
		},
		{
			description: "Unfound error, empty map of Validators",
			m:           map[MessageType]Validator{},
			msg:         Message{Type: CreateMessageType},
			expectedErr: ErrInvalidMsgType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			msgv, err := NewTypeValidator(tc.m, tc.defaultValidator)
			require.NoError(err)
			require.NotNil(msgv)
			assert.NotZero(msgv)
			err = msgv.Validate(tc.msg)
			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
				return
			}

			assert.NoError(err)
		})
	}
}

func testTypeValidatorFactory(t *testing.T) {
	tests := []struct {
		description      string
		m                map[MessageType]Validator
		defaultValidator Validator
		expectedErr      error
	}{
		// Success case
		{
			description: "Default Validators success",
			m: map[MessageType]Validator{
				SimpleEventMessageType: AlwaysValid,
			},
			defaultValidator: AlwaysValid,
			expectedErr:      nil,
		},
		{
			description: "Omit default Validators success",
			m: map[MessageType]Validator{
				SimpleEventMessageType: AlwaysValid,
			},
			expectedErr: nil,
		},
		// Failure case
		{
			description:      "Nil map of Validators error",
			m:                nil,
			defaultValidator: AlwaysValid,
			expectedErr:      ErrInvalidValidator,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			msgv, err := NewTypeValidator(tc.m, tc.defaultValidator)
			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
				// Zero asserts that msgv is the zero value for its type and not nil.
				assert.Zero(msgv)
				return
			}

			assert.NoError(err)
			assert.NotNil(msgv)
			assert.NotZero(msgv)
		})
	}
}

func testAlwaysValid(t *testing.T) {
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
			description: "Not UTF8 success",
			msg: Message{
				Type:   SimpleRequestResponseMessageType,
				Source: "dns:external.com",
				// Not UFT8 Destination string
				Destination:             "mac:\xed\xbf\xbf",
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
			description: "Filled message success",
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
			description: "Empty message success",
			msg:         Message{},
		},
		{
			description: "Bad message type success",
			msg: Message{
				Type:        lastMessageType + 1,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := AlwaysValid.Validate(tc.msg)
			assert.NoError(err)
		})
	}
}

func testAlwaysInvalid(t *testing.T) {
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
		// Failure case
		{
			description: "Not UTF8 error",
			msg: Message{
				Type:   SimpleRequestResponseMessageType,
				Source: "dns:external.com",
				// Not UFT8 Destination string
				Destination:             "mac:\xed\xbf\xbf",
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
			description: "Filled message error",
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
			description: "Empty message error",
			msg:         Message{},
		},
		{
			description: "Bad message type error",
			msg: Message{
				Type:        lastMessageType + 1,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := AlwaysInvalid.Validate(tc.msg)
			assert.ErrorIs(err, ErrInvalidMsgType)
		})
	}
}
