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

	SchemeMAC     = "mac"
	SchemeUUID    = "uuid"
	SchemeDNS     = "dns"
	SchemeSerial  = "serial"
	SchemeSelf    = "self"
	SchemeEvent   = "event"
	SchemeUnknown = ""
)

var (
	ErrorInvalidDeviceName = errors.New("invalid device name")
	ErrorInvalidLocator    = errors.New("invalid locator")

	invalidDeviceID = DeviceID("")

	// Locator/DeviceID form:
	//   {scheme|prefix}:{authority|id}/{service}/{ignored}
	//
	//  If the scheme is "mac", "uuid", or "serial" then the authority is the
	//	device identifier.
	//  If the scheme is "dns" then the authority is the FQDN of the service.
	//  If the scheme is "event" then the authority is the event name.
	//  If the scheme is "self" then the authority is the empty string.

	// devicIDPattern is the precompiled regular expression that all device identifiers must match.
	// Matching is partial, as everything after the service is ignored.
	deviceIDPattern = regexp.MustCompile(
		`^(?P<prefix>(?i)mac|uuid|dns|serial|self):(?P<id>[^/]+)(?P<service>/[^/]+)?`,
	)

	// locatorPattern is the precompiled regular expression that all locators must match.
	locatorPattern = regexp.MustCompile(
		`^(?P<scheme>(?i)mac|uuid|dns|serial|event|self):(?P<authority>[^/]+)?(?P<service>/[^/]+)?(?P<ignored>.+)?`,
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

// AsLocator converts a device identifier into a locator.
func (id DeviceID) AsLocator() Locator {
	prefix, idPart := id.split()
	return Locator{
		Scheme:    prefix,
		Authority: idPart,
		ID:        id,
	}
}

// Is returns true if the device identifier is one of the provided device identifiers.
// This is a convenience function to avoid having to compare the ID field directly
// & makes code more readable.
func (id DeviceID) Is(oneOf ...DeviceID) bool {
	for _, v := range oneOf {
		if id == v {
			return true
		}
	}

	return false
}

// ParseID parses a raw device name into a canonicalized identifier.
func ParseDeviceID(deviceName string) (DeviceID, error) {
	match := deviceIDPattern.FindStringSubmatch(deviceName)
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
//
// See https://xmidt.io/docs/wrp/basics/#locators for more details.
type Locator struct {
	// Scheme is the scheme type of the locator.  A CPE will have the forms
	// `mac`, `uuid`, `serial`, `self`.  A server or cloud service will have
	// the form `dns`.  An event locator that is used for pub-sub listeners
	// will have the form `event`.
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

	// ID is the device identifier portion of the locator if it is one.
	ID DeviceID
}

// Validate ensures that the locator is well-formed and adheres to the WRP
// specification.
func (l Locator) Validate() error {
	switch l.Scheme {
	case SchemeSelf:
		if l.Authority != "" {
			return fmt.Errorf("%w: authority `%s` is not allowed for self locators", ErrorInvalidLocator, l.Authority)
		}
		if l.ID != DeviceID("self:") {
			return fmt.Errorf("%w: ID `%s` does not match scheme `self`", ErrorInvalidLocator, l.ID)
		}
	case SchemeMAC, SchemeUUID, SchemeSerial:
		if l.Authority == "" {
			return fmt.Errorf("%w: empty authority", ErrorInvalidLocator)
		}
		id, err := makeDeviceID(l.Scheme, l.Authority)
		if err != nil {
			return err
		}

		if l.ID != id {
			return fmt.Errorf("%w: ID `%s` does not match scheme `%s` and authority `%s`", ErrorInvalidLocator, l.ID, l.Scheme, l.Authority)
		}
	case SchemeEvent:
		if l.Authority == "" {
			return fmt.Errorf("%w: empty authority", ErrorInvalidLocator)
		}
		if l.Service != "" {
			return fmt.Errorf("%w: service `%s` is not allowed for event locators", ErrorInvalidLocator, l.Service)
		}
		if l.ID != DeviceID("") {
			return fmt.Errorf("%w: ID `%s` is not allowed for event locators", ErrorInvalidLocator, l.ID)
		}

	case SchemeDNS:
		if l.Authority == "" {
			return fmt.Errorf("%w: empty authority", ErrorInvalidLocator)
		}
		if l.Service != "" {
			return fmt.Errorf("%w: service `%s` is not allowed for dns locators", ErrorInvalidLocator, l.Service)
		}
		if l.ID != DeviceID("") {
			return fmt.Errorf("%w: ID `%s` is not allowed for dns locators", ErrorInvalidLocator, l.ID)
		}
	default:
		return fmt.Errorf("%w: unknown scheme `%s`", ErrorInvalidLocator, l.Scheme)
	}

	if strings.Contains(l.Service, "/") {
		return fmt.Errorf("%w: service `%s` contains a `/` character", ErrorInvalidLocator, l.Service)
	}

	return nil
}

// ParseLocator parses a raw locator string into a canonicalized locator.
func ParseLocator(locator string) (Locator, error) {
	match := locatorPattern.FindStringSubmatch(locator)
	if match == nil {
		return Locator{}, fmt.Errorf("%w: `%s` does not match expected locator pattern", ErrorInvalidLocator, locator)
	}

	var l Locator

	l.Scheme = strings.TrimSpace(strings.ToLower(match[1]))
	l.Authority = strings.TrimSpace(match[2])
	if len(match) > 3 {
		l.Service = strings.TrimSpace(strings.TrimPrefix(match[3], "/"))
	}
	if len(match) > 4 {
		l.Ignored = strings.TrimSpace(match[4])
	}

	// If the locator is a device identifier, then we need to parse it.
	switch l.Scheme {
	case SchemeEvent, SchemeDNS:
		if l.Service != "" {
			l.Ignored = "/" + l.Service + l.Ignored
			l.Service = ""
		}
	case SchemeMAC, SchemeUUID, SchemeSerial, SchemeSelf:
		id, err := makeDeviceID(l.Scheme, l.Authority)
		if err != nil {
			return Locator{}, fmt.Errorf("%w: unable to make a device ID with scheme `%s` and authority `%s`", err, l.Scheme, l.Authority)
		}
		l.ID = id
	}

	if err := l.Validate(); err != nil {
		return Locator{}, err
	}

	return l, nil
}

// HasDeviceID returns true if the locator is a device identifier.
func (l Locator) HasDeviceID() bool {
	return l.ID != ""
}

// IsSelf returns true if the locator is a self locator.
func (l Locator) IsSelf() bool {
	return l.Scheme == SchemeSelf
}

func (l Locator) String() string {
	var buf strings.Builder

	buf.WriteString(l.Scheme)
	buf.WriteString(":")
	buf.WriteString(l.Authority)
	if l.Service != "" {
		buf.WriteString("/")
		buf.WriteString(l.Service)
	}

	if l.Ignored != "" {
		buf.WriteString(l.Ignored)
	}

	return buf.String()
}

// Is returns true if the locator is one of the provided device identifiers.
// This is a convenience function to avoid having to compare the ID field
// directly & makes code more readable.
func (l Locator) Is(oneOf ...DeviceID) bool {
	return l.ID.Is(oneOf...)
}

func makeDeviceID(prefix, idPart string) (DeviceID, error) {
	prefix = strings.ToLower(prefix)
	switch prefix {
	case SchemeSelf:
		if idPart != "" {
			return invalidDeviceID, ErrorInvalidDeviceName
		}
	case SchemeUUID, SchemeSerial, SchemeDNS, SchemeEvent:
		if idPart == "" {
			return invalidDeviceID, ErrorInvalidDeviceName
		}
	case SchemeMAC:
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

		if invalidCharacter != -1 ||
			((len(idPart) != 12) && (len(idPart) != 16) && (len(idPart) != 40)) {
			return invalidDeviceID, ErrorInvalidDeviceName
		}
	}

	return DeviceID(fmt.Sprintf("%s:%s", prefix, idPart)), nil
}
