// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetScheme(t *testing.T) {
	tests := []struct {
		description    string
		input          string
		expectedScheme string
		expectedRest   string
		expectedAlter  bool
	}{
		// Valid schemes - lowercase (no alteration)
		{
			description:    "mac scheme lowercase",
			input:          "mac:112233445566",
			expectedScheme: SchemeMAC,
			expectedRest:   "112233445566",
			expectedAlter:  false,
		},
		{
			description:    "uuid scheme lowercase",
			input:          "uuid:12345",
			expectedScheme: SchemeUUID,
			expectedRest:   "12345",
			expectedAlter:  false,
		},
		{
			description:    "dns scheme lowercase",
			input:          "dns:example.com",
			expectedScheme: SchemeDNS,
			expectedRest:   "example.com",
			expectedAlter:  false,
		},
		{
			description:    "serial scheme lowercase",
			input:          "serial:ABC123",
			expectedScheme: SchemeSerial,
			expectedRest:   "ABC123",
			expectedAlter:  false,
		},
		{
			description:    "self scheme lowercase",
			input:          "self:",
			expectedScheme: SchemeSelf,
			expectedRest:   "",
			expectedAlter:  false,
		},
		{
			description:    "event scheme lowercase",
			input:          "event:device-status",
			expectedScheme: SchemeEvent,
			expectedRest:   "device-status",
			expectedAlter:  false,
		},

		// Valid schemes - uppercase (with alteration)
		{
			description:    "MAC scheme uppercase",
			input:          "MAC:112233445566",
			expectedScheme: SchemeMAC,
			expectedRest:   "112233445566",
			expectedAlter:  true,
		},
		{
			description:    "UUID scheme uppercase",
			input:          "UUID:12345",
			expectedScheme: SchemeUUID,
			expectedRest:   "12345",
			expectedAlter:  true,
		},
		{
			description:    "DNS scheme uppercase",
			input:          "DNS:example.com",
			expectedScheme: SchemeDNS,
			expectedRest:   "example.com",
			expectedAlter:  true,
		},
		{
			description:    "SERIAL scheme uppercase",
			input:          "SERIAL:ABC123",
			expectedScheme: SchemeSerial,
			expectedRest:   "ABC123",
			expectedAlter:  true,
		},
		{
			description:    "SELF scheme uppercase",
			input:          "SELF:",
			expectedScheme: SchemeSelf,
			expectedRest:   "",
			expectedAlter:  true,
		},
		{
			description:    "EVENT scheme uppercase",
			input:          "EVENT:device-status",
			expectedScheme: SchemeEvent,
			expectedRest:   "device-status",
			expectedAlter:  true,
		},

		// Valid schemes - mixed case (with alteration)
		{
			description:    "Mac scheme mixed case",
			input:          "Mac:112233445566",
			expectedScheme: SchemeMAC,
			expectedRest:   "112233445566",
			expectedAlter:  true,
		},
		{
			description:    "Uuid scheme mixed case",
			input:          "Uuid:12345",
			expectedScheme: SchemeUUID,
			expectedRest:   "12345",
			expectedAlter:  true,
		},
		{
			description:    "dNs scheme mixed case",
			input:          "dNs:example.com",
			expectedScheme: SchemeDNS,
			expectedRest:   "example.com",
			expectedAlter:  true,
		},
		{
			description:    "Serial scheme mixed case",
			input:          "Serial:ABC123",
			expectedScheme: SchemeSerial,
			expectedRest:   "ABC123",
			expectedAlter:  true,
		},
		{
			description:    "Self scheme mixed case",
			input:          "Self:",
			expectedScheme: SchemeSelf,
			expectedRest:   "",
			expectedAlter:  true,
		},
		{
			description:    "Event scheme mixed case",
			input:          "Event:device-status",
			expectedScheme: SchemeEvent,
			expectedRest:   "device-status",
			expectedAlter:  true,
		},

		// Schemes with additional content after
		{
			description:    "mac with path",
			input:          "mac:112233445566/service/foo",
			expectedScheme: SchemeMAC,
			expectedRest:   "112233445566/service/foo",
			expectedAlter:  false,
		},
		{
			description:    "dns with path",
			input:          "dns:example.com/service/ignored",
			expectedScheme: SchemeDNS,
			expectedRest:   "example.com/service/ignored",
			expectedAlter:  false,
		},
		{
			description:    "self with service",
			input:          "self:/service",
			expectedScheme: SchemeSelf,
			expectedRest:   "/service",
			expectedAlter:  false,
		},

		// 's' ambiguity - 'self' should match before 'serial'
		{
			description:    "self matches before serial",
			input:          "self:something",
			expectedScheme: SchemeSelf,
			expectedRest:   "something",
			expectedAlter:  false,
		},
		{
			description:    "SELF matches before SERIAL",
			input:          "SELF:something",
			expectedScheme: SchemeSelf,
			expectedRest:   "something",
			expectedAlter:  true,
		},

		// Invalid schemes
		{
			description:    "unknown scheme",
			input:          "unknown:value",
			expectedScheme: "",
			expectedRest:   "unknown:value",
			expectedAlter:  false,
		},
		{
			description:    "invalid starting with m",
			input:          "magic:value",
			expectedScheme: "",
			expectedRest:   "magic:value",
			expectedAlter:  false,
		},
		{
			description:    "invalid starting with s",
			input:          "service:value",
			expectedScheme: "",
			expectedRest:   "service:value",
			expectedAlter:  false,
		},
		{
			description:    "invalid starting with d",
			input:          "device:value",
			expectedScheme: "",
			expectedRest:   "device:value",
			expectedAlter:  false,
		},
		{
			description:    "invalid starting with e",
			input:          "error:value",
			expectedScheme: "",
			expectedRest:   "error:value",
			expectedAlter:  false,
		},
		{
			description:    "invalid starting with u",
			input:          "user:value",
			expectedScheme: "",
			expectedRest:   "user:value",
			expectedAlter:  false,
		},

		// Missing colon
		{
			description:    "mac without colon",
			input:          "mac112233445566",
			expectedScheme: "",
			expectedRest:   "mac112233445566",
			expectedAlter:  false,
		},
		{
			description:    "dns without colon",
			input:          "dnsexample.com",
			expectedScheme: "",
			expectedRest:   "dnsexample.com",
			expectedAlter:  false,
		},

		// Too short
		{
			description:    "too short for mac",
			input:          "ma:",
			expectedScheme: "",
			expectedRest:   "ma:",
			expectedAlter:  false,
		},
		{
			description:    "too short for dns",
			input:          "dn:",
			expectedScheme: "",
			expectedRest:   "dn:",
			expectedAlter:  false,
		},
		{
			description:    "single character",
			input:          "m:",
			expectedScheme: "",
			expectedRest:   "m:",
			expectedAlter:  false,
		},

		// Empty and edge cases
		{
			description:    "empty string",
			input:          "",
			expectedScheme: "",
			expectedRest:   "",
			expectedAlter:  false,
		},
		{
			description:    "just colon",
			input:          ":",
			expectedScheme: "",
			expectedRest:   ":",
			expectedAlter:  false,
		},
		{
			description:    "colon at start",
			input:          ":value",
			expectedScheme: "",
			expectedRest:   ":value",
			expectedAlter:  false,
		},

		// Schemes with empty authority
		{
			description:    "mac with empty authority",
			input:          "mac:",
			expectedScheme: SchemeMAC,
			expectedRest:   "",
			expectedAlter:  false,
		},
		{
			description:    "dns with empty authority",
			input:          "dns:",
			expectedScheme: SchemeDNS,
			expectedRest:   "",
			expectedAlter:  false,
		},

		{
			description:    "empty string",
			input:          "",
			expectedScheme: "",
			expectedRest:   "",
			expectedAlter:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			scheme, rest, altered := getScheme(tc.input)

			assert.Equal(tc.expectedScheme, scheme, "scheme mismatch")
			assert.Equal(tc.expectedRest, rest, "rest mismatch")
			assert.Equal(tc.expectedAlter, altered, "altered flag mismatch")
		})
	}
}

