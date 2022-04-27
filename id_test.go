/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package wrp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseID(t *testing.T) {
	assert := assert.New(t)
	testData := []struct {
		id           string
		expected     ID
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
			id, err := ParseID(record.id)
			assert.Equal(record.expected, id)
			assert.Equal(record.expectsError, err != nil)
			assert.Equal([]byte(record.expected), id.Bytes())
		})
	}
}
