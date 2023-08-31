// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDeviceID(t *testing.T) {
	assert := assert.New(t)
	testData := []struct {
		id           string
		expected     DeviceID
		expectsError bool
	}{
		{"MAC:11:22:33:44:55:66", "mac:112233445566", false},
		{"MAC:11aaBB445566", "mac:11aabb445566", false},
		{"mac:11-aa-BB-44-55-66", "mac:11aabb445566", false},
		{"mac:11,aa,BB,44,55,66", "mac:11aabb445566", false},
		{"uuid:anything Goes!", "uuid:anything Goes!", false},
		{"dns:anything Goes!", "dns:anything Goes!", false},
		{"serial:1234", "serial:1234", false},
		{"mac:11-aa-BB-44-55-66/service", "mac:11aabb445566", false},
		{"mac:11-aa-BB-44-55-66/service/", "mac:11aabb445566", false},
		{"mac:11-aa-BB-44-55-66/service/ignoreMe", "mac:11aabb445566", false},
		{"mac:11-aa-BB-44-55-66/service/foo/bar", "mac:11aabb445566", false},
		{"invalid:a-BB-44-55", "", true},
		{"mac:11-aa-BB-44-55", "", true},
		{"MAC:invalid45566", "", true},
		{"mac:481d70187fef", "mac:481d70187fef", false},
		{"mac:481d70187fef/parodus/tag/test0", "mac:481d70187fef", false},
	}

	for _, record := range testData {
		t.Run(record.id, func(t *testing.T) {
			id, err := ParseDeviceID(record.id)
			assert.Equal(record.expected, id)
			assert.Equal(record.expectsError, err != nil)
			assert.Equal([]byte(record.expected), id.Bytes())
		})
	}
}
