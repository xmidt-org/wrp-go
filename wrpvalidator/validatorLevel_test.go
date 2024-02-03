// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelUnmarshalling(t *testing.T) {
	tests := []struct {
		description string
		config      []byte
		invalid     bool
	}{
		{
			description: "UnknownLevel success",
			config:      []byte("unknown"),
		},
		{
			description: "InfoLevel success",
			config:      []byte("info"),
		},
		{
			description: "WarningLevel success",
			config:      []byte("warning"),
		},
		{
			description: "ErrorLevel success",
			config:      []byte("error"),
		},
		{
			description: "Nonexistent level success",
			config:      []byte("FOOBAR"),
			invalid:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			var l validatorLevel

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

func TestLevelState(t *testing.T) {
	tests := []struct {
		description string
		val         validatorLevel
		expectedVal string
		invalid     bool
		empty       bool
	}{
		{
			description: "unknown valid",
			val:         UnknownLevel,
			expectedVal: "unknown",
			empty:       true,
			invalid:     true,
		},
		{
			description: "InfoLevel valid",
			val:         InfoLevel,
			expectedVal: "info",
		},
		{
			description: "WarningLevel valid",
			val:         WarningLevel,
			expectedVal: "warning",
		},
		{
			description: "ErrorLevel valid",
			val:         ErrorLevel,
			expectedVal: "error",
		},
		{
			description: "lastLevel valid",
			val:         lastLevel,
			expectedVal: "unknown",
			invalid:     true,
		},
		{
			description: "Nonexistent positive validatorLevel invalid",
			val:         lastLevel + 1,
			expectedVal: "unknown",
			invalid:     true,
		},
		{
			description: "Nonexistent negative validatorLevel invalid",
			val:         UnknownLevel - 1,
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
