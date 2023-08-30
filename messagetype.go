/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package wrp

import (
	"fmt"
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
	lastMessageType
)

// RequiresTransaction tests if messages of this type are allowed to participate in transactions.
// If this method returns false, the TransactionUUID field should be ignored (but passed through
// where applicable). If this method returns true, TransactionUUID must be included in request.
func (mt MessageType) RequiresTransaction() bool {
	switch mt {
	case SimpleRequestResponseMessageType, CreateMessageType, RetrieveMessageType, UpdateMessageType, DeleteMessageType:
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
	case SimpleEventMessageType:
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

var (
	// stringToMessageType is a simple map of allowed strings which uniquely indicate MessageType values.
	// Included in this map are integral string keys.  Keys are assumed to be case insensitive.
	stringToMessageType map[string]MessageType

	// friendlyNames are the string representations of each message type without the "MessageType" suffix
	friendlyNames map[MessageType]string
)

func init() {
	stringToMessageType = make(map[string]MessageType, lastMessageType-1)
	friendlyNames = make(map[MessageType]string, lastMessageType-1)
	suffixLength := len("MessageType")

	// for each MessageType, allow the following string representations:
	//
	// The integral value of the constant
	// The String() value
	// The String() value minus the MessageType suffix
	for v := SimpleRequestResponseMessageType; v < lastMessageType; v++ {
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
func StringToMessageType(value string) (MessageType, error) {
	mt, ok := stringToMessageType[value]
	if !ok {
		return MessageType(-1), fmt.Errorf("invalid message type: %s", value)
	}

	return mt, nil
}
