// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/sallust"
	"github.com/xmidt-org/touchstone"
	"github.com/xmidt-org/wrp-go/v3"
)

func TestSpecHelperValidators(t *testing.T) {
	tests := []struct {
		description string
		test        func(*testing.T)
	}{
		{"UTF8", testUTF8},
		{"MessageType", testMessageType},
		{"Source", testSource},
		{"Destination", testDestination},
		{"validateLocator", testValidateLocator},
	}

	for _, tc := range tests {
		t.Run(tc.description, tc.test)
	}
}

func TestSpecWithMetrics(t *testing.T) {
	var (
		expectedStatus                  int64 = 3471
		expectedRequestDeliveryResponse int64 = 34
		expectedIncludeSpans            bool  = true
	)

	tests := []struct {
		description string
		msg         wrp.Message
		expectedErr []error
	}{
		// Success case
		{
			description: "Valid spec success",
			msg: wrp.Message{
				Type:                    wrp.SimpleRequestResponseMessageType,
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
			description: "Duplicate validators",
			msg: wrp.Message{
				Type: wrp.Invalid0MessageType,
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
			description: "Invaild spec error",
			msg: wrp.Message{
				Type: wrp.Invalid0MessageType,
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
			msg:         wrp.Message{},
			expectedErr: []error{ErrorInvalidMessageType, ErrorInvalidSource, ErrorInvalidDestination},
		},
		{
			description: "Invaild spec error, nonexistent wrp.MessageType",
			msg: wrp.Message{
				Type:        wrp.LastMessageType + 1,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
			expectedErr: []error{ErrorInvalidMessageType},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			cfg := touchstone.Config{
				DefaultNamespace: "n",
				DefaultSubsystem: "s",
			}
			_, pr, err := touchstone.New(cfg)
			require.NoError(err)

			tf := touchstone.NewFactory(cfg, sallust.Default(), pr)
			sv, err := SpecWithMetrics(tf)
			require.NoError(err)

			err = sv.Validate(tc.msg, prometheus.Labels{})
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

func TestSpecWithDuplicateValidators(t *testing.T) {
	tests := []struct {
		description string
		msg         wrp.Message
		expectedErr []error
	}{
		// Failure case
		{
			description: "Duplicate validators",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			require := require.New(t)
			cfg := touchstone.Config{
				DefaultNamespace: "n",
				DefaultSubsystem: "s",
			}
			_, pr, err := touchstone.New(cfg)
			require.NoError(err)

			tf := touchstone.NewFactory(cfg, sallust.Default(), pr)
			_, err = NewUTF8WithMetric(tf)
			require.NoError(err)
			_, err = SpecWithMetrics(tf)
			require.Error(err)

			_, pr2, err := touchstone.New(cfg)
			require.NoError(err)

			f2 := touchstone.NewFactory(cfg, sallust.Default(), pr2)
			_, err = NewMessageTypeWithMetric(f2)
			require.NoError(err)
			_, err = SpecWithMetrics(f2)
			require.Error(err)

			_, pr3, err := touchstone.New(cfg)
			require.NoError(err)

			f3 := touchstone.NewFactory(cfg, sallust.Default(), pr3)
			_, err = NewSourceWithMetric(f3)
			require.NoError(err)
			_, err = SpecWithMetrics(f3)
			require.Error(err)

			_, pr4, err := touchstone.New(cfg)
			require.NoError(err)

			f4 := touchstone.NewFactory(cfg, sallust.Default(), pr4)
			_, err = NewDestinationWithMetric(f4)
			require.NoError(err)
			_, err = SpecWithMetrics(f4)
			require.Error(err)
		})
	}
}

func ExampleTypeValidator_Validate_specValidators() {
	cfg := touchstone.Config{
		DefaultNamespace: "n",
		DefaultSubsystem: "s",
	}
	_, pr, err := touchstone.New(cfg)
	if err != nil {
		panic(err)
	}

	tf := touchstone.NewFactory(cfg, sallust.Default(), pr)
	specv, err := SpecWithMetrics(tf)
	if err != nil {
		panic(err)
	}

	_, pr2, err := touchstone.New(cfg)
	if err != nil {
		panic(err)
	}

	f2 := touchstone.NewFactory(cfg, sallust.Default(), pr2)
	sv, err := NewSourceWithMetric(f2)
	if err != nil {
		panic(err)
	}

	ai, err := NewAlwaysInvalidWithMetric(tf)
	if err != nil {
		panic(err)
	}

	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[wrp.MessageType]Validator{
			// Validates opinionated portions of the spec
			wrp.SimpleEventMessageType: specv,
			// Only validates Source and nothing else
			wrp.SimpleRequestResponseMessageType: sv,
		},
		// Validates unfound msg types
		ai,
		tf)
	if err != nil {
		return
	}

	var (
		expectedStatus                  int64 = 3471
		expectedRequestDeliveryResponse int64 = 34
		expectedIncludeSpans            bool  = true
	)
	foundErrFailure := msgv.Validate(wrp.Message{
		Type: wrp.SimpleEventMessageType,
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
	}, prometheus.Labels{}) // Found error
	foundErrSuccess1 := msgv.Validate(wrp.Message{
		Type:        wrp.SimpleEventMessageType,
		Source:      "MAC:11:22:33:44:55:66",
		Destination: "MAC:11:22:33:44:55:61",
	}, prometheus.Labels{}) // Found success
	foundErrSuccess2 := msgv.Validate(wrp.Message{
		Type:        wrp.SimpleRequestResponseMessageType,
		Source:      "MAC:11:22:33:44:55:66",
		Destination: "invalid:a-BB-44-55",
	}, prometheus.Labels{}) // Found success
	unfoundErrFailure := msgv.Validate(wrp.Message{Type: wrp.CreateMessageType}, prometheus.Labels{}) // Unfound error
	fmt.Println(foundErrFailure == nil, foundErrSuccess1 == nil, foundErrSuccess2 == nil, unfoundErrFailure == nil)
	// Output: false true true false
}

func testUTF8(t *testing.T) {
	var (
		expectedStatus                  int64 = 3471
		expectedRequestDeliveryResponse int64 = 34
		expectedIncludeSpans            bool  = true
	)

	tests := []struct {
		description string
		msg         wrp.Message
		expectedErr error
	}{
		// Success case
		{
			description: "UTF8 success",
			msg: wrp.Message{
				Type:                    wrp.SimpleRequestResponseMessageType,
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
			msg: wrp.Message{
				Type:   wrp.SimpleRequestResponseMessageType,
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
			err := UTF8(tc.msg)
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

func testMessageType(t *testing.T) {
	tests := []struct {
		description string
		msg         wrp.Message
		expectedErr error
	}{
		// Success case
		{
			description: "AuthorizationMessageType success",
			msg:         wrp.Message{Type: wrp.AuthorizationMessageType},
		},
		{
			description: "SimpleRequestResponseMessageType success",
			msg:         wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
		},
		{
			description: "SimpleEventMessageType success",
			msg:         wrp.Message{Type: wrp.SimpleEventMessageType},
		},
		{
			description: "CreateMessageType success",
			msg:         wrp.Message{Type: wrp.CreateMessageType},
		},
		{
			description: "RetrieveMessageType success",
			msg:         wrp.Message{Type: wrp.RetrieveMessageType},
		},
		{
			description: "UpdateMessageType success",
			msg:         wrp.Message{Type: wrp.UpdateMessageType},
		},
		{
			description: "DeleteMessageType success",
			msg:         wrp.Message{Type: wrp.DeleteMessageType},
		},
		{
			description: "ServiceRegistrationMessageType success",
			msg:         wrp.Message{Type: wrp.ServiceRegistrationMessageType},
		},
		{
			description: "ServiceAliveMessageType success",
			msg:         wrp.Message{Type: wrp.ServiceAliveMessageType},
		},
		{
			description: "UnknownMessageType success",
			msg:         wrp.Message{Type: wrp.UnknownMessageType},
		},
		// Failure case
		{
			description: "Invalid0MessageType error",
			msg:         wrp.Message{Type: wrp.Invalid0MessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Invalid0MessageType error",
			msg:         wrp.Message{Type: wrp.Invalid0MessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Invalid1MessageType error",
			msg:         wrp.Message{Type: wrp.Invalid1MessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "lastMessageType error",
			msg:         wrp.Message{Type: wrp.LastMessageType},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Nonexistent negative wrp.MessageType error",
			msg:         wrp.Message{Type: -10},
			expectedErr: ErrorInvalidMessageType,
		},
		{
			description: "Nonexistent positive wrp.MessageType error",
			msg:         wrp.Message{Type: wrp.LastMessageType + 1},
			expectedErr: ErrorInvalidMessageType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := MessageType(tc.msg)
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

func testSource(t *testing.T) {
	// Source is mainly a wrapper for validateLocator.
	// This test mainly ensures that Source returns nil for non errors
	// and wraps errors with ErrorInvalidSource.
	// testValidateLocator covers the actual spectrum of test cases.

	tests := []struct {
		description string
		msg         wrp.Message
		expectedErr error
	}{
		// Success case
		{
			description: "Source success",
			msg:         wrp.Message{Source: "MAC:11:22:33:44:55:66"},
		},
		// Failures
		{
			description: "Source error",
			msg:         wrp.Message{Source: "invalid:a-BB-44-55"},
			expectedErr: ErrorInvalidSource,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := Source(tc.msg)
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

func testDestination(t *testing.T) {
	// Destination is mainly a wrapper for validateLocator.
	// This test mainly ensures that Destination returns nil for non errors
	// and wraps errors with ErrorInvalidDestination.
	// testValidateLocator covers the actual spectrum of test cases.

	tests := []struct {
		description string
		msg         wrp.Message
		expectedErr error
	}{
		// Success case
		{
			description: "Destination success",
			msg:         wrp.Message{Destination: "MAC:11:22:33:44:55:66"},
		},
		// Failures
		{
			description: "Destination error",
			msg:         wrp.Message{Destination: "invalid:a-BB-44-55"},
			expectedErr: ErrorInvalidDestination,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := Destination(tc.msg)
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
