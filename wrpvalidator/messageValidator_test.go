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
		msg         wrp.Message
		expectedErr []error
	}{
		// Success case
		{
			description: "Valid simple event message success",
			msg: wrp.Message{
				Type:                    wrp.SimpleEventMessageType,
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
			msg: wrp.Message{
				Type: wrp.Invalid0MessageType,
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
			msg:         wrp.Message{},
			expectedErr: []error{ErrorInvalidMessageType, ErrorInvalidSource, ErrorInvalidDestination, ErrorNotSimpleEventType},
		},
		{
			description: "Invaild simple event message error, non wrp.SimpleEventMessageType",
			msg: wrp.Message{
				Type:        wrp.CreateMessageType,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
			expectedErr: []error{ErrorNotSimpleEventType},
		},
		{
			description: "Invaild simple event message error, nonexistent MessageType",
			msg: wrp.Message{
				Type:        wrp.LastMessageType + 1,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
			expectedErr: []error{ErrorInvalidMessageType, ErrorNotSimpleEventType},
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

			f := touchstone.NewFactory(cfg, sallust.Default(), pr)
			sev, err := SimpleEventValidators(f)
			require.NoError(err)
			err = sev.Validate(tc.msg, prometheus.Labels{})
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

func TestSimpleEventValidatorsBadTouchStoneFactory(t *testing.T) {
	tests := []struct {
		description string
		msg         wrp.Message
		expectedErr []error
	}{
		// Failure case
		{
			description: "Invaild touchstone factory",
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

			f := touchstone.NewFactory(cfg, sallust.Default(), pr)
			_, err = NewUTF8Validator(f)
			require.NoError(err)
			_, err = SimpleEventValidators(f)
			require.Error(err)

			_, pr2, err := touchstone.New(cfg)
			require.NoError(err)

			f2 := touchstone.NewFactory(cfg, sallust.Default(), pr2)
			_, err = NewSimpleEventTypeValidator(f2)
			require.NoError(err)
			_, err = SimpleEventValidators(f2)
			require.Error(err)
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
		msg         wrp.Message
		expectedErr []error
	}{
		// Success case
		{
			description: "Valid simple request response message success",
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
			msg:         wrp.Message{},
			expectedErr: []error{ErrorInvalidMessageType, ErrorInvalidSource, ErrorInvalidDestination, ErrorNotSimpleResponseRequestType},
		},
		{
			description: "Invaild simple request response message error, non wrp.SimpleEventMessageType",
			msg: wrp.Message{
				Type:        wrp.CreateMessageType,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
			expectedErr: []error{ErrorNotSimpleResponseRequestType},
		},
		{
			description: "Invaild simple request response message error, nonexistent MessageType",
			msg: wrp.Message{
				Type:        wrp.LastMessageType + 1,
				Source:      "dns:external.com",
				Destination: "MAC:11:22:33:44:55:66",
			},
			expectedErr: []error{ErrorInvalidMessageType, ErrorNotSimpleResponseRequestType},
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

			f := touchstone.NewFactory(cfg, sallust.Default(), pr)
			srv, err := SimpleResponseRequestValidators(f)
			require.NoError(err)
			err = srv.Validate(tc.msg, prometheus.Labels{})
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

func TestSimpleResponseRequestValidatorsBadTouchStoneFactory(t *testing.T) {
	tests := []struct {
		description string
		msg         wrp.Message
		expectedErr []error
	}{
		// Failure case
		{
			description: "Invaild touchstone factory",
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

			f := touchstone.NewFactory(cfg, sallust.Default(), pr)
			_, err = NewUTF8Validator(f)
			require.NoError(err)
			_, err = SimpleResponseRequestValidators(f)
			require.Error(err)

			_, pr2, err := touchstone.New(cfg)
			require.NoError(err)

			f2 := touchstone.NewFactory(cfg, sallust.Default(), pr2)
			_, err = NewSimpleResponseRequestTypeValidator(f2)
			require.NoError(err)
			_, err = SimpleResponseRequestValidators(f2)
			require.Error(err)

			_, pr3, err := touchstone.New(cfg)
			require.NoError(err)

			f3 := touchstone.NewFactory(cfg, sallust.Default(), pr3)
			_, err = NewSpansValidator(f3)
			require.NoError(err)
			_, err = SimpleResponseRequestValidators(f3)
			require.Error(err)
		})
	}
}

func ExampleTypeValidator_Validate_simpleTypesValidators() {
	cfg := touchstone.Config{
		DefaultNamespace: "n",
		DefaultSubsystem: "s",
	}
	_, pr, err := touchstone.New(cfg)
	if err != nil {
		panic(err)
	}

	f := touchstone.NewFactory(cfg, sallust.Default(), pr)
	sev, err := SimpleEventValidators(f)
	if err != nil {
		panic(err)
	}

	_, pr2, err := touchstone.New(cfg)
	if err != nil {
		panic(err)
	}

	f2 := touchstone.NewFactory(cfg, sallust.Default(), pr2)
	srv, err := SimpleResponseRequestValidators(f2)
	if err != nil {
		panic(err)
	}

	aiv, err := NewAlwaysInvalid(f)
	if err != nil {
		panic(err)
	}

	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[wrp.MessageType]Validator{
			wrp.SimpleEventMessageType:           sev,
			wrp.SimpleRequestResponseMessageType: srv,
		},
		// Validates unfound msg types
		aiv,
		f)
	if err != nil {
		return
	}

	var (
		expectedStatus                  int64 = 3471
		expectedRequestDeliveryResponse int64 = 34
		expectedIncludeSpans            bool  = true
	)
	foundErrFailure := msgv.Validate(wrp.Message{
		Type: wrp.SimpleRequestResponseMessageType,
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
	}, prometheus.Labels{}) // Found error
	foundErrSuccess1 := msgv.Validate(wrp.Message{
		Type:        wrp.SimpleRequestResponseMessageType,
		Source:      "MAC:11:22:33:44:55:66",
		Destination: "MAC:11:22:33:44:55:61",
	}, prometheus.Labels{}) // Found success
	foundErrSuccess2 := msgv.Validate(wrp.Message{
		Type:   wrp.SimpleEventMessageType,
		Source: "MAC:11:22:33:44:55:66",
		// Invalid Destination
		Destination: "invalid:a-BB-44-55",
	}, prometheus.Labels{}) // Found error
	unfoundErrFailure := msgv.Validate(wrp.Message{Type: wrp.CreateMessageType}, prometheus.Labels{}) // Unfound error
	fmt.Println(foundErrFailure == nil, foundErrSuccess1 == nil, foundErrSuccess2 == nil, unfoundErrFailure == nil)
	// Output: false true false false
}

func testSpansValidator(t *testing.T) {
	tests := []struct {
		description string
		msg         wrp.Message
		expectedErr error
	}{
		// Success case
		{
			description: "Valid spans",
			msg: wrp.Message{
				Spans: [][]string{{"parent", "name", "1234", "1234", "1234"}},
			},
		},
		// Failure case
		{
			description: "Invaild spans error, empty span",
			msg: wrp.Message{
				Spans: [][]string{
					{},
				},
			},
			expectedErr: ErrorInvalidSpanLength,
		},
		{
			description: "Invaild spans error, span length too large",
			msg: wrp.Message{
				Spans: [][]string{
					{"3", "3", "3", "3", "3", "3", "3"},
				},
			},
			expectedErr: ErrorInvalidSpanLength,
		},
		{
			description: "Invaild spans error, bad 'start time' 'duration' and 'status' components",
			msg: wrp.Message{
				Spans: [][]string{
					{"parent", "name", "not start time", "not duration", "not status"},
				},
			},
			expectedErr: ErrorInvalidSpanFormat,
		},
		{
			description: "Invaild spans error, bad 'parent' and 'name' components",
			msg: wrp.Message{
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
				var targetErr ValidatorError

				assert.ErrorAs(expectedErr, &targetErr)
				assert.ErrorIs(err, targetErr.Err)
				return
			}

			assert.NoError(err)
		})
	}
}

func testSimpleEventTypeValidator(t *testing.T) {
	tests := []struct {
		description string
		msg         wrp.Message
		expectedErr error
	}{
		// Success case
		{
			description: "wrp.SimpleEventMessageType success",
			msg:         wrp.Message{Type: wrp.SimpleEventMessageType},
		},
		// Failure case
		{
			description: "Invalid0MessageType error",
			msg:         wrp.Message{Type: wrp.Invalid0MessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "wrp.SimpleRequestResponseMessageType error",
			msg:         wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "wrp.CreateMessageType error",
			msg:         wrp.Message{Type: wrp.CreateMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "RetrieveMessageType error",
			msg:         wrp.Message{Type: wrp.RetrieveMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "UpdateMessageType error",
			msg:         wrp.Message{Type: wrp.UpdateMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "DeleteMessageType error",
			msg:         wrp.Message{Type: wrp.DeleteMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "ServiceRegistrationMessageType error",
			msg:         wrp.Message{Type: wrp.ServiceRegistrationMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "ServiceAliveMessageType error",
			msg:         wrp.Message{Type: wrp.ServiceAliveMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "UnknownMessageType error",
			msg:         wrp.Message{Type: wrp.UnknownMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "AuthorizationMessageType error",
			msg:         wrp.Message{Type: wrp.AuthorizationMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "Invalid0MessageType error",
			msg:         wrp.Message{Type: wrp.Invalid0MessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "Invalid1MessageType error",
			msg:         wrp.Message{Type: wrp.Invalid1MessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "lastMessageType error",
			msg:         wrp.Message{Type: wrp.LastMessageType},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "Nonexistent negative MessageType error",
			msg:         wrp.Message{Type: -10},
			expectedErr: ErrorNotSimpleEventType,
		},
		{
			description: "Nonexistent positive MessageType error",
			msg:         wrp.Message{Type: wrp.LastMessageType + 1},
			expectedErr: ErrorNotSimpleEventType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := SimpleEventTypeValidator(tc.msg)
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

func testSimpleResponseRequestTypeValidator(t *testing.T) {
	tests := []struct {
		description string
		msg         wrp.Message
		expectedErr error
	}{
		// Success case
		{
			description: "wrp.SimpleRequestResponseMessageType success",
			msg:         wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
		},
		// Failure case
		{
			description: "Invalid0MessageType error",
			msg:         wrp.Message{Type: wrp.Invalid0MessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "wrp.SimpleEventMessageType error",
			msg:         wrp.Message{Type: wrp.SimpleEventMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "wrp.CreateMessageType error",
			msg:         wrp.Message{Type: wrp.CreateMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "RetrieveMessageType error",
			msg:         wrp.Message{Type: wrp.RetrieveMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "UpdateMessageType error",
			msg:         wrp.Message{Type: wrp.UpdateMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "DeleteMessageType error",
			msg:         wrp.Message{Type: wrp.DeleteMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "ServiceRegistrationMessageType error",
			msg:         wrp.Message{Type: wrp.ServiceRegistrationMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "ServiceAliveMessageType error",
			msg:         wrp.Message{Type: wrp.ServiceAliveMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "UnknownMessageType error",
			msg:         wrp.Message{Type: wrp.UnknownMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "AuthorizationMessageType error",
			msg:         wrp.Message{Type: wrp.AuthorizationMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "Invalid0MessageType error",
			msg:         wrp.Message{Type: wrp.Invalid0MessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "Invalid1MessageType error",
			msg:         wrp.Message{Type: wrp.Invalid1MessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "lastMessageType error",
			msg:         wrp.Message{Type: wrp.LastMessageType},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "Nonexistent negative MessageType error",
			msg:         wrp.Message{Type: -10},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
		{
			description: "Nonexistent positive MessageType error",
			msg:         wrp.Message{Type: wrp.LastMessageType + 1},
			expectedErr: ErrorNotSimpleResponseRequestType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := SimpleResponseRequestTypeValidator(tc.msg)
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
