// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"errors"
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/sallust"
	"github.com/xmidt-org/touchstone"
	"github.com/xmidt-org/wrp-go/v3"
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
	require := require.New(t)
	cfg := touchstone.Config{
		DefaultNamespace: "n",
		DefaultSubsystem: "s",
	}
	_, pr, err := touchstone.New(cfg)
	require.NoError(err)

	f := touchstone.NewFactory(cfg, sallust.Default(), pr)
	av, err := NewAlwaysValid(f)
	require.NoError(err)

	ai, err := NewAlwaysInvalid(f)
	require.NoError(err)

	subvs := Validators{}.AddFunc(av, nil, ai)
	vs := Validators{}.AddFunc(av, nil, ai)
	vs = vs.Add(subvs, nil)
	tests := []struct {
		description string
		vs          Validators
		msg         wrp.Message
		expectedErr []error
	}{
		// Success case
		{
			description: "Empty Validators success",
			vs:          Validators{},
			msg:         wrp.Message{Type: wrp.SimpleEventMessageType},
		},
		// Failure case
		{
			description: "Mix Validators error",
			vs:          vs,
			msg:         wrp.Message{Type: wrp.SimpleEventMessageType},
			expectedErr: []error{ErrorInvalidMsgType, ErrorInvalidMsgType},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := tc.vs.Validate(tc.msg, prometheus.Labels{})
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

func TestTypeValidatorBadTouchStoneFactory(t *testing.T) {
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
			ai, err := NewAlwaysInvalid(f)
			require.NoError(err)
			_, err = NewTypeValidator(map[wrp.MessageType]Validator{
				wrp.SimpleEventMessageType: ai,
			}, nil, f)
			require.Error(err)
		})
	}
}

func ExampleNewTypeValidator() {
	cfg := touchstone.Config{
		DefaultNamespace: "n",
		DefaultSubsystem: "s",
	}
	_, pr, err := touchstone.New(cfg)
	f := touchstone.NewFactory(cfg, sallust.Default(), pr)
	ai, err := NewAlwaysInvalid(f)
	if err != nil {
		panic(err)
	}

	av, err := NewAlwaysValid(f)
	if err != nil {
		panic(err)
	}

	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[wrp.MessageType]Validator{wrp.SimpleEventMessageType: av},
		// Validates unfound msg types
		ai,
		f)
	fmt.Printf("%v %T", err == nil, msgv)
	// Output: true wrpvalidator.TypeValidator
}

func ExampleTypeValidator_Validate() {
	cfg := touchstone.Config{
		DefaultNamespace: "n",
		DefaultSubsystem: "s",
	}
	_, pr, err := touchstone.New(cfg)
	f := touchstone.NewFactory(cfg, sallust.Default(), pr)
	ai, err := NewAlwaysInvalid(f)
	if err != nil {
		panic(err)
	}

	av, err := NewAlwaysValid(f)
	if err != nil {
		panic(err)
	}

	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[wrp.MessageType]Validator{wrp.SimpleEventMessageType: av},
		// Validates unfound msg types
		ai,
		f)
	if err != nil {
		return
	}

	foundErr := msgv.Validate(wrp.Message{Type: wrp.SimpleEventMessageType}, prometheus.Labels{}) // Found success
	unfoundErr := msgv.Validate(wrp.Message{Type: wrp.CreateMessageType}, prometheus.Labels{})    // Unfound error
	fmt.Println(foundErr == nil, unfoundErr == nil)
	// Output: true false
}

