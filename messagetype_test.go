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
		}
	)

	for messageType, expected := range expectedSupportsTransaction {
		assert.Equal(expected, messageType.SupportsTransaction())
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
