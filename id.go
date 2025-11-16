// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"errors"
	"fmt"
	"strings"
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
)

// DeviceID represents a normalized identifier for a device.
type DeviceID string

// Bytes is a convenience function to obtain the []byte representation of a DeviceID.
func (id DeviceID) Bytes() []byte {
	return []byte(id)
}

// Prefix returns the scheme portion of the DeviceID (the part before the colon).
// For example, "mac:112233445566" returns "mac".
func (id DeviceID) Prefix() string {
	prefix, _, _ := strings.Cut(string(id), ":")
	return prefix
}

// ID returns the identifier portion of the DeviceID (the part after the colon).
// For example, "mac:112233445566" returns "112233445566".
func (id DeviceID) ID() string {
	_, idPart, _ := strings.Cut(string(id), ":")
	return idPart
}

// AsLocator converts a device identifier into a locator.
func (id DeviceID) AsLocator() Locator {
	prefix, idPart, _ := strings.Cut(string(id), ":")

	l := Locator{
		Scheme:    prefix,
		Authority: idPart,
		ID:        id,
	}

	switch prefix {
	case SchemeDNS, SchemeEvent:
		l.ID = invalidDeviceID
	}

	return l
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

// ParseDeviceID parses a raw device name into a normalized DeviceID.
// It handles case-insensitive scheme names, MAC address normalization,
// and removes any service/path suffix from the input.
//
// Supported schemes: mac, uuid, dns, serial, self, event
//
// Example inputs:
//   - "MAC:11:22:33:44:55:66" -> DeviceID("mac:112233445566")
//   - "uuid:12345" -> DeviceID("uuid:12345")
//   - "mac:112233445566/service" -> DeviceID("mac:112233445566")
func ParseDeviceID(deviceName string) (DeviceID, error) {
	id, _, err := parseDeviceID(deviceName)
	return id, err
}

// parseDeviceID parses a raw device name into a canonicalized identifier,
// returning any remaining unparsed portion as well.
func parseDeviceID(s string) (DeviceID, string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return DeviceID(""), "", fmt.Errorf("%w: empty device ID", ErrorInvalidDeviceName)
	}

	scheme, rest, altered := getScheme(s)
	if scheme == "" {
		return DeviceID(""), "", fmt.Errorf("%w: unknown or missing scheme", ErrorInvalidDeviceName)
	}

	// Self is a special case with no authority
	if scheme == SchemeSelf {
		return DeviceID("self:"), rest, nil
	}

	authority, _, _ := strings.Cut(rest, "/")
	rest, _ = strings.CutPrefix(rest, authority)

	trimmedAuthority := strings.TrimSpace(authority)
	if trimmedAuthority != authority {
		altered = true
		authority = trimmedAuthority
	}
	if authority == "" {
		return DeviceID(""), "", fmt.Errorf("%w: missing authority in device ID", ErrorInvalidDeviceName)
	}

	if scheme == SchemeMAC {
		normalized, err := normalizeMAC(authority)
		if err != nil {
			return DeviceID(""), "", err
		}

		if authority != normalized {
			altered = true
			authority = normalized
		}
	}

	if !altered {
		// No alterations were made, return original string as DeviceID
		return DeviceID(s[0 : len(scheme)+1+len(authority)]), rest, nil
	}

	return DeviceID(scheme + ":" + authority), rest, nil
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
	if strings.Contains(l.Service, "/") {
		return fmt.Errorf("%w: service `%s` contains a `/` character", ErrorInvalidLocator, l.Service)
	}

	switch l.Scheme {
	case SchemeDNS:
		return l.validateSchemeDNS()
	case SchemeEvent:
		return l.validateSchemeEvent()
	case SchemeMAC:
		return l.validateSchemeMAC()
	case SchemeSelf:
		return l.validateSchemeSelf()
	case SchemeUUID, SchemeSerial:
		return l.validateSchemeUUID()
	}

	return fmt.Errorf("%w: unknown scheme `%s`", ErrorInvalidLocator, l.Scheme)
}

