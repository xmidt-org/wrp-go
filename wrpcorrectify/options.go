// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpcorrectify

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/xmidt-org/wrp-go/v3"
)

// ErrorOption returns an option that always returns the given error.
func ErrorOption(err error) Option {
	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
		return err
	})
}

// Options returns a new option that applies all of the given options in order.
func Options(opts ...Option) Option {
	return OptionFunc(func(ctx context.Context, m *wrp.Message) error {
		for _, opt := range opts {
			if opt != nil {
				if err := opt.Correctify(ctx, m); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// ReplaceAnySelfLocator replaces any `self:` based locator with the scheme and
// authority of the given locator.  If the given locator is not valid, the
// option returns an error.
func ReplaceAnySelfLocator(me string) Option {
	return Options(
		ReplaceSourceSelfLocator(me),
		ReplaceDestinationSelfLocator(me),
	)
}

// ReplaceSourceSelfLocator replaces a `self:` based source locator with the
// scheme and authority of the given locator.  If the given locator is not valid,
// the option returns an error.
func ReplaceSourceSelfLocator(me string) Option {
	full, err := wrp.ParseLocator(me)
	if err != nil {
		return ErrorOption(err)
	}

	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
		src, err := wrp.ParseLocator(m.Source)
		if err != nil {
			return err
		}

		if src.Scheme == "self" {
			src.Scheme = full.Scheme
			m.Source = src.String()
		}

		return nil
	})
}

// ReplaceDestinationSelfLocator replaces the destination of the message with the
// given locator if the destination is a `self:` based locator.  If the given
// locator is not valid, the option returns an error.
func ReplaceDestinationSelfLocator(me string) Option {
	full, err := wrp.ParseLocator(me)
	if err != nil {
		return ErrorOption(err)
	}

	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
		dst, err := wrp.ParseLocator(m.Destination)
		if err != nil {
			return err
		}

		if dst.Scheme == "self" {
			dst.Scheme = full.Scheme
			m.Destination = dst.String()
		}

		return nil
	})
}

// EnsureTransactionUUID ensures that the message has a transaction UUID.  If
// the message does not have a transaction UUID, a new one is generated.
func EnsureTransactionUUID() Option {
	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
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

// EnsurePartnerID ensures that the message has the given partner ID in
// the list.
func EnsurePartnerID(partnerID string) Option {
	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
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

// SetPartnerID ensures that the message has only the given partner ID.
func SetPartnerID(partnerID string) Option {
	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
		if m.PartnerIDs == nil {
			m.PartnerIDs = make([]string, 0, 1)
		}
		m.PartnerIDs = append(m.PartnerIDs, partnerID)
		return nil
	})
}

// SetSessionID ensures that the message has the given session ID.
func SetSessionID(sessionID string) Option {
	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
		m.SessionID = sessionID
		return nil
	})
}

// EnsureMetadataString ensures that the message has the given string metadata.
func EnsureMetadataString(key, value string) Option {
	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
		if m.Metadata == nil {
			m.Metadata = make(map[string]string)
		}
		m.Metadata[key] = value
		return nil
	})
}

// EnsureMetadataTime ensures that the message has the given time metadata.
func EnsureMetadataTime(key string, t time.Time) Option {
	return EnsureMetadataString(key, t.Format(time.RFC3339))
}

// EnsureMetadataInt ensures that the message has the given integer metadata.
func EnsureMetadataInt(key string, i int64) Option {
	return EnsureMetadataString(key, strconv.FormatInt(i, 10))
}

// ValidateSource ensures that the source locator is valid.
func ValidateSource() Option {
	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
		if _, err := wrp.ParseLocator(m.Source); err != nil {
			return err
		}
		return nil
	})
}

// ValidateDestination ensures that the destination locator is valid.
func ValidateDestination() Option {
	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
		if _, err := wrp.ParseLocator(m.Destination); err != nil {
			return err
		}
		return nil
	})
}

func ValidateMessageType() Option {
	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
		if m.Type <= wrp.Invalid1MessageType || m.Type >= wrp.LastMessageType {
			return ErrorInvalidMessageType
		}
		return nil
	})
}

func ValidateOnlyUTF8Strings() Option {
	return OptionFunc(func(_ context.Context, m *wrp.Message) error {
		if err := wrp.UTF8(m); err != nil {
			return err
		}
		return nil
	})
}
