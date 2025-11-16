// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"fmt"
)

// normalizeMAC normalizes a MAC address by removing delimiters and lowercasing hex digits.
// Returns an error if the MAC address is invalid.
// This function allocates a new string for the normalized MAC.
func normalizeMAC(mac string) (string, error) {
	if len(mac) == 12 && isLowerHexString(mac) {
		return mac, nil // Already normalized, no allocation needed
	}

	rv := make([]byte, 0, len(mac))

	for i := 0; i < len(mac); i++ {
		switch mac[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f':
			rv = append(rv, mac[i])
		case 'A', 'B', 'C', 'D', 'E', 'F':
			rv = append(rv, mac[i]+('a'-'A'))
		case '-', ':', '.', ',':
			// Skip delimiters
		default:
			return "", fmt.Errorf("%w: invalid character `%c` in MAC address", ErrorInvalidDeviceName, mac[i])
		}
	}

	if len(rv) != 12 {
		return "", fmt.Errorf("%w: invalid MAC address length", ErrorInvalidDeviceName)
	}

	return string(rv), nil
}

func isLowerHexString(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'a', 'b', 'c', 'd', 'e', 'f':
		default:
			return false
		}
	}
	return true
}
