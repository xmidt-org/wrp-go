// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeUnmarshalling(t *testing.T) {
	tests := []struct {
		description string
		config      []byte
		invalid     bool
	}{
		{
			description: "UnknownType valid",
			config:      []byte("unknown"),
		},
		{
			description: "AlwaysInvalidType valid",
			config:      []byte("always_invalid"),
		},
		{
			description: "AlwaysValidType valid",
			config:      []byte("always_valid"),
		},
		{
			description: "UTF8Type valid",
			config:      []byte("utf8"),
		},
		{
			description: "MessageTypeType valid",
			config:      []byte("msg_type"),
		},
		{
			description: "SourceType valid",
			config:      []byte("source"),
		},
		{
			description: "DestinationType valid",
			config:      []byte("destination"),
		},
		{
			description: "SimpleResponseRequestTypeType valid",
			config:      []byte("simple_res_req"),
		},
		{
			description: "SimpleEventTypeType valid",
			config:      []byte("simple_event"),
		},
		{
			description: "SpansType valid",
			config:      []byte("spans"),
		},
		{
			description: "Nonexistent type invalid",
			config:      []byte("FOOBAR"),
			invalid:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			var l validatorType

			err := l.UnmarshalText(tc.config)
			assert.NotEmpty(l.getKeys())
			if tc.invalid {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(string(tc.config), l.String())
			}
		})
	}
}

func TestTypeState(t *testing.T) {
	tests := []struct {
		description string
		val         validatorType
		expectedVal string
		invalid     bool
		empty       bool
	}{
		{
			description: "UnknownLevel valid",
			val:         UnknownType,
			expectedVal: "unknown",
			empty:       true,
			invalid:     true,
		},
		{
			description: "AlwaysInvalidType valid",
			val:         AlwaysInvalidType,
			expectedVal: "always_invalid",
		},
		{
			description: "AlwaysValidType valid",
			val:         AlwaysValidType,
			expectedVal: "always_valid",
		},
		{
			description: "UTF8Type valid",
			val:         UTF8Type,
			expectedVal: "utf8",
		},
		{
			description: "MessageTypeType valid",
			val:         MessageTypeType,
			expectedVal: "msg_type",
		},
		{
			description: "SourceType valid",
			val:         SourceType,
			expectedVal: "source",
		},
		{
			description: "DestinationType valid",
			val:         DestinationType,
			expectedVal: "destination",
		},
		{
			description: "SimpleResponseRequestTypeType valid",
			val:         SimpleResponseRequestTypeType,
			expectedVal: "simple_res_req",
		},
		{
			description: "SimpleEventTypeType valid",
			val:         SimpleEventTypeType,
			expectedVal: "simple_event",
		},
		{
			description: "SpansType valid",
			val:         SpansType,
			expectedVal: "spans",
		},
		{
			description: "lastLevel valid",
			val:         lastType,
			expectedVal: "unknown",
			invalid:     true,
		},
		{
			description: "Nonexistent positive validatorLevel invalid",
			val:         lastType + 1,
			expectedVal: "unknown",
			invalid:     true,
		},
		{
			description: "Nonexistent negative validatorLevel invalid",
			val:         UnknownType - 1,
			expectedVal: "unknown",
			invalid:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.expectedVal, tc.val.String())
			assert.Equal(!tc.invalid, tc.val.IsValid())
			assert.Equal(tc.empty, tc.val.IsEmpty())
		})
	}
}
