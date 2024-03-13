// SPDX-FileCopyrightText: 2024 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrInvalidPartnerID   = errors.New("invalid partner ID")
	ErrInvalidSource      = errors.New("invalid source locator")
	ErrInvalidDest        = errors.New("invalid destination locator")
	ErrInvalidString      = errors.New("invalid UTF-8 string")
)

// Normify applies a series of normalizing options to a WRP message.
type Normify struct {
	opts []NormifyOption
}

// NormifyOption is a functional option for normalizing a WRP message.
type NormifyOption interface {
	// normify applies the option to the given message.
	normify(*Message) error
}

// optionFunc is an adapter to allow the use of ordinary functions as
// normalizing options.
type optionFunc func(*Message) error

var _ NormifyOption = optionFunc(nil)

func (f optionFunc) normify(m *Message) error {
	return f(m)
}

// New creates a new Correctifier with the given options.
func New(opts ...NormifyOption) *Normify {
	return &Normify{
		opts: opts,
	}
}

// Process applies the normalizing and validating options to the message.  It
// returns an error if any of the options fail.
func (n *Normify) Process(m *Message) error {
	for _, opt := range n.opts {
		if opt != nil {
			if err := opt.normify(m); err != nil {
				return err
			}
		}
	}
	return nil
}

// errorOption returns an option that always returns the given error.
func errorOption(err error) NormifyOption {
	return optionFunc(func(*Message) error {
		return err
	})
}