func testTypeValidatorValidate(t *testing.T) {
	r := require.New(t)
	cfg := touchstone.Config{
		DefaultNamespace: "n",
		DefaultSubsystem: "s",
	}
	_, pr, err := touchstone.New(cfg)
	r.NoError(err)

	f := touchstone.NewFactory(cfg, sallust.Default(), pr)
	av, err := NewAlwaysValid(f)
	r.NoError(err)

	ai, err := NewAlwaysInvalid(f)
	r.NoError(err)

	tests := []struct {
		description      string
		m                map[wrp.MessageType]Validator
		defaultValidator Validator
		msg              wrp.Message
		expectedErr      error
	}{
		// Success case
		{
			description: "Found success",
			m: map[wrp.MessageType]Validator{
				wrp.SimpleEventMessageType: av,
			},
			msg: wrp.Message{Type: wrp.SimpleEventMessageType},
		},
		{
			description: "Unfound success",
			m: map[wrp.MessageType]Validator{
				wrp.SimpleEventMessageType: ai,
			},
			defaultValidator: av,
			msg:              wrp.Message{Type: wrp.CreateMessageType},
		},
		{
			description: "Unfound success, nil list of default Validators",
			m: map[wrp.MessageType]Validator{
				wrp.SimpleEventMessageType: ai,
			},
			defaultValidator: Validators{nil},
			msg:              wrp.Message{Type: wrp.CreateMessageType},
		},
		{
			description: "Unfound success, empty map of default Validators",
			m: map[wrp.MessageType]Validator{
				wrp.SimpleEventMessageType: ai,
			},
			defaultValidator: Validators{},
			msg:              wrp.Message{Type: wrp.CreateMessageType},
		},
		// Failure case
		{
			description: "Found error",
			m: map[wrp.MessageType]Validator{
				wrp.SimpleEventMessageType: ai,
			},
			defaultValidator: av,
			msg:              wrp.Message{Type: wrp.SimpleEventMessageType},
			expectedErr:      ErrorInvalidMsgType,
		},
		{
			description: "Found error, nil Validator",
			m: map[wrp.MessageType]Validator{
				wrp.SimpleEventMessageType: nil,
			},
			msg:         wrp.Message{Type: wrp.SimpleEventMessageType},
			expectedErr: ErrorInvalidMsgType,
		},
		{
			description: "Unfound error",
			m: map[wrp.MessageType]Validator{
				wrp.SimpleEventMessageType: av,
			},
			msg:         wrp.Message{Type: wrp.CreateMessageType},
			expectedErr: ErrorInvalidMsgType,
		},
		{
			description: "Unfound error, nil default Validators",
			m: map[wrp.MessageType]Validator{
				wrp.SimpleEventMessageType: ai,
			},
			defaultValidator: nil,
			msg:              wrp.Message{Type: wrp.CreateMessageType},
			expectedErr:      ErrorInvalidMsgType,
		},
		{
			description: "Unfound error, empty map of Validators",
			m:           map[wrp.MessageType]Validator{},
			msg:         wrp.Message{Type: wrp.CreateMessageType},
			expectedErr: ErrorInvalidMsgType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			_, pr, err := touchstone.New(cfg)
			r.NoError(err)
			f := touchstone.NewFactory(cfg, sallust.Default(), pr)
			msgv, err := NewTypeValidator(tc.m, tc.defaultValidator, f)
			require.NoError(err)
			require.NotNil(msgv)
			assert.NotZero(msgv)
			err = msgv.Validate(tc.msg, prometheus.Labels{})
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
	require := require.New(t)
	cfg := touchstone.Config{
		DefaultNamespace: "n",
		DefaultSubsystem: "s",
	}
	_, pr, err := touchstone.New(cfg)
	require.NoError(err)

	f := touchstone.NewFactory(cfg, sallust.Default(), pr)
	av, err := NewAlwaysValid(f)
	require.NoError(err)

	tests := []struct {
		description      string
		m                map[wrp.MessageType]Validator
		defaultValidator Validator
		expectedErr      error
	}{
		// Success case
		{
			description: "Default Validators success",
			m: map[wrp.MessageType]Validator{
				wrp.SimpleEventMessageType: av,
			},
			defaultValidator: av,
			expectedErr:      nil,
		},
		{
			description: "Omit default Validators success",
			m: map[wrp.MessageType]Validator{
				wrp.SimpleEventMessageType: av,
			},
			expectedErr: nil,
		},
		// Failure case
		{
			description:      "Nil map of Validators error",
			m:                nil,
			defaultValidator: av,
			expectedErr:      ErrorInvalidValidator,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			msgv, err := NewTypeValidator(tc.m, tc.defaultValidator, f)
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
		msg         wrp.Message
		expectedErr []error
	}{
		// Success case
		{
			description: "Not UTF8 success",
			msg: wrp.Message{
				Type:   wrp.SimpleRequestResponseMessageType,
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
			description: "Empty message success",
			msg:         wrp.Message{},
		},
		{
			description: "Bad message type success",
			msg: wrp.Message{
				Type:        wrp.LastMessageType + 1,
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
		msg         wrp.Message
		expectedErr []error
	}{
		// Failure case
		{
			description: "Not UTF8 error",
			msg: wrp.Message{
				Type:   wrp.SimpleRequestResponseMessageType,
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
			description: "Empty message error",
			msg:         wrp.Message{},
		},
		{
			description: "Bad message type error",
			msg: wrp.Message{
				Type:        wrp.LastMessageType + 1,
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
