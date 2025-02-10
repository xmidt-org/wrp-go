// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQOSLevel(t *testing.T) {
	tests := []struct {
		level    QOSLevel
		expected string
	}{
		{
			level:    QOSLow,
			expected: "Low",
		}, {
			level:    QOSMedium,
			expected: "Medium",
		}, {
			level:    QOSHigh,
			expected: "High",
		}, {
			level:    QOSCritical,
			expected: "Critical",
		}, {
			level:    -1,
			expected: "QOSLevel(-1)",
		}, {
			level:    100,
			expected: "QOSLevel(100)",
		}, {
			level:    14,
			expected: "QOSLevel(14)",
		},
	}
	for _, tc := range tests {
		t.Run(strconv.Itoa(int(tc.level)), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.level.String())
		})
	}
}

func TestQOSValue(t *testing.T) {
	t.Run("Level", func(t *testing.T) {
		testCases := []struct {
			value    QOSValue
			expected QOSLevel
			valid    bool
		}{
			{
				value:    QOSLowValue,
				expected: QOSLow,
				valid:    true,
			},
			{
				value:    QOSMediumValue,
				expected: QOSMedium,
				valid:    true,
			},
			{
				value:    QOSHighValue,
				expected: QOSHigh,
				valid:    true,
			},
			{
				value:    QOSCriticalValue,
				expected: QOSCritical,
				valid:    true,
			},
			{
				value:    -1,
				expected: QOSLow,
				valid:    false,
			},
			{
				value:    0,
				expected: QOSLow,
				valid:    true,
			},
			{
				value:    9,
				expected: QOSLow,
				valid:    true,
			},
			{
				value:    24,
				expected: QOSLow,
				valid:    true,
			},
			{
				value:    25,
				expected: QOSMedium,
				valid:    true,
			},
			{
				value:    32,
				expected: QOSMedium,
				valid:    true,
			},
			{
				value:    49,
				expected: QOSMedium,
				valid:    true,
			},
			{
				value:    50,
				expected: QOSHigh,
				valid:    true,
			},
			{
				value:    61,
				expected: QOSHigh,
				valid:    true,
			},
			{
				value:    74,
				expected: QOSHigh,
				valid:    true,
			},
			{
				value:    75,
				expected: QOSCritical,
				valid:    true,
			},
			{
				value:    84,
				expected: QOSCritical,
				valid:    true,
			},
			{
				value:    99,
				expected: QOSCritical,
				valid:    true,
			},
			{
				value:    34876123,
				expected: QOSCritical,
				valid:    false,
			},
		}

		for _, testCase := range testCases {
			t.Run(strconv.Itoa(int(testCase.value)), func(t *testing.T) {
				assert.Equal(t, testCase.expected, testCase.value.Level())
			})
			t.Run(strconv.Itoa(int(testCase.value)), func(t *testing.T) {
				assert.Equal(t, testCase.valid, testCase.value.Valid())
			})
		}
	})
}
