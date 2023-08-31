// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const (
	hexDigits     = "0123456789abcdefABCDEF"
	macDelimiters = ":-.,"
	macPrefix     = "mac"
	macLength     = 12
)

var (
	ErrorInvalidDeviceName = errors.New("invalid device name")

	invalidDeviceID = DeviceID("")

	// idPattern is the precompiled regular expression that all device identifiers must match.
	// Matching is partial, as everything after the service is ignored.
	DeviceIDPattern = regexp.MustCompile(
		`^(?P<prefix>(?i)mac|uuid|dns|serial):(?P<id>[^/]+)(?P<service>/[^/]+)?`,
	)
)

// ID represents a normalized identifier for a device.
type DeviceID string

// Bytes is a convenience function to obtain the []byte representation of an ID.
func (id DeviceID) Bytes() []byte {
	return []byte(id)
}

// ParseID parses a raw device name into a canonicalized identifier.
func ParseDeviceID(deviceName string) (DeviceID, error) {
	match := DeviceIDPattern.FindStringSubmatch(deviceName)
	if match == nil {
		return invalidDeviceID, ErrorInvalidDeviceName
	}

	var (
		prefix = strings.ToLower(match[1])
		idPart = match[2]
	)

	if prefix == macPrefix {
		var invalidCharacter rune = -1
		idPart = strings.Map(
			func(r rune) rune {
				switch {
				case strings.ContainsRune(hexDigits, r):
					return unicode.ToLower(r)
				case strings.ContainsRune(macDelimiters, r):
					return -1
				default:
					invalidCharacter = r
					return -1
				}
			},
			idPart,
		)

		if invalidCharacter != -1 || len(idPart) != macLength {
			return invalidDeviceID, ErrorInvalidDeviceName
		}
	}

	return DeviceID(fmt.Sprintf("%s:%s", prefix, idPart)), nil
}
