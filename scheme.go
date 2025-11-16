// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

const (
	SchemeMAC     = "mac"
	SchemeUUID    = "uuid"
	SchemeDNS     = "dns"
	SchemeSerial  = "serial"
	SchemeSelf    = "self"
	SchemeEvent   = "event"
	SchemeUnknown = ""
)

// getScheme extracts the scheme from the locator string and returns the const
// representation along with the remaining string after the scheme.
func getScheme(s string) (string, string, bool) {
	if len(s) == 0 {
		return "", s, false
	}

	switch s[0] {
	case 'd', 'D':
		if match, scheme, rest, altered := schemeHelper(s, SchemeDNS); match {
			return scheme, rest, altered
		}
	case 'e', 'E':
		if match, scheme, rest, altered := schemeHelper(s, SchemeEvent); match {
			return scheme, rest, altered
		}
	case 'm', 'M':
		if match, scheme, rest, altered := schemeHelper(s, SchemeMAC); match {
			return scheme, rest, altered
		}
	case 's', 'S':
		if match, scheme, rest, altered := schemeHelper(s, SchemeSelf); match {
			return scheme, rest, altered
		}

		if match, scheme, rest, altered := schemeHelper(s, SchemeSerial); match {
			return scheme, rest, altered
		}
	case 'u', 'U':
		if match, scheme, rest, altered := schemeHelper(s, SchemeUUID); match {
			return scheme, rest, altered
		}
	}

	return "", s, false
}

func schemeHelper(s string, scheme string) (bool, string, string, bool) {
	hasPrefix, altered := isScheme(s, scheme)
	if hasPrefix {
		return true, scheme, s[len(scheme)+1:], altered
	}
	return false, "", s, altered
}

// isScheme checks if the string s starts with the given prefix and the next
// character is a colon, ignoring case.
// It returns true if it matches, along with a boolean indicating if any case
// alterations were made.
func isScheme(s, scheme string) (bool, bool) {
	if len(s) < len(scheme)+1 {
		return false, false
	}

	var altered bool
	for i := 0; i < len(scheme); i++ {
		c1 := s[i]
		c2 := scheme[i]
		if c1 >= 'A' && c1 <= 'Z' {
			altered = true
			c1 += 'a' - 'A'
		}
		if c1 != c2 {
			return false, altered
		}
	}

	// Next character must be a colon
	if s[len(scheme)] != ':' {
		return false, altered
	}

	return true, altered
}
