// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

// Is reports whether the msg is the same type as the target, or is convertible
// to the target.
//
// If the msg is a *Message, it will be validated against the provided
// validators.  If the msg is not a *Message, the validators will be ignored.
// If the validators are not provided, the msg will be validated against the
// default validators.  To skip validation, provide the NoStandardValidation()
// as a validator.
func Is(msg, target Union, validators ...Processor) bool {
	if msg == nil || target == nil {
		return msg == target
	}

	msgType := msg.MsgType()
	targetType := target.MsgType()

	if !msgType.IsValid() || !targetType.IsValid() || msgType != targetType {
		return false
	}

	if m, ok := msg.(*Message); ok {
		if err := m.Validate(validators...); err != nil {
			return false
		}
	}

	return true
}

// As converts the msg into the target, if possible.
//
// One of the msg or target must be a *Message.
//
// When the msg is a *Message, it will be validated against the provided
// validators.  If the validators are not provided, the msg will be validated
// against the default validators.  To skip validation, provide the
// NoStandardValidation() as a validator.  If the message is not convertible to
// the target, an error will be returned.
func As(msg, target Union, validators ...Processor) error {
	if msg == nil || target == nil {
		if msg == target {
			return nil
		}
		return ErrInvalidMessageType
	}

	msgType := msg.MsgType()
	targetType := target.MsgType()

	if !msgType.IsValid() || !targetType.IsValid() || msgType != targetType {
		return ErrInvalidMessageType
	}

	if m, ok := msg.(*Message); ok {
		return target.From(m, validators...)
	} else if s, ok := target.(*Message); ok {
		return s.To(s, validators...)
	}
	return ErrInvalidMessageType
}
