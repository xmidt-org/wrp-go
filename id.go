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

// String is a convenience function to obtain the string representation of the
// prefix portion of the ID.
func (id DeviceID) Prefix() string {
	prefix, _ := id.split()
	return prefix
}

// ID is a convenience function to obtain the string representation of the
// identifier portion of the ID.
func (id DeviceID) ID() string {
	_, idPart := id.split()
	return idPart
}

func (id DeviceID) split() (prefix, idPart string) {
	parts := strings.SplitN(string(id), ":", 2)
	if len(parts) != 2 {
		return parts[0], ""
	}

	return parts[0], parts[1]
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
