package wrp

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQOSLevel(t *testing.T) {
	for _, v := range []QOSLevel{QOSLow, QOSMedium, QOSHigh, QOSCritical} {
		t.Run(strconv.Itoa(int(v)), func(t *testing.T) {
			assert.NotEmpty(t, v.String())
		})
	}
}

func TestQOSValue(t *testing.T) {
	t.Run("Level", func(t *testing.T) {
		testCases := []struct {
			value    QOSValue
			expected QOSLevel
		}{
			{
				value:    -1,
				expected: QOSLow,
			},
			{
				value:    0,
				expected: QOSLow,
			},
			{
				value:    9,
				expected: QOSLow,
			},
			{
				value:    24,
				expected: QOSLow,
			},
			{
				value:    25,
				expected: QOSMedium,
			},
			{
				value:    32,
				expected: QOSMedium,
			},
			{
				value:    49,
				expected: QOSMedium,
			},
			{
				value:    50,
				expected: QOSHigh,
			},
			{
				value:    61,
				expected: QOSHigh,
			},
			{
				value:    74,
				expected: QOSHigh,
			},
			{
				value:    75,
				expected: QOSCritical,
			},
			{
				value:    84,
				expected: QOSCritical,
			},
			{
				value:    99,
				expected: QOSCritical,
			},
			{
				value:    34876123,
				expected: QOSCritical,
			},
		}

		for _, testCase := range testCases {
			t.Run(strconv.Itoa(int(testCase.value)), func(t *testing.T) {
				assert.Equal(t, testCase.expected, testCase.value.Level())
			})
		}
	})
}
