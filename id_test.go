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
		zeroCopy     bool
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
			zeroCopy:  true,
		}, {
			id:        "dns:anything Goes!",
			expected:  "dns:anything Goes!",
			prefix:    "dns",
			literalID: "anything Goes!",
			zeroCopy:  true,
		}, {
			id:        "serial:1234",
			expected:  "serial:1234",
			prefix:    "serial",
			literalID: "1234",
			zeroCopy:  true,
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
			zeroCopy:  true,
		}, {
			id:        "mac:481d70187fef/parodus/tag/test0",
			expected:  "mac:481d70187fef",
			prefix:    "mac",
			literalID: "481d70187fef",
			zeroCopy:  true,
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

			if !record.expectsError {
				if record.zeroCopy {
					// When zero-copy is expected, the string backing the ID should be the same as input
					got := string(id)
					have := record.id
					assert.Exactly(have[0:len(got)], got, "zero-copy: strings should be exactly the same")
				}
			}
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
			locator:     "DNS:foo.bar.com/service/ignored/really/really/ignored",
			str:         "dns:foo.bar.com/service/ignored/really/really/ignored",
			want: Locator{
				Scheme:    SchemeDNS,
				Authority: "foo.bar.com",
				Ignored:   "/service/ignored/really/really/ignored",
			},
		}, {
			description: "locator with service everything",
			locator:     "event:event_name/ignored/really/really/ignored",
			str:         "event:event_name/ignored/really/really/ignored",
			want: Locator{
				Scheme:    SchemeEvent,
				Authority: "event_name",
				Service:   "",
				Ignored:   "/ignored/really/really/ignored",
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
			locator:     "self:/service/ignored/really/really/ignored",
			str:         "self:/service/ignored/really/really/ignored",
			want: Locator{
				Scheme:  SchemeSelf,
				Service: "service",
				Ignored: "/ignored/really/really/ignored",
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
			description: "event scheme (with spaces)",
			locator:     "event:   targetedEvent     ",
			str:         "event:targetedEvent",
			want: Locator{
				Scheme:    SchemeEvent,
				Authority: "targetedEvent",
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
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "invalid event scheme (no authority)",
			locator:     "event:/anything",
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "invalid event scheme (no authority and with spaces)",
			locator:     "event:    /anything",
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "invalid dns scheme (no authority)",
			locator:     "dns:/anything",
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "invalid dns scheme (no authority and with spaces)",
			locator:     "dns:      /anything",
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "invalid uuid scheme (no authority)",
			locator:     "uuid:/anything",
			expectedErr: ErrorInvalidDeviceName,
		}, {
			description: "invalid uuid scheme (no authority and with spaces)",
			locator:     "uuid:      /anything",
			expectedErr: ErrorInvalidDeviceName,
		}, {
			description: "invalid serial scheme (no authority)",
			locator:     "serial:/anything",
			expectedErr: ErrorInvalidDeviceName,
		}, {
			description: "invalid serial scheme (no authority and with spaces)",
			locator:     "serial:      /anything",
			expectedErr: ErrorInvalidDeviceName,
		}, {
			description: "invalid event scheme (no service)",
			locator:     "event:/invalid",
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
	assert.Equal(l.ID, DeviceID("mac:112233445566"))

	alt := l.ID.AsLocator()
	assert.Equal(l, alt)
}

func TestLocatorIs(t *testing.T) {
	self := DeviceID("self:")

	tests := []struct {
		description string
		target      DeviceID
		list        []DeviceID
		expected    bool
	}{
		{
			description: "self",
			target:      self,
			list: []DeviceID{
				self,
				DeviceID("mac:112233445566"),
			},
			expected: true,
		}, {
			description: "self, not in list",
			target:      self,
			list: []DeviceID{
				DeviceID("mac:112233445566"),
			},
			expected: false,
		}, {
			description: "not self",
			target:      DeviceID("mac:112233445566"),
			list: []DeviceID{
				self,
				DeviceID("mac:112233445566"),
			},
			expected: true,
		}, {
			description: "not self, not in list",
			target:      DeviceID("mac:112233445566"),
			list: []DeviceID{
				DeviceID("dns:example.com"),
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.expected, tc.target.Is(tc.list...))

			l := tc.target.AsLocator()
			assert.Equal(tc.expected, l.Is(tc.list...))
		})
	}
}

func TestValidateLocator(t *testing.T) {
	tests := []struct {
		description string
		in          Locator
		err         bool
	}{
		{
			description: "valid mac",
			in: Locator{
				Scheme:    SchemeMAC,
				Authority: "112233445566",
				ID:        DeviceID("mac:112233445566"),
			},
		}, {
			description: "valid dns",
			in: Locator{
				Scheme:    SchemeDNS,
				Authority: "example.com",
			},
		}, {
			description: "valid event",
			in: Locator{
				Scheme:    SchemeEvent,
				Authority: "event_name",
			},
		}, {
			description: "valid uuid",
			in: Locator{
				Scheme:    SchemeUUID,
				Authority: "bbee1f69-2f64-4aa9-a422-27d68b40b152",
				ID:        DeviceID("uuid:bbee1f69-2f64-4aa9-a422-27d68b40b152"),
			},
		}, {
			description: "valid serial",
			in: Locator{
				Scheme:    SchemeSerial,
				Authority: "AsdfSerial",
				ID:        DeviceID("serial:AsdfSerial"),
			},
		}, {
			description: "valid self",
			in: Locator{
				Scheme:  SchemeSelf,
				ID:      DeviceID("self:"),
				Service: "service",
			},
		}, {
			description: "invalid self",
			in: Locator{
				Scheme:    SchemeSelf,
				Authority: "example.com",
				ID:        DeviceID("self:"),
			},
			err: true,
		}, {
			description: "invalid self",
			in: Locator{
				Scheme: SchemeSelf,
			},
			err: true,
		}, {
			description: "invalid scheme",
			in: Locator{
				Scheme: "invalid",
			},
			err: true,
		}, {
			description: "invalid mac",
			in: Locator{
				Scheme:    SchemeMAC,
				Authority: "112invalid66",
				ID:        DeviceID("mac:112invalid66"),
			},
			err: true,
		}, {
			description: "invalid mac id",
			in: Locator{
				Scheme:    SchemeMAC,
				Authority: "112233445566",
				ID:        DeviceID("mac:665544332211"),
			},
			err: true,
		}, {
			description: "invalid mac no authority",
			in: Locator{
				Scheme: SchemeMAC,
				ID:     DeviceID("mac:665544332211"),
			},
			err: true,
		}, {
			description: "invalid event no authority",
			in: Locator{
				Scheme: SchemeEvent,
			},
			err: true,
		}, {
			description: "invalid event service is set",
			in: Locator{
				Scheme:    SchemeEvent,
				Authority: "event_name",
				Service:   "service",
			},
			err: true,
		}, {
			description: "invalid event id is set",
			in: Locator{
				Scheme:    SchemeEvent,
				Authority: "event_name",
				ID:        DeviceID("event:event_name"),
			},
			err: true,
		}, {
			description: "invalid dns id is set",
			in: Locator{
				Scheme:    SchemeDNS,
				Authority: "example.com",
				ID:        DeviceID("dns:example.com"),
			},
			err: true,
		}, {
			description: "invalid dns service is set",
			in: Locator{
				Scheme:    SchemeDNS,
				Authority: "example.com",
				Service:   "service",
			},
			err: true,
		}, {
			description: "invalid service contains /",
			in: Locator{
				Scheme:    SchemeSerial,
				Authority: "AsdfSerial",
				Service:   "service/with/slashes",
				ID:        DeviceID("serial:AsdfSerial"),
			},
			err: true,
		},
		// Additional MAC validation tests
		{
			description: "invalid mac - wrong prefix in ID",
			in: Locator{
				Scheme:    SchemeMAC,
				Authority: "112233445566",
				ID:        DeviceID("uuid:112233445566"),
			},
			err: true,
		},
		// UUID validation tests
		{
			description: "invalid uuid - empty authority",
			in: Locator{
				Scheme: SchemeUUID,
				ID:     DeviceID("uuid:something"),
			},
			err: true,
		}, {
			description: "invalid uuid - ID doesn't match authority",
			in: Locator{
				Scheme:    SchemeUUID,
				Authority: "bbee1f69-2f64-4aa9-a422-27d68b40b152",
				ID:        DeviceID("uuid:different-uuid"),
			},
			err: true,
		}, {
			description: "invalid uuid - wrong prefix in ID",
			in: Locator{
				Scheme:    SchemeUUID,
				Authority: "bbee1f69-2f64-4aa9-a422-27d68b40b152",
				ID:        DeviceID("mac:bbee1f69-2f64-4aa9-a422-27d68b40b152"),
			},
			err: true,
		},
		// Serial validation tests
		{
			description: "invalid serial - empty authority",
			in: Locator{
				Scheme: SchemeSerial,
				ID:     DeviceID("serial:something"),
			},
			err: true,
		}, {
			description: "invalid serial - ID doesn't match authority",
			in: Locator{
				Scheme:    SchemeSerial,
				Authority: "ABC123",
				ID:        DeviceID("serial:XYZ789"),
			},
			err: true,
		}, {
			description: "invalid serial - wrong prefix in ID",
			in: Locator{
				Scheme:    SchemeSerial,
				Authority: "ABC123",
				ID:        DeviceID("uuid:ABC123"),
			},
			err: true,
		},
		// DNS validation tests
		{
			description: "invalid dns - empty authority",
			in: Locator{
				Scheme: SchemeDNS,
			},
			err: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			err := tc.in.Validate()
			assert.Equal(tc.err, err != nil)
		})
	}
}
