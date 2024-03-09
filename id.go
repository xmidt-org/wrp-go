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
	ErrorInvalidLocator    = errors.New("invalid locator")

	invalidDeviceID = DeviceID("")

	// DevicIDPattern is the precompiled regular expression that all device identifiers must match.
	// Matching is partial, as everything after the service is ignored.
	DeviceIDPattern = regexp.MustCompile(
		`^(?P<prefix>(?i)mac|uuid|dns|serial):(?P<id>[^/]+)(?P<service>/[^/]+)?`,
	)

	// LocatorPattern is the precompiled regular expression that all locators must match.
	LocatorPattern = regexp.MustCompile(
		`^(?P<scheme>(?i)mac|uuid|dns|serial|event):(?P<authority>[^/]+)(?P<service>/[^/]+)?(?P<ignored>/[^/]+)?`,
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

	return makeDeviceID(match[1], match[2])

}

// Locator represents a device locator, which is a device identifier an optional
// service name and an optional ignored portion at the end.
//
// The general format is:
//
//	{scheme}:{authority}/{service}/{ignored}
type Locator struct {
	// Scheme is the scheme type of the locator.  A CPE will have the forms
	// `mac`, `uuid`, `serial`.  A server or cloud service will have the form
	// `dns`.  An event locator that is used for pub-sub listeners will have
	// the form `event`.
	//
	// The Scheme MUST NOT be used to determine where to send a message, but
	// rather to determine how to interpret the authority and service.
	//
	// The Scheme value will always be lower case.
	Scheme string

	// Authority is the authority portion of the locator.  For a CPE, this
	// will be the device identifier.  For a server or cloud service, this
	// will be the DNS name of the service.  For an event locator, this will
	// be the event name.
	Authority string

	// Service is the service name portion of the locator.  This is optional
	// and is used to identify which service(s) the message is targeting or
	// originated from.  A Service value will not contain any `/` characters.
	Service string

	// Ignored is an optional portion of the locator that is ignored by the
	// WRP spec, but is provided to consumers for their usage.  The Ignored
	// value will contain a prefix of the `/` character.
	Ignored string

	// id is the device identifier portion of the locator if it is one.
	id *DeviceID
}

// ParseLocator parses a raw locator string into a canonicalized locator.
func ParseLocator(locator string) (*Locator, error) {
	match := LocatorPattern.FindStringSubmatch(locator)
	if match == nil {
		return nil, ErrorInvalidLocator
	}

	var l Locator

	l.Scheme = strings.ToLower(match[1])
	l.Authority = match[2]
	if len(match) > 3 {
		l.Service = strings.TrimPrefix(match[3], "/")
	}
	if len(match) > 4 {
		l.Ignored = match[4]
	}

	switch l.Scheme {
	case "mac", "uuid", "serial": // device_id locators
		id, err := makeDeviceID(l.Scheme, l.Authority)
		if err != nil {
			return nil, err
		}
		l.id = &id
	default:
	}

	return &l, nil
}

// DeviceID returns the device identifier portion of the locator.
func (l *Locator) DeviceID() *DeviceID {
	return l.id
}

// IsDeviceID returns true if the locator is a device identifier.
func (l *Locator) IsDeviceID() bool {
	return l.id != nil
}

func (l *Locator) String() string {
	var buf strings.Builder

	buf.WriteString(l.Scheme)
	buf.WriteString(":")
	buf.WriteString(l.Authority)
	if l.Service != "" {
		buf.WriteString("/")
		buf.WriteString(l.Service)

		if l.Ignored != "" {
			buf.WriteString(l.Ignored)
		}
	}

	return buf.String()
}

func makeDeviceID(prefix, idPart string) (DeviceID, error) {
	prefix = strings.ToLower(prefix)
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
