// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type validatorType int

const (
	UnknownType validatorType = iota
	AlwaysInvalidType
	AlwaysValidType
	UTF8Type
	MessageTypeType
	SourceType
	DestinationType
	SimpleResponseRequestTypeType
	SimpleEventTypeType
	SpansType
	lastType
)

var errValidatorTypeInvalid = errors.New("validator type is invalid")

var (
	validatorTypeUnmarshal = map[string]validatorType{
		"unknown":        UnknownType,
		"always_invalid": AlwaysInvalidType,
		"always_valid":   AlwaysValidType,
		"utf8":           UTF8Type,
		"msg_type":       MessageTypeType,
		"source":         SourceType,
		"destination":    DestinationType,
		"simple_res_req": SimpleResponseRequestTypeType,
		"simple_event":   SimpleEventTypeType,
		"spans":          SpansType,
	}
	validatorTypeMarshal = map[validatorType]string{
		UnknownType:                   "unknown",
		AlwaysInvalidType:             "always_invalid",
		AlwaysValidType:               "always_valid",
		UTF8Type:                      "utf8",
		MessageTypeType:               "msg_type",
		SourceType:                    "source",
		DestinationType:               "destination",
		SimpleResponseRequestTypeType: "simple_res_req",
		SimpleEventTypeType:           "simple_event",
		SpansType:                     "spans",
	}
)

// Empty returns true if the value is UnknownType (the default),
// otherwise false is returned.
func (vt validatorType) IsEmpty() bool {
	return vt == UnknownType
}

func (vt validatorType) IsValid() bool {
	return UnknownType < vt && vt < lastType
}

// String returns a human-readable string representation for an existing validatorType,
// otherwise String returns the unknownEnum string value.
func (vt validatorType) String() string {
	if value, ok := validatorTypeMarshal[vt]; ok {
		return value
	}

	return validatorTypeMarshal[UnknownType]
}

// UnmarshalText returns the validatorType's enum value
func (vt *validatorType) UnmarshalText(b []byte) error {
	s := strings.ToLower(string(b))
	r, ok := validatorTypeUnmarshal[s]
	if !ok {
		return fmt.Errorf("ValidatorType error: '%s' does not match any valid options: %s",
			s, vt.getKeys())
	}

	*vt = r
	return nil
}

// getKeys returns the string keys for the validatorType enums.
func (vt validatorType) getKeys() string {
	keys := make([]string, 0, len(validatorTypeUnmarshal))
	for k := range validatorTypeUnmarshal {
		k = "'" + k + "'"
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
