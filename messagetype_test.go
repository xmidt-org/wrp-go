// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mtToStruct = map[MessageType]any{
		Invalid0MessageType:              nil,
		Invalid1MessageType:              nil,
		AuthorizationMessageType:         Authorization{},
		SimpleRequestResponseMessageType: SimpleRequestResponse{},
		SimpleEventMessageType:           SimpleEvent{},
		CreateMessageType:                CRUD{},
		RetrieveMessageType:              CRUD{},
		UpdateMessageType:                CRUD{},
		DeleteMessageType:                CRUD{},
		ServiceRegistrationMessageType:   ServiceRegistration{},
		ServiceAliveMessageType:          ServiceAlive{},
		UnknownMessageType:               Unknown{},
		LastMessageType:                  nil,
	}
)

func TestMessageTypeString(t *testing.T) {
	var (
		assert       = assert.New(t)
		messageTypes = []MessageType{
			Invalid0MessageType,
			Invalid1MessageType,
			AuthorizationMessageType,
			SimpleRequestResponseMessageType,
			SimpleEventMessageType,
			CreateMessageType,
			RetrieveMessageType,
			UpdateMessageType,
			DeleteMessageType,
			ServiceRegistrationMessageType,
			ServiceAliveMessageType,
			UnknownMessageType,
			MessageType(-1),
		}

		strings = make(map[string]bool, len(messageTypes))
	)

	for _, messageType := range messageTypes {
		stringValue := messageType.String()
		assert.NotEmpty(stringValue)

		assert.NotContains(strings, stringValue)
		strings[stringValue] = true
	}

	assert.Equal(len(messageTypes), len(strings))
}

func testStringToMessageTypeValid(t *testing.T, expected MessageType) {
	var (
		assert         = assert.New(t)
		expectedString = expected.String()
	)

	actual := StringToMessageType(expectedString)
	assert.Equal(expected, actual)

	actual = StringToMessageType(expectedString[0 : len(expectedString)-len("MessageType")])
	assert.Equal(expected, actual)

	actual = StringToMessageType(strconv.Itoa(int(expected)))
	assert.Equal(expected, actual)
}

func testStringToMessageTypeInvalid(t *testing.T, invalid string) {
	assert := assert.New(t)

	actual := StringToMessageType(invalid)
	assert.Equal(LastMessageType, actual)
}

func TestStringToMessageType(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for v := SimpleRequestResponseMessageType; v < LastMessageType; v++ {
			testStringToMessageTypeValid(t, v)
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		for _, v := range []string{"-1", "", "    ", "a;slkdfja;ksjdf"} {
			testStringToMessageTypeInvalid(t, v)
		}
	})
}

func TestMtToStructContainsAllMessageTypes(t *testing.T) {
	for mt := Invalid0MessageType; mt <= LastMessageType; mt++ {
		_, found := mtToStruct[mt]
		assert.True(t, found, "mtToStruct should contain MessageType %v", mt)
	}
}

func TestSupportsQOSAck(t *testing.T) {
	for msgType, specificStruct := range mtToStruct {
		if specificStruct == nil {
			continue
		}

		// Check if the struct contains the QualityOfService field
		structType := reflect.TypeOf(specificStruct)
		_, found := structType.FieldByName("QualityOfService")

		if found {
			assert.True(t, msgType.SupportsQOSAck(), "MessageType %v should support QOS Ack", msgType)
		} else {
			assert.False(t, msgType.SupportsQOSAck(), "MessageType %v should not support QOS Ack", msgType)
		}
	}
}

func TestMessageTypeSupportsTransaction(t *testing.T) {
	for msgType, specificStruct := range mtToStruct {
		if specificStruct == nil {
			continue
		}

		// Check if the struct contains the QualityOfService field
		structType := reflect.TypeOf(specificStruct)
		field, found := structType.FieldByName("TransactionUUID")
		if !found {
			assert.Equal(t, found, msgType.RequiresTransaction(), "MessageType %v should not require a transaction", msgType)
			continue
		}

		_, required := field.Tag.Lookup("required")
		if required {
			assert.True(t, msgType.RequiresTransaction(), "MessageType %v should require a transaction", msgType)
		} else {
			assert.False(t, msgType.RequiresTransaction(), "MessageType %v should not require a transaction", msgType)
		}
	}
}
