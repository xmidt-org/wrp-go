// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone"
	"github.com/xmidt-org/wrp-go/v3"
	"go.uber.org/multierr"
)

const (
	serialPrefix = "serial"
	uuidPrefix   = "uuid"
	eventPrefix  = "event"
	dnsPrefix    = "dns"
	macPrefix    = "mac"
)

const (
	hexDigits     = "0123456789abcdefABCDEF"
	macDelimiters = ":-.,"
	macLength     = 12
	scheme        = macPrefix + `|` + uuidPrefix + `|` + eventPrefix + `|` + dnsPrefix + `|` + serialPrefix
)

var (
	ErrorInvalidMessageEncoding = NewValidatorError(errors.New("invalid message encoding"), "", nil)
	ErrorInvalidMessageType     = NewValidatorError(errors.New("invalid message type"), "", []string{"Type"})
	ErrorInvalidSource          = NewValidatorError(errors.New("invalid Source name"), "", []string{"Source"})
	ErrorInvalidDestination     = NewValidatorError(errors.New("invalid Destination name"), "", []string{"Destination"})
	errorInvalidUUID            = errors.New("invalid UUID")
	errorEmptyAuthority         = errors.New("invalid empty authority (ID)")
	errorInvalidMacLength       = errors.New("invalid mac length")
	errorInvalidCharacter       = errors.New("invalid character")
	errorInvalidLocatorPattern  = errors.New("value given doesn't match expected locator pattern")
)

// locatorPattern is the precompiled regular expression that all source and dest locators must match.
// Matching is partial, as everything after the authority (ID) is ignored. https://xmidt.io/docs/wrp/basics/#locators
var locatorPattern = regexp.MustCompile(
	`^(?P<scheme>(?i)` + scheme + `):(?P<authority>[^/]+)?`,
)

// SpecWithMetrics ensures messages are valid based on each spec validator in the list.
// Only validates the opinionated portions of the spec.
// SpecWithMetrics validates the following fields: UTF8 (all string fields), MessageType, Source, Destination
func SpecWithMetrics(tf *touchstone.Factory, labelNames ...string) (Validators, error) {
	var errs error
	utf8v, err := NewUTF8WithMetric(tf, labelNames...)
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	mtv, err := NewMessageTypeWithMetric(tf, labelNames...)
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	sv, err := NewSourceWithMetric(tf, labelNames...)
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	dv, err := NewDestinationWithMetric(tf, labelNames...)
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	return Validators{}.AddFunc(utf8v, mtv, sv, dv), errs
}

// NewUTF8WithMetric returns a UTF8 validator with a metric middleware.
func NewUTF8WithMetric(tf *touchstone.Factory, labelNames ...string) (ValidatorFunc, error) {
	m, err := newUTF8ErrorTotal(tf, labelNames...)

	return func(msg wrp.Message, ls prometheus.Labels) error {
		err := UTF8(msg)
		if err != nil {
			m.With(ls).Add(1.0)
		}

		return err
	}, err
}

// NewMessageTypeWithMetric returns a MessageType validator with a metric middleware.
func NewMessageTypeWithMetric(tf *touchstone.Factory, labelNames ...string) (ValidatorFunc, error) {
	m, err := newMessageTypeErrorTotal(tf, labelNames...)

	return func(msg wrp.Message, ls prometheus.Labels) error {
		err := MessageType(msg)
		if err != nil {
			m.With(ls).Add(1.0)
		}

		return err
	}, err
}

// NewSourceWithMetric returns a Source validator with a metric middleware.
func NewSourceWithMetric(tf *touchstone.Factory, labelNames ...string) (ValidatorFunc, error) {
	m, err := newSourceErrorTotal(tf, labelNames...)

	return func(msg wrp.Message, ls prometheus.Labels) error {
		err := Source(msg)
		if err != nil {
			m.With(ls).Add(1.0)
		}

		return err
	}, err
}

// NewDestinationWithMetric returns a Destination validator with a metric middleware.
func NewDestinationWithMetric(tf *touchstone.Factory, labelNames ...string) (ValidatorFunc, error) {
	m, err := newDestinationErrorTotal(tf, labelNames...)

	return func(msg wrp.Message, ls prometheus.Labels) error {
		err := Destination(msg)
		if err != nil {
			m.With(ls).Add(1.0)
		}

		return err
	}, err
}

// UTF8 takes messages and validates that it contains UTF-8 strings.
func UTF8(m wrp.Message) error {
	if err := wrp.UTF8(m); err != nil {
		return fmt.Errorf("%w: %v", ErrorInvalidMessageEncoding, err)
	}

	return nil
}

// MessageType takes messages and validates their Type.
func MessageType(m wrp.Message) error {
	if m.Type < wrp.Invalid0MessageType || m.Type > wrp.LastMessageType {
		return ErrorInvalidMessageType
	}

	switch m.Type {
	case wrp.Invalid0MessageType, wrp.Invalid1MessageType, wrp.LastMessageType:
		return ErrorInvalidMessageType
	}

	return nil
}

// Source takes messages and validates their Source.
// Only mac and uuid sources are validated. Serial, event and dns sources are
// not validated.
func Source(m wrp.Message) error {
	if err := validateLocator(m.Source); err != nil {
		return fmt.Errorf("%w '%s': %v", ErrorInvalidSource, m.Source, err)
	}

	return nil
}

// Destination takes messages and validates their Destination.
// Only mac and uuid destinations are validated. Serial, event and dns destinations are
// not validated.
func Destination(m wrp.Message) error {
	if err := validateLocator(m.Destination); err != nil {
		return fmt.Errorf("%w '%s': %v", ErrorInvalidDestination, m.Destination, err)
	}

	return nil
}

// validateLocator validates a given locator's scheme and authority (ID).
// Only mac and uuid schemes' IDs are validated. IDs from serial, event and dns schemes are
// not validated.
func validateLocator(l string) error {
	match := locatorPattern.FindStringSubmatch(l)
	if match == nil {
		return fmt.Errorf("%w: %s", errorInvalidLocatorPattern, scheme)
	}

	idPart := match[2]
	if len(idPart) == 0 {
		return errorEmptyAuthority
	}

	switch strings.ToLower(match[1]) {
	case macPrefix:
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

		if invalidCharacter != -1 {
			return fmt.Errorf("%w: %v", errorInvalidCharacter, strconv.QuoteRune(invalidCharacter))
		} else if len(idPart) != macLength {
			return errorInvalidMacLength
		}
	case uuidPrefix:
		if _, err := uuid.Parse(idPart); err != nil {
			return fmt.Errorf("%w: %v", errorInvalidUUID, err)
		}
	}

	return nil
}