func (l Locator) validateSchemeDNS() error {
	if l.Authority == "" {
		return fmt.Errorf("%w: empty authority", ErrorInvalidLocator)
	}
	if l.Service != "" {
		return fmt.Errorf("%w: service `%s` is not allowed for event locators", ErrorInvalidLocator, l.Service)
	}
	if l.ID != DeviceID("") {
		return fmt.Errorf("%w: ID `%s` is not allowed for dns locators", ErrorInvalidLocator, l.ID)
	}
	return nil
}

func (l Locator) validateSchemeSelf() error {
	if l.Authority != "" {
		return fmt.Errorf("%w: authority `%s` is not allowed for self locators", ErrorInvalidLocator, l.Authority)
	}
	if l.ID != DeviceID("self:") {
		return fmt.Errorf("%w: ID `%s` does not match scheme `self`", ErrorInvalidLocator, l.ID)
	}

	return nil
}

func (l Locator) validateSchemeMAC() error {
	if l.Authority == "" {
		return fmt.Errorf("%w: empty authority", ErrorInvalidLocator)
	}
	if l.ID.ID() != l.Authority {
		return fmt.Errorf("%w: ID `%s` does not match scheme `%s` and authority `%s`", ErrorInvalidLocator, l.ID, l.Scheme, l.Authority)
	}
	if l.ID.Prefix() != SchemeMAC {
		return fmt.Errorf("%w: ID `%s` does not match scheme `%s`", ErrorInvalidLocator, l.ID, l.Scheme)
	}
	if _, err := normalizeMAC(l.Authority); err != nil {
		return fmt.Errorf("%w: authority `%s` is not a valid MAC address: %v", ErrorInvalidLocator, l.Authority, err)
	}
	return nil
}

func (l Locator) validateSchemeUUID() error {
	if l.Authority == "" {
		return fmt.Errorf("%w: empty authority", ErrorInvalidLocator)
	}
	if l.ID.ID() != l.Authority {
		return fmt.Errorf("%w: ID `%s` does not match scheme `%s` and authority `%s`", ErrorInvalidLocator, l.ID, l.Scheme, l.Authority)
	}
	if l.ID.Prefix() != l.Scheme {
		return fmt.Errorf("%w: ID `%s` does not match scheme `%s`", ErrorInvalidLocator, l.ID, l.Scheme)
	}
	return nil
}

func (l Locator) validateSchemeEvent() error {
	if l.Authority == "" {
		return fmt.Errorf("%w: empty authority", ErrorInvalidLocator)
	}
	if l.Service != "" {
		return fmt.Errorf("%w: service `%s` is not allowed for event locators", ErrorInvalidLocator, l.Service)
	}
	if l.ID != DeviceID("") {
		return fmt.Errorf("%w: ID `%s` is not allowed for event locators", ErrorInvalidLocator, l.ID)
	}
	return nil
}

// ParseLocator parses a locator string into a Locator struct.
// It extracts the scheme, authority, service, and ignored portions.
//
// Format: {scheme}:{authority}[/service][/ignored...]
//
// The function attempts to minimize allocations, though some allocations
// may occur for DeviceID normalization (e.g., MAC address formatting).
//
// Example inputs:
//   - "mac:112233445566" -> Locator with MAC scheme and authority
//   - "dns:example.com/service/path" -> DNS locator with service and ignored path
//   - "self:/service" -> Self locator with service
func ParseLocator(locator string) (Locator, error) {
	id, rest, err := parseDeviceID(locator)
	if err != nil {
		return Locator{}, errors.Join(ErrorInvalidLocator, err)
	}

	l := id.AsLocator()

	// For event/dns schemes, everything after authority goes to Ignored
	// For device schemes, next component is Service, rest is Ignored
	switch l.Scheme {
	case SchemeDNS, SchemeEvent:
		l.Ignored = rest
	default:
		if rest != "" && !strings.HasPrefix(rest, "/") {
			return Locator{}, ErrorInvalidLocator
		}
		rest = strings.TrimPrefix(rest, "/")
		l.Service, _, _ = strings.Cut(rest, "/")

		rest = strings.TrimPrefix(rest, l.Service)
		l.Ignored = rest
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
	needed := len(l.Scheme) + 1 + len(l.Authority)
	if l.Service != "" {
		needed += 1 + len(l.Service)
	}
	needed += len(l.Ignored)

	var buf strings.Builder

	buf.Grow(needed)

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
