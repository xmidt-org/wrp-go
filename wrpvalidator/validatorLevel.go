// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"fmt"
	"sort"
	"strings"
)

type validatorLevel int

const (
	UnknownLevel validatorLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	lastLevel
)

var (
	validatorLevelUnmarshal = map[string]validatorLevel{
		"unknown": UnknownLevel,
		"info":    InfoLevel,
		"warning": WarningLevel,
		"error":   ErrorLevel,
	}
	validatorLevelMarshal = map[validatorLevel]string{
		UnknownLevel: "unknown",
		InfoLevel:    "info",
		WarningLevel: "warning",
		ErrorLevel:   "error",
	}
)

// Empty returns true if the value is UnknownLevel (the default),
// otherwise false is returned.
func (vt validatorLevel) IsEmpty() bool {
	return UnknownLevel == vt
}

func (vt validatorLevel) IsValid() bool {
	return UnknownLevel < vt && vt < lastLevel
}

// String returns a human-readable string representation for an existing validatorLevel,
// otherwise String returns the unknownEnum string value.
func (vt validatorLevel) String() string {
	if value, ok := validatorLevelMarshal[vt]; ok {
		return value
	}

	return validatorLevelMarshal[UnknownLevel]
}

// UnmarshalText returns the validatorLevel's enum value
func (vt *validatorLevel) UnmarshalText(b []byte) error {
	s := string(b)
	r, ok := validatorLevelUnmarshal[s]
	if !ok {
		return fmt.Errorf("ValidatorLevel error: '%s' does not match any valid options: %s",
			s, vt.getKeys())
	}

	*vt = r
	return nil
}

// getKeys returns the string keys for the validatorLevel enums.
func (vt validatorLevel) getKeys() string {
	keys := make([]string, 0, len(validatorLevelUnmarshal))
	for k := range validatorLevelUnmarshal {
		k = "'" + k + "'"
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
