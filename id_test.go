// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
//
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
		prefix       string
		literalID    string
		expectsError bool
	}{
		{
			id:        "MAC:11:22:33:44:55:66",
			expected:  "mac:112233445566",
			prefix:    "mac",
			literalID: "112233445566",
		}, {
			id:        "MAC:11aaBB445566",
			expected:  "mac:11aabb445566",
			prefix:    "mac",
			literalID: "11aabb445566",
		}, {
			id:        "mac:11-aa-BB-44-55-66",
			expected:  "mac:11aabb445566",
			prefix:    "mac",
			literalID: "11aabb445566",
		}, {
			id:        "mac:11,aa,BB,44,55,66",
			expected:  "mac:11aabb445566",
			prefix:    "mac",
			literalID: "11aabb445566",
		}, {
			id:        "uuid:anything Goes!",
			expected:  "uuid:anything Goes!",
			prefix:    "uuid",
			literalID: "anything Goes!",
		}, {
			id:        "dns:anything Goes!",
			expected:  "dns:anything Goes!",
			prefix:    "dns",
			literalID: "anything Goes!",
		}, {
			id:        "serial:1234",
			expected:  "serial:1234",
			prefix:    "serial",
			literalID: "1234",
		}, {
			id:        "mac:11-aa-BB-44-55-66/service",
			expected:  "mac:11aabb445566",
			prefix:    "mac",
			literalID: "11aabb445566",
		}, {
			id:        "mac:11-aa-BB-44-55-66/service/",
			expected:  "mac:11aabb445566",
			prefix:    "mac",
			literalID: "11aabb445566",
		}, {
			id:        "MAC:11-aa-BB-44-55-66/service/ignoreMe",
			expected:  "mac:11aabb445566",
			prefix:    "mac",
			literalID: "11aabb445566",
		}, {
			id:        "mac:11-aa-BB-44-55-66/service/foo/bar",
			expected:  "mac:11aabb445566",
			prefix:    "mac",
			literalID: "11aabb445566",
		}, {
			id:           "invalid:a-BB-44-55",
			expectsError: true,
		}, {
			id:           "mac:11-aa-BB-44-55",
			expectsError: true,
		}, {
			id:           "MAC:invalid45566",
			expectsError: true,
		}, {
			id:           "invalid:random stuff",
			expectsError: true,
		}, {
			id:           "mac:11223344556w",
			expectsError: true,
		}, {
			id:        "mac:481d70187fef",
			expected:  "mac:481d70187fef",
			prefix:    "mac",
			literalID: "481d70187fef",
		}, {
			id:        "mac:481d70187fef/parodus/tag/test0",
			expected:  "mac:481d70187fef",
			prefix:    "mac",
			literalID: "481d70187fef",
		},
	}

	for _, record := range testData {
		t.Run(record.id, func(t *testing.T) {
			id, err := ParseDeviceID(record.id)
			assert.Equal(record.expected, id)
			assert.Equal(record.expectsError, err != nil)
			assert.Equal([]byte(record.expected), id.Bytes())
			assert.Equal(record.prefix, id.Prefix())
			assert.Equal(record.literalID, id.ID())
		})
	}
}

func TestParseLocator(t *testing.T) {
	tests := []struct {
		description string
		locator     string
		want        Locator
		str         string
		expectedErr error
	}{
		{
			description: "cpe locator",
			locator:     "mac:112233445566",
			str:         "mac:112233445566",
			want: Locator{
				Scheme:    SchemeMAC,
				Authority: "112233445566",
				ID:        "mac:112233445566",
			},
		}, {
			description: "cpe locator ensure lowercase",
			locator:     "Mac:112233445566",
			str:         "mac:112233445566",
			want: Locator{
				Scheme:    SchemeMAC,
				Authority: "112233445566",
				ID:        "mac:112233445566",
			},
		}, {
			description: "locator with service",
			locator:     "DNS:foo.bar.com/service",
			str:         "dns:foo.bar.com/service",
			want: Locator{
				Scheme:    SchemeDNS,
				Authority: "foo.bar.com",
				Service:   "service",
			},
		}, {
			description: "locator with service everything",
			locator:     "event:something/service/ignored",
			str:         "event:something/service/ignored",
			want: Locator{
				Scheme:    SchemeEvent,
				Authority: "something",
				Service:   "service",
				Ignored:   "/ignored",
			},
		}, {
			description: "self locator with service",
			locator:     "SELF:/service",
			str:         "self:/service",
			want: Locator{
				Scheme:  SchemeSelf,
				Service: "service",
				ID:      "self:",
			},
		}, {
			description: "self locator with service everything",
			locator:     "self:/service/ignored",
			str:         "self:/service/ignored",
			want: Locator{
				Scheme:  SchemeSelf,
				Service: "service",
				Ignored: "/ignored",
				ID:      "self:",
			},
		},

		// Validate all the schemes
		{
			description: "dns scheme",
			locator:     "dns:foo.bar.com",
			str:         "dns:foo.bar.com",
			want: Locator{
				Scheme:    SchemeDNS,
				Authority: "foo.bar.com",
			},
		}, {
			description: "event scheme",
			locator:     "event:targetedEvent",
			str:         "event:targetedEvent",
			want: Locator{
				Scheme:    SchemeEvent,
				Authority: "targetedEvent",
			},
		}, {
			description: "mac scheme",
			locator:     "mac:112233445566",
			str:         "mac:112233445566",
			want: Locator{
				Scheme:    SchemeMAC,
				Authority: "112233445566",
				ID:        "mac:112233445566",
			},
		}, {
			description: "serial scheme",
			locator:     "serial:AsdfSerial",
			str:         "serial:AsdfSerial",
			want: Locator{
				Scheme:    SchemeSerial,
				Authority: "AsdfSerial",
				ID:        "serial:AsdfSerial",
			},
		}, {
			description: "uuid scheme",
			locator:     "uuid:bbee1f69-2f64-4aa9-a422-27d68b40b152",
			str:         "uuid:bbee1f69-2f64-4aa9-a422-27d68b40b152",
			want: Locator{
				Scheme:    SchemeUUID,
				Authority: "bbee1f69-2f64-4aa9-a422-27d68b40b152",
				ID:        "uuid:bbee1f69-2f64-4aa9-a422-27d68b40b152",
			},
		}, {
			description: "self scheme",
			locator:     "self:",
			str:         "self:",
			want: Locator{
				Scheme: SchemeSelf,
				ID:     "self:",
			},
		},

		// Validate invalid locators are caught
		{
			description: "empty",
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "invalid scheme",
			locator:     "invalid:foo.bar.com",
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "invalid mac scheme",
			locator:     "mac:112invalid66",
			expectedErr: ErrorInvalidDeviceName,
		}, {
			description: "invalid self scheme",
			locator:     "self:anything",
			expectedErr: ErrorInvalidDeviceName,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			got, err := ParseLocator(tc.locator)

			assert.ErrorIs(err, tc.expectedErr)
			assert.Equal(tc.want, got)

			l := Locator{}
			if tc.want != l {
				assert.Equal(tc.str, got.String())
			}
		})
	}
}

func TestLocatorDeviceID(t *testing.T) {
	assert := assert.New(t)

	l, err := ParseLocator("mac:112233445566")
	assert.NoError(err)

	assert.True(l.HasDeviceID())
	assert.NotEqual(l.ID, "")
}
