// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestNewValidatorError(t *testing.T) {
	tests := []struct {
		description string
		err         error
		m           string
		f           []string
		expectedErr error
	}{
		// Success case
		{
			description: "Valid args",
			err:         errors.New("Test"),
			m:           "extra message",
			f:           []string{"Type", "Source", "PayloadRelatedField"},
		},
		{
			description: "No Feilds",
			err:         errors.New("Test"),
			m:           "extra message",
			f:           nil,
		},
		{
			description: "Nil Err",
			err:         nil,
			m:           "extra message",
			f:           []string{"Type", "Source", "PayloadRelatedField"},
		},
		{
			description: "Empty Err",
			err:         errors.New(""),
			m:           "extra message",
			f:           []string{"Type", "Source", "PayloadRelatedField"},
		},
		{
			description: "Empty Message",
			err:         errors.New("Test"),
			m:           "",
			f:           []string{"Type", "Source", "PayloadRelatedField"},
		},
		// Failure case
		{
			description: "Nil Err and empty Message panic",
			err:         nil,
			m:           "",
			f:           nil,
			expectedErr: ErrorInvalidValidatorError,
		},
		{
			description: "Empty Err and Message panic",
			err:         errors.New(""),
			m:           "",
			f:           []string{"Type", "Source", "PayloadRelatedField"},
			expectedErr: ErrorInvalidValidatorError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			if tc.expectedErr != nil {
				assert.PanicsWithError(tc.expectedErr.Error(), func() { _ = NewValidatorError(tc.err, tc.m, tc.f) })
				return
			}

			require.NotPanics(func() { _ = NewValidatorError(tc.err, tc.m, tc.f) })
			verr := NewValidatorError(tc.err, tc.m, tc.f)
			assert.NotEmpty(verr.Error())
		})
	}
}

func TestValidators(t *testing.T) {
	subvs := Validators{}.AddFunc(AlwaysValid, nil, AlwaysInvalid)
	vs := Validators{}.AddFunc(AlwaysValid, nil, AlwaysInvalid)
	vs = vs.Add(subvs, nil)
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
			vs:          vs,
			msg:         Message{Type: SimpleEventMessageType},
			expectedErr: []error{ErrorInvalidMsgType, ErrorInvalidMsgType},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := tc.vs.Validate(tc.msg)
			if tc.expectedErr != nil {
				assert.Equal(multierr.Errors(err), tc.expectedErr)
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
		map[MessageType]Validator{SimpleEventMessageType: ValidatorFunc(AlwaysValid)},
		// Validates unfound msg types
		ValidatorFunc(AlwaysInvalid))
	fmt.Printf("%v %T", err == nil, msgv)
	// Output: true wrp.TypeValidator
}

func ExampleTypeValidator_Validate() {
	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[MessageType]Validator{SimpleEventMessageType: ValidatorFunc(AlwaysValid)},
		// Validates unfound msg types
		ValidatorFunc(AlwaysInvalid))
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
				SimpleEventMessageType: ValidatorFunc(AlwaysValid),
			},
			msg: Message{Type: SimpleEventMessageType},
		},
		{
			description: "Unfound success",
			m: map[MessageType]Validator{
				SimpleEventMessageType: ValidatorFunc(AlwaysInvalid),
			},
			defaultValidator: ValidatorFunc(AlwaysValid),
			msg:              Message{Type: CreateMessageType},
		},
		{
			description: "Unfound success, nil list of default Validators",
			m: map[MessageType]Validator{
				SimpleEventMessageType: ValidatorFunc(AlwaysInvalid),
			},
			defaultValidator: Validators{nil},
			msg:              Message{Type: CreateMessageType},
		},
		{
			description: "Unfound success, empty map of default Validators",
			m: map[MessageType]Validator{
				SimpleEventMessageType: ValidatorFunc(AlwaysInvalid),
			},
			defaultValidator: Validators{},
			msg:              Message{Type: CreateMessageType},
		},
		// Failure case
		{
			description: "Found error",
			m: map[MessageType]Validator{
				SimpleEventMessageType: ValidatorFunc(AlwaysInvalid),
			},
			defaultValidator: ValidatorFunc(AlwaysValid),
			msg:              Message{Type: SimpleEventMessageType},
			expectedErr:      ErrorInvalidMsgType,
		},
		{
			description: "Found error, nil Validator",
			m: map[MessageType]Validator{
				SimpleEventMessageType: nil,
			},
			msg:         Message{Type: SimpleEventMessageType},
			expectedErr: ErrorInvalidMsgType,
		},
		{
			description: "Unfound error",
			m: map[MessageType]Validator{
				SimpleEventMessageType: ValidatorFunc(AlwaysValid),
			},
			msg:         Message{Type: CreateMessageType},
			expectedErr: ErrorInvalidMsgType,
		},
		{
			description: "Unfound error, nil default Validators",
			m: map[MessageType]Validator{
				SimpleEventMessageType: ValidatorFunc(AlwaysInvalid),
			},
			defaultValidator: nil,
			msg:              Message{Type: CreateMessageType},
			expectedErr:      ErrorInvalidMsgType,
		},
		{
			description: "Unfound error, empty map of Validators",
			m:           map[MessageType]Validator{},
			msg:         Message{Type: CreateMessageType},
			expectedErr: ErrorInvalidMsgType,
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
				SimpleEventMessageType: ValidatorFunc(AlwaysValid),
			},
			defaultValidator: ValidatorFunc(AlwaysValid),
			expectedErr:      nil,
		},
		{
			description: "Omit default Validators success",
			m: map[MessageType]Validator{
				SimpleEventMessageType: ValidatorFunc(AlwaysValid),
			},
			expectedErr: nil,
		},
		// Failure case
		{
			description:      "Nil map of Validators error",
			m:                nil,
			defaultValidator: ValidatorFunc(AlwaysValid),
			expectedErr:      ErrorInvalidValidator,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			msgv, err := NewTypeValidator(tc.m, tc.defaultValidator)
			if expectedErr := tc.expectedErr; expectedErr != nil {
				var targetErr ValidatorError

				assert.ErrorAs(expectedErr, &targetErr)
				assert.ErrorIs(err, targetErr.Err)
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
			err := AlwaysValid(tc.msg)
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
			err := AlwaysInvalid(tc.msg)
			assert.ErrorIs(err, ErrorInvalidMsgType.Err)
		})
	}
}
