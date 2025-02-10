// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"context"
	"strconv"
	"time"
)

func handleSelf(me Locator, target *string) error {
	if target == nil || *target == "" {
		return nil
	}

	l, err := ParseLocator(*target)
	if err != nil {
		return err
	}
	if l.IsSelf() {
		l.Scheme = me.Scheme
		l.Authority = me.Authority
		*target = l.String()
	}

	return nil
}

// ReplaceSelfDestinationLocator replaces the destination of the message with the
// given locator if the destination is a `self:` based locator.  ErrNotHandled
// is returned along with the latest verion of the message, unless the given
// locator is not valid, the option returns that  error.
func ReplaceSelfDestinationLocator(me Locator) Modifier {
	return ModifierFunc(func(_ context.Context, msg Message) (Message, error) {
		err := handleSelf(me, &msg.Destination)
		if err != nil {
			return Message{}, err
		}
		return msg, ErrNotHandled
	})
}

// ReplaceSelfSourceLocator replaces the source of the message with the
// given locator if the destination is a `self:` based locator.  ErrNotHandled
// is returned along with the latest verion of the message, unless the given
// locator is not valid, the option returns that  error.
func ReplaceSelfSourceLocator(me Locator) Modifier {
	return ModifierFunc(func(_ context.Context, msg Message) (Message, error) {
		err := handleSelf(me, &msg.Source)
		if err != nil {
			return Message{}, err
		}
		return msg, ErrNotHandled
	})
}

// ReplaceAnySelfLocator replaces any `self:` based locator with the scheme and
// authority of the given locator.  If the given locator is not valid, the
// option returns an error.  ErrNotHandled is returned unless the format of the
// locator found in the message is invalid.  Then that error is returned.
func ReplaceAnySelfLocator(me Locator) Modifier {
	return Modifiers{
		ReplaceSelfDestinationLocator(me),
		ReplaceSelfSourceLocator(me),
	}
}

// MetadataValue is a type constraint that allows only string, int64, and time.Time types.
type MetadataValue interface {
	bool |
		string |
		int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 | uintptr |
		float32 | float64 |
		complex64 | complex128 |
		time.Time | time.Duration
}

// EnsureMetadata ensures that the message has the given metadata key and value.
// This will always set the metadata, replacing any existing value.
// ErrNotHandled is always returned along with the updated message.
func EnsureMetadata[T MetadataValue](key string, value T) Modifier {
	return ModifierFunc(func(_ context.Context, msg Message) (Message, error) {
		if msg.Metadata == nil {
			msg.Metadata = make(map[string]string)
		}

		switch v := any(value).(type) {
		case bool:
			msg.Metadata[key] = strconv.FormatBool(v)
		case string:
			msg.Metadata[key] = v
		case int:
			msg.Metadata[key] = strconv.FormatInt(int64(v), 10)
		case int8:
			msg.Metadata[key] = strconv.FormatInt(int64(v), 10)
		case int16:
			msg.Metadata[key] = strconv.FormatInt(int64(v), 10)
		case int32:
			msg.Metadata[key] = strconv.FormatInt(int64(v), 10)
		case int64:
			msg.Metadata[key] = strconv.FormatInt(v, 10)
		case uint:
			msg.Metadata[key] = strconv.FormatUint(uint64(v), 10)
		case uint8:
			msg.Metadata[key] = strconv.FormatUint(uint64(v), 10)
		case uint16:
			msg.Metadata[key] = strconv.FormatUint(uint64(v), 10)
		case uint32:
			msg.Metadata[key] = strconv.FormatUint(uint64(v), 10)
		case uint64:
			msg.Metadata[key] = strconv.FormatUint(v, 10)
		case uintptr:
			msg.Metadata[key] = strconv.FormatUint(uint64(v), 10)
		case float32:
			msg.Metadata[key] = strconv.FormatFloat(float64(v), 'f', -1, 32)
		case float64:
			msg.Metadata[key] = strconv.FormatFloat(v, 'f', -1, 64)
		case complex64:
			msg.Metadata[key] = strconv.FormatComplex(complex128(v), 'f', -1, 64)
		case complex128:
			msg.Metadata[key] = strconv.FormatComplex(v, 'f', -1, 128)
		case time.Time:
			msg.Metadata[key] = v.Format(time.RFC3339)
		case time.Duration:
			msg.Metadata[key] = v.String()
		}

		return msg, ErrNotHandled
	})
}
