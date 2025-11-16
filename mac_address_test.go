// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		description string
		input       string
		expected    string
		expectError bool
	}{
		// Already normalized cases (should return input unchanged)
		{
			description: "already normalized lowercase",
			input:       "aabbccddeeff",
			expected:    "aabbccddeeff",
		},
		{
			description: "already normalized with numbers",
			input:       "112233445566",
			expected:    "112233445566",
		},
		{
			description: "already normalized mixed",
			input:       "1a2b3c4d5e6f",
			expected:    "1a2b3c4d5e6f",
		},

		// Delimiter removal cases
		{
			description: "colon delimiter",
			input:       "11:22:33:44:55:66",
			expected:    "112233445566",
		},
		{
			description: "dash delimiter",
			input:       "11-22-33-44-55-66",
			expected:    "112233445566",
		},
		{
			description: "dot delimiter",
			input:       "1122.3344.5566",
			expected:    "112233445566",
		},
		{
			description: "comma delimiter",
			input:       "11,22,33,44,55,66",
			expected:    "112233445566",
		},
		{
			description: "mixed delimiters",
			input:       "11:22-33.44,55:66",
			expected:    "112233445566",
		},

		// Case normalization
		{
			description: "uppercase letters",
			input:       "AABBCCDDEEFF",
			expected:    "aabbccddeeff",
		},
		{
			description: "mixed case",
			input:       "AaBbCcDdEeFf",
			expected:    "aabbccddeeff",
		},
		{
			description: "uppercase with delimiters",
			input:       "AA:BB:CC:DD:EE:FF",
			expected:    "aabbccddeeff",
		},
		{
			description: "mixed case with delimiters",
			input:       "Aa:Bb:Cc:Dd:Ee:Ff",
			expected:    "aabbccddeeff",
		},

		// Complex valid cases
		{
			description: "complex format 1",
			input:       "1A-2B-3C-4D-5E-6F",
			expected:    "1a2b3c4d5e6f",
		},
		{
			description: "complex format 2",
			input:       "1a2b.3c4d.5e6f",
			expected:    "1a2b3c4d5e6f",
		},

		// Error cases - invalid characters
		{
			description: "invalid character g",
			input:       "112233445566g",
			expectError: true,
		},
		{
			description: "invalid character z",
			input:       "11:22:33:44:55:zz",
			expectError: true,
		},
		{
			description: "invalid character space",
			input:       "11 22 33 44 55 66",
			expectError: true,
		},
		{
			description: "invalid character underscore",
			input:       "11_22_33_44_55_66",
			expectError: true,
		},
		{
			description: "invalid character slash",
			input:       "11/22/33/44/55/66",
			expectError: true,
		},

		// Error cases - invalid length
		{
			description: "too short - 11 chars",
			input:       "11223344556",
			expectError: true,
		},
		{
			description: "too short - 10 chars",
			input:       "1122334455",
			expectError: true,
		},
		{
			description: "too short with delimiters",
			input:       "11:22:33:44:55",
			expectError: true,
		},
		{
			description: "too long - 13 chars",
			input:       "1122334455667",
			expectError: true,
		},
		{
			description: "too long - 14 chars",
			input:       "11223344556677",
			expectError: true,
		},
		{
			description: "empty string",
			input:       "",
			expectError: true,
		},
		{
			description: "only delimiters",
			input:       ":::::",
			expectError: true,
		},

		// Edge cases
		{
			description: "all zeros",
			input:       "000000000000",
			expected:    "000000000000",
		},
		{
			description: "all zeros with delimiters",
			input:       "00:00:00:00:00:00",
			expected:    "000000000000",
		},
		{
			description: "all F's",
			input:       "ffffffffffff",
			expected:    "ffffffffffff",
		},
		{
			description: "all F's uppercase",
			input:       "FFFFFFFFFFFF",
			expected:    "ffffffffffff",
		},
		{
			description: "all F's with delimiters",
			input:       "FF:FF:FF:FF:FF:FF",
			expected:    "ffffffffffff",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			result, err := normalizeMAC(tc.input)

			if tc.expectError {
				assert.Error(err)
				assert.True(errors.Is(err, ErrorInvalidDeviceName))
				assert.Empty(result)
			} else {
				assert.NoError(err)
				assert.Equal(tc.expected, result)
			}
		})
	}
}

func TestIsLowerHexString(t *testing.T) {
	tests := []struct {
		description string
		input       string
		expected    bool
	}{
		// Valid lowercase hex strings
		{
			description: "all lowercase hex",
			input:       "abcdef",
			expected:    true,
		},
		{
			description: "all numbers",
			input:       "0123456789",
			expected:    true,
		},
		{
			description: "mixed numbers and lowercase",
			input:       "1a2b3c4d5e6f",
			expected:    true,
		},
		{
			description: "valid MAC address",
			input:       "112233445566",
			expected:    true,
		},
		{
			description: "all zeros",
			input:       "000000000000",
			expected:    true,
		},
		{
			description: "all f's lowercase",
			input:       "ffffffffffff",
			expected:    true,
		},
		{
			description: "empty string",
			input:       "",
			expected:    true, // empty string has no invalid characters
		},

		// Invalid - uppercase letters
		{
			description: "uppercase A",
			input:       "Aabbccddeeff",
			expected:    false,
		},
		{
			description: "all uppercase",
			input:       "ABCDEF",
			expected:    false,
		},
		{
			description: "mixed case",
			input:       "AaBbCc",
			expected:    false,
		},

		// Invalid - non-hex characters
		{
			description: "contains g",
			input:       "abcdefg",
			expected:    false,
		},
		{
			description: "contains z",
			input:       "11223z",
			expected:    false,
		},
		{
			description: "contains delimiter colon",
			input:       "11:22:33",
			expected:    false,
		},
		{
			description: "contains delimiter dash",
			input:       "11-22-33",
			expected:    false,
		},
		{
			description: "contains space",
			input:       "112233 445566",
			expected:    false,
		},
		{
			description: "contains special char",
			input:       "112233!445566",
			expected:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			result := isLowerHexString(tc.input)
			assert.Equal(tc.expected, result)
		})
	}
}

// TestNormalizeMACNoAllocation verifies that normalizeMAC doesn't allocate
// when the input is already normalized.
func TestNormalizeMACNoAllocation(t *testing.T) {
	input := "112233445566"
	result, err := normalizeMAC(input)

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(input, result)

	// Verify zero-copy behavior - the result should be the same string as input
	// (not just equal content, but the exact same underlying string)
	// This is implementation detail testing, but important for performance
	inputPtr := (*[2]uintptr)(unsafe.Pointer(&input))
	resultPtr := (*[2]uintptr)(unsafe.Pointer(&result))

	// Compare data pointers - they should be the same for zero-copy
	assert.Equal(inputPtr[0], resultPtr[0], "normalized MAC should share data with input (zero-copy)")
}