func TestIsScheme(t *testing.T) {
	tests := []struct {
		description   string
		input         string
		scheme        string
		expectedMatch bool
		expectedAlter bool
	}{
		// Exact matches (lowercase)
		{
			description:   "exact match mac",
			input:         "mac:value",
			scheme:        SchemeMAC,
			expectedMatch: true,
			expectedAlter: false,
		},
		{
			description:   "exact match uuid",
			input:         "uuid:value",
			scheme:        SchemeUUID,
			expectedMatch: true,
			expectedAlter: false,
		},
		{
			description:   "exact match dns",
			input:         "dns:value",
			scheme:        SchemeDNS,
			expectedMatch: true,
			expectedAlter: false,
		},
		{
			description:   "exact match serial",
			input:         "serial:value",
			scheme:        SchemeSerial,
			expectedMatch: true,
			expectedAlter: false,
		},
		{
			description:   "exact match self",
			input:         "self:value",
			scheme:        SchemeSelf,
			expectedMatch: true,
			expectedAlter: false,
		},
		{
			description:   "exact match event",
			input:         "event:value",
			scheme:        SchemeEvent,
			expectedMatch: true,
			expectedAlter: false,
		},

		// Uppercase matches (with alteration)
		{
			description:   "uppercase MAC",
			input:         "MAC:value",
			scheme:        SchemeMAC,
			expectedMatch: true,
			expectedAlter: true,
		},
		{
			description:   "uppercase UUID",
			input:         "UUID:value",
			scheme:        SchemeUUID,
			expectedMatch: true,
			expectedAlter: true,
		},
		{
			description:   "uppercase DNS",
			input:         "DNS:value",
			scheme:        SchemeDNS,
			expectedMatch: true,
			expectedAlter: true,
		},
		{
			description:   "uppercase SERIAL",
			input:         "SERIAL:value",
			scheme:        SchemeSerial,
			expectedMatch: true,
			expectedAlter: true,
		},
		{
			description:   "uppercase SELF",
			input:         "SELF:value",
			scheme:        SchemeSelf,
			expectedMatch: true,
			expectedAlter: true,
		},
		{
			description:   "uppercase EVENT",
			input:         "EVENT:value",
			scheme:        SchemeEvent,
			expectedMatch: true,
			expectedAlter: true,
		},

		// Mixed case matches (with alteration)
		{
			description:   "mixed case Mac",
			input:         "Mac:value",
			scheme:        SchemeMAC,
			expectedMatch: true,
			expectedAlter: true,
		},
		{
			description:   "mixed case Uuid",
			input:         "Uuid:value",
			scheme:        SchemeUUID,
			expectedMatch: true,
			expectedAlter: true,
		},
		{
			description:   "mixed case dNs",
			input:         "dNs:value",
			scheme:        SchemeDNS,
			expectedMatch: true,
			expectedAlter: true,
		},
		{
			description:   "mixed case SeRiAl",
			input:         "SeRiAl:value",
			scheme:        SchemeSerial,
			expectedMatch: true,
			expectedAlter: true,
		},

		// Non-matches
		{
			description:   "wrong scheme",
			input:         "dns:value",
			scheme:        SchemeMAC,
			expectedMatch: false,
			expectedAlter: false,
		},
		{
			description:   "prefix but not exact",
			input:         "maci:value",
			scheme:        SchemeMAC,
			expectedMatch: false,
			expectedAlter: false,
		},
		{
			description:   "missing colon",
			input:         "macvalue",
			scheme:        SchemeMAC,
			expectedMatch: false,
			expectedAlter: false,
		},
		{
			description:   "too short - no colon",
			input:         "mac",
			scheme:        SchemeMAC,
			expectedMatch: false,
			expectedAlter: false,
		},
		{
			description:   "too short - only 3 chars",
			input:         "ma:",
			scheme:        SchemeMAC,
			expectedMatch: false,
			expectedAlter: false,
		},
		{
			description:   "empty string",
			input:         "",
			scheme:        SchemeMAC,
			expectedMatch: false,
			expectedAlter: false,
		},
		{
			description:   "colon in wrong position",
			input:         "m:ac",
			scheme:        SchemeMAC,
			expectedMatch: false,
			expectedAlter: false,
		},

		// Edge cases
		{
			description:   "scheme with empty value",
			input:         "mac:",
			scheme:        SchemeMAC,
			expectedMatch: true,
			expectedAlter: false,
		},
		{
			description:   "scheme with slash immediately",
			input:         "mac:/",
			scheme:        SchemeMAC,
			expectedMatch: true,
			expectedAlter: false,
		},
		{
			description:   "scheme with special chars after",
			input:         "mac:!@#$",
			scheme:        SchemeMAC,
			expectedMatch: true,
			expectedAlter: false,
		},
		{
			description:   "long value after scheme",
			input:         "dns:very.long.domain.name.example.com/path/to/resource",
			scheme:        SchemeDNS,
			expectedMatch: true,
			expectedAlter: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			match, altered := isScheme(tc.input, tc.scheme)

			assert.Equal(tc.expectedMatch, match, "match result mismatch")
			assert.Equal(tc.expectedAlter, altered, "altered flag mismatch")
		})
	}
}

