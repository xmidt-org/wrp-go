// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"strconv"
)

//go:generate go install golang.org/x/tools/cmd/stringer@latest
//go:generate stringer -type=MessageType

// MessageType indicates the kind of WRP message
type MessageType int64

const (
	Invalid0MessageType MessageType = iota
	Invalid1MessageType
	AuthorizationMessageType
	SimpleRequestResponseMessageType
	SimpleEventMessageType
	CreateMessageType
	RetrieveMessageType
	UpdateMessageType
	DeleteMessageType
	ServiceRegistrationMessageType
	ServiceAliveMessageType
	UnknownMessageType
	LastMessageType
)

// RequiresTransaction tests if messages of this type are allowed to participate in transactions.
// If this method returns false, the TransactionUUID field should be ignored (but passed through
// where applicable). If this method returns true, TransactionUUID must be included in request.
func (mt MessageType) RequiresTransaction() bool {
	switch mt {
	case SimpleRequestResponseMessageType,
		CreateMessageType,
		RetrieveMessageType,
		UpdateMessageType,
		DeleteMessageType:
		return true
	default:
		return false
	}
}

// SupportsQOSAck tests if messages of this type are allowed to participate in QOS Ack
// as specified in https://xmidt.io/docs/wrp/basics/#qos-description-qos .
// If this method returns false, QOS Ack is foregone.
func (mt MessageType) SupportsQOSAck() bool {
	switch mt {
	case SimpleRequestResponseMessageType,
		SimpleEventMessageType,
		CreateMessageType,
		RetrieveMessageType,
		UpdateMessageType,
		DeleteMessageType:
		return true
	default:
		return false
	}
}

// FriendlyName is just the String version of this type minus the "MessageType" suffix.
// This is used in most textual representations, such as HTTP headers.
func (mt MessageType) FriendlyName() string {
	return friendlyNames[mt]
}

// IsInvalid provides a simple way to see if this MessageType is understood by
// this library and considered valid.
func (mt MessageType) IsValid() bool {
	return Invalid1MessageType < mt && mt < LastMessageType
}

var (
	// stringToMessageType is a simple map of allowed strings which uniquely indicate MessageType values.
	// Included in this map are integral string keys.  Keys are assumed to be case insensitive.
	stringToMessageType map[string]MessageType

	// friendlyNames are the string representations of each message type without the "MessageType" suffix
	friendlyNames map[MessageType]string
)

func init() {
	stringToMessageType = make(map[string]MessageType, LastMessageType-1)
	friendlyNames = make(map[MessageType]string, LastMessageType-1)
	suffixLength := len("MessageType")

	// for each MessageType, allow the following string representations:
	//
	// The integral value of the constant
	// The String() value
	// The String() value minus the MessageType suffix
	for v := SimpleRequestResponseMessageType; v < LastMessageType; v++ {
		stringToMessageType[strconv.Itoa(int(v))] = v

		vs := v.String()
		f := vs[0 : len(vs)-suffixLength]

		stringToMessageType[vs] = v
		stringToMessageType[f] = v
		friendlyNames[v] = f
	}

	stringToMessageType["event"] = SimpleEventMessageType
}

// StringToMessageType converts a string into an enumerated MessageType constant.
// If the value equals the friendly name of a type, e.g. "Auth" for AuthMessageType,
// that type is returned.  Otherwise, the value is converted to an integer and looked up,
// with an error being returned in the event the integer value is not valid.
func StringToMessageType(value string) MessageType {
	mt, ok := stringToMessageType[value]
	if !ok {
		return LastMessageType
	}

	return mt
}
