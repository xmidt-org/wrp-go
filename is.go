// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

// Is reports whether the msg is the same type as the target, or is convertible
// to the target.
//
// If the validators are not provided, the msg will be validated against the
// default validators.  To skip validation, provide the NoStandardValidation()
// as a validator.
func Is(msg, target Union, validators ...Processor) bool {
	if msg == nil || target == nil {
		return msg == nil && target == nil
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

	return msg.To(&Message{}, validators...) == nil
}

// As converts the src into the dest, if possible.
//
// The src will be converted into the dest if the types match.  If the src is
// not a *Message, the src will be converted into a *Message first.  Validators
// are then applied to the intermediate *Message.  If the dest is a *Message,
// the type of the dest will be overwritten by the type of the src and is
// ignored for comparison purposes.
//
// If the validators are not provided, the msg will be validated
// against the default validators.  To skip validation, provide the
// NoStandardValidation() as a validator.  If the message is not convertible to
// the target, or validation fails, an error will be returned.
func As(src, dst Union, validators ...Processor) error {
	if src == nil || dst == nil {
		if src == dst {
			return nil
		}
		return ErrInvalidMessageType
	}

	srcType := src.MsgType()
	dstType := dst.MsgType()

	if !srcType.IsValid() {
		return ErrInvalidMessageType
	}

	if d, ok := dst.(*Message); ok {
		// Because the target is the *Message, we don't care what the type is.
		// The type will be overwritten by the source.
		return src.To(d, validators...)
	}

	if !dstType.IsValid() || srcType != dstType {
		return ErrInvalidMessageType
	}

	if s, ok := src.(*Message); ok {
		return dst.From(s, validators...)
	}

	var tmp Message
	_ = src.To(&tmp, NoStandardValidation())

	return dst.From(&tmp, validators...)
}
