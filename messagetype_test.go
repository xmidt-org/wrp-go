// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestMessageTypeSupportsTransaction(t *testing.T) {
	var (
		assert                      = assert.New(t)
		expectedSupportsTransaction = map[MessageType]bool{
			Invalid0MessageType:              false,
			Invalid1MessageType:              false,
			AuthorizationMessageType:         false,
			SimpleRequestResponseMessageType: true,
			SimpleEventMessageType:           false,
			CreateMessageType:                true,
			RetrieveMessageType:              true,
			UpdateMessageType:                true,
			DeleteMessageType:                true,
			ServiceRegistrationMessageType:   false,
			ServiceAliveMessageType:          false,
			UnknownMessageType:               false,
			lastMessageType:                  false,
			lastMessageType + 1:              false,
		}
	)

	for messageType, expected := range expectedSupportsTransaction {
		assert.Equal(expected, messageType.RequiresTransaction())
	}
}

func TestMessageTypeSupportsQOSAck(t *testing.T) {
	var (
		assert                 = assert.New(t)
		expectedSupportsQOSAck = map[MessageType]bool{
			Invalid0MessageType:              false,
			Invalid1MessageType:              false,
			AuthorizationMessageType:         false,
			SimpleRequestResponseMessageType: false,
			SimpleEventMessageType:           true,
			CreateMessageType:                false,
			RetrieveMessageType:              false,
			UpdateMessageType:                false,
			DeleteMessageType:                false,
			ServiceRegistrationMessageType:   false,
			ServiceAliveMessageType:          false,
			UnknownMessageType:               false,
			lastMessageType:                  false,
			lastMessageType + 1:              false,
		}
	)

	for messageType, expected := range expectedSupportsQOSAck {
		assert.Equal(expected, messageType.SupportsQOSAck())
	}
}

func testStringToMessageTypeValid(t *testing.T, expected MessageType) {
	var (
		assert         = assert.New(t)
		expectedString = expected.String()
	)

	actual, err := StringToMessageType(expectedString)
	assert.Equal(expected, actual)
	assert.NoError(err)

	actual, err = StringToMessageType(expectedString[0 : len(expectedString)-len("MessageType")])
	assert.Equal(expected, actual)
	assert.NoError(err)

	actual, err = StringToMessageType(strconv.Itoa(int(expected)))
	assert.Equal(expected, actual)
	assert.NoError(err)
}

func testStringToMessageTypeInvalid(t *testing.T, invalid string) {
	assert := assert.New(t)

	actual, err := StringToMessageType(invalid)
	assert.Equal(MessageType(-1), actual)
	assert.Error(err)
}

func TestStringToMessageType(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for v := SimpleRequestResponseMessageType; v < lastMessageType; v++ {
			testStringToMessageTypeValid(t, v)
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		for _, v := range []string{"-1", "", "    ", "a;slkdfja;ksjdf"} {
			testStringToMessageTypeInvalid(t, v)
		}
	})
}