// Options returns a new option that applies all of the given options in order.
func Options(opts ...NormifyOption) NormifyOption {
	return optionFunc(func(m *Message) error {
		for _, opt := range opts {
			if opt != nil {
				if err := opt.normify(m); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// -- Normalizers --------------------------------------------------------------

// ReplaceAnySelfLocator replaces any `self:` based locator with the scheme and
// authority of the given locator.  If the given locator is not valid, the
// option returns an error.
func ReplaceAnySelfLocator(me string) NormifyOption {
	return Options(
		ReplaceSourceSelfLocator(me),
		ReplaceDestinationSelfLocator(me),
	)
}

// ReplaceSourceSelfLocator replaces a `self:` based source locator with the
// scheme and authority of the given locator.  If the given locator is not valid,
// the option returns an error.
func ReplaceSourceSelfLocator(me string) NormifyOption {
	full, err := ParseLocator(me)
	if err != nil {
		return errorOption(err)
	}

	return optionFunc(func(m *Message) error {
		src, err := ParseLocator(m.Source)
		if err != nil {
			return err
		}

		if src.Scheme == "self" {
			src.Scheme = full.Scheme
			src.Authority = full.Authority
			m.Source = src.String()
		}

		return nil
	})
}

// ReplaceDestinationSelfLocator replaces the destination of the message with the
// given locator if the destination is a `self:` based locator.  If the given
// locator is not valid, the option returns an error.
func ReplaceDestinationSelfLocator(me string) NormifyOption {
	full, err := ParseLocator(me)
	if err != nil {
		return errorOption(err)
	}

	return optionFunc(func(m *Message) error {
		dst, err := ParseLocator(m.Destination)
		if err != nil {
			return err
		}

		if dst.Scheme == "self" {
			dst.Scheme = full.Scheme
			dst.Authority = full.Authority
			m.Destination = dst.String()
		}

		return nil
	})
}

// EnsureTransactionUUID ensures that the message has a transaction UUID.  If
// the message does not have a transaction UUID, a new one is generated and
// added to the message.
func EnsureTransactionUUID() NormifyOption {
	return optionFunc(func(m *Message) error {
		if m.TransactionUUID == "" {
			id, err := uuid.NewRandom()
			if err != nil {
				return err
			}

			m.TransactionUUID = id.String()
		}
		return nil
	})
}

// EnsurePartnerID ensures that the message includes the given partner ID in
// the list.  If not present, the partner ID is added to the list.
func EnsurePartnerID(partnerID string) NormifyOption {
	return optionFunc(func(m *Message) error {
		if m.PartnerIDs == nil {
			m.PartnerIDs = make([]string, 0, 1)
		}
		for _, id := range m.PartnerIDs {
			if id == partnerID {
				return nil
			}
		}
		m.PartnerIDs = append(m.PartnerIDs, partnerID)
		return nil
	})
}

// SetPartnerID ensures that the message has only the given partner ID.  This
// will always set the partner ID, replacing any existing partner IDs.
func SetPartnerID(partnerID string) NormifyOption {
	return optionFunc(func(m *Message) error {
		m.PartnerIDs = []string{partnerID}
		return nil
	})
}

// SetSessionID ensures that the message has the given session ID.  This will
// always set the session ID, replacing any existing session ID
func SetSessionID(sessionID string) NormifyOption {
	return optionFunc(func(m *Message) error {
		m.SessionID = sessionID
		return nil
	})
}

// EnsureMetadataString ensures that the message has the given string metadata.
// This will always set the metadata.
func EnsureMetadataString(key, value string) NormifyOption {
	return optionFunc(func(m *Message) error {
		if m.Metadata == nil {
			m.Metadata = make(map[string]string)
		}
		m.Metadata[key] = value
		return nil
	})
}

// EnsureMetadataTime ensures that the message has the given time metadata.
// This will always set the metadata.  The time is formatted using RFC3339.
func EnsureMetadataTime(key string, t time.Time) NormifyOption {
	return EnsureMetadataString(key, t.Format(time.RFC3339))
}

// EnsureMetadataInt64 ensures that the message has the given integer metadata.
// This will always set the metadata.  The integer is converted to a string
// using base 10.
func EnsureMetadataInt64(key string, i int64) NormifyOption {
	return EnsureMetadataString(key, strconv.FormatInt(i, 10))
}

// -- Validators ---------------------------------------------------------------

// ValidateSource ensures that the source locator is valid.
func ValidateSource() NormifyOption {
	return optionFunc(func(m *Message) error {
		if _, err := ParseLocator(m.Source); err != nil {
			return errors.Join(err, ErrInvalidSource)
		}
		return nil
	})
}

// ValidateDestination ensures that the destination locator is valid.
func ValidateDestination() NormifyOption {
	return optionFunc(func(m *Message) error {
		if _, err := ParseLocator(m.Destination); err != nil {
			return errors.Join(err, ErrInvalidDest)
		}
		return nil
	})
}

// ValidateMessageType ensures that the message type is valid.
func ValidateMessageType() NormifyOption {
	return optionFunc(func(m *Message) error {
		if m.Type <= Invalid1MessageType || m.Type >= LastMessageType {
			return ErrInvalidMessageType
		}
		return nil
	})
}

// ValidateOnlyUTF8Strings ensures that all string fields in the message are
// valid UTF-8.
func ValidateOnlyUTF8Strings() NormifyOption {
	return optionFunc(func(m *Message) error {
		if err := UTF8(m); err != nil {
			return errors.Join(err, ErrInvalidString)
		}
		return nil
	})
}

// ValidateIsPartner ensures that the message has the given partner ID.
func ValidateIsPartner(partner string) NormifyOption {
	return optionFunc(func(m *Message) error {
		list := m.TrimmedPartnerIDs()
		if len(list) != 1 || list[0] != partner {
			return ErrInvalidPartnerID
		}

		return nil
	})
}

// ValidateHasPartner ensures that the message has one of the given partner
// IDs.
func ValidateHasPartner(partners ...string) NormifyOption {
	trimmed := make([]string, 0, len(partners))
	for _, p := range partners {
		if p != "" {
			trimmed = append(trimmed, p)
		}
	}

	return optionFunc(func(m *Message) error {
		list := m.TrimmedPartnerIDs()
		for _, p := range trimmed {
			for _, id := range list {
				if id == p {
					return nil
				}
			}
		}
		return ErrInvalidPartnerID
	})
}
