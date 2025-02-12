// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

type MsgType interface {
	Type() MessageType
	to() *Message
}

func Is(msg, target MsgType, validators ...Processor) bool {
	if msg == nil || target == nil {
		return msg == target
	}

	return msg.Type() == target.Type()
}

func As[T MessageStructs](msg MsgType, target *T, validators ...Processor) bool {
	return false
}