func TestSchemeHelper(t *testing.T) {
	tests := []struct {
		description    string
		input          string
		scheme         string
		expectedMatch  bool
		expectedScheme string
		expectedRest   string
		expectedAlter  bool
	}{
		{
			description:    "valid mac scheme",
			input:          "mac:112233445566",
			scheme:         SchemeMAC,
			expectedMatch:  true,
			expectedScheme: SchemeMAC,
			expectedRest:   "112233445566",
			expectedAlter:  false,
		},
		{
			description:    "valid MAC uppercase",
			input:          "MAC:112233445566",
			scheme:         SchemeMAC,
			expectedMatch:  true,
			expectedScheme: SchemeMAC,
			expectedRest:   "112233445566",
			expectedAlter:  true,
		},
		{
			description:    "no match",
			input:          "dns:example.com",
			scheme:         SchemeMAC,
			expectedMatch:  false,
			expectedScheme: "",
			expectedRest:   "dns:example.com",
			expectedAlter:  false,
		},
		{
			description:    "empty input",
			input:          "",
			scheme:         SchemeMAC,
			expectedMatch:  false,
			expectedScheme: "",
			expectedRest:   "",
			expectedAlter:  false,
		},
		{
			description:    "scheme with path",
			input:          "serial:ABC123/service/path",
			scheme:         SchemeSerial,
			expectedMatch:  true,
			expectedScheme: SchemeSerial,
			expectedRest:   "ABC123/service/path",
			expectedAlter:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			match, scheme, rest, altered := schemeHelper(tc.input, tc.scheme)

			assert.Equal(tc.expectedMatch, match, "match result mismatch")
			assert.Equal(tc.expectedScheme, scheme, "scheme mismatch")
			assert.Equal(tc.expectedRest, rest, "rest mismatch")
			assert.Equal(tc.expectedAlter, altered, "altered flag mismatch")
		})
	}
}
