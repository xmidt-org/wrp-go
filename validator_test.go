/**
 *  Copyright (c) 2022  Comcast Cable Communications Management, LLC
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func testMsgTypeValidatorValidate(t *testing.T) {
	type Test struct {
		m                map[MessageType]Validators
		defaultValidator Validator
		msg              Message
	}

	var alwaysValidMsg ValidatorFunc = func(msg Message) error { return nil }
	tests := []struct {
		description string
		value       Test
		expectedErr error
	}{
		// Success case
		{
			description: "known message type with successful Validators",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {alwaysValidMsg},
				},
				msg: Message{Type: SimpleEventMessageType},
			},
		},
		{
			description: "unknown message type with provided default Validator",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {alwaysValidMsg},
				},
				defaultValidator: alwaysValidMsg,
				msg:              Message{Type: CreateMessageType},
			},
		},
		// Failure case
		{
			description: "known message type with failing Validators",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {alwaysInvalidMsg()},
				},
				msg: Message{Type: SimpleEventMessageType},
			},
			expectedErr: ErrInvalidMsgType,
		},
		{
			description: "unknown message type without provided default Validator",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {alwaysValidMsg},
				},
				msg: Message{Type: CreateMessageType},
			},
			expectedErr: ErrInvalidMsgType,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			msgv, err := NewMsgTypeValidator(tc.value.m, tc.value.defaultValidator)
			assert.NotNil(msgv)
			assert.Nil(err)
			err = msgv.Validate(tc.value.msg)
			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
				return
			}

			assert.Nil(err)
		})
	}
}

func testNewMsgTypeValidator(t *testing.T) {
	type Test struct {
		m                map[MessageType]Validators
		defaultValidator Validator
	}

	var alwaysValidMsg ValidatorFunc = func(msg Message) error { return nil }
	tests := []struct {
		description string
		value       Test
		expectedErr error
	}{
		// Success case
		{
			description: "with provided default Validator",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {alwaysValidMsg},
				},
				defaultValidator: alwaysValidMsg,
			},
			expectedErr: nil,
		},
		{
			description: "without provided default Validator",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {alwaysValidMsg},
				},
			},
			expectedErr: nil,
		},
		{
			description: "empty list of message type Validators",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {},
				},
				defaultValidator: alwaysValidMsg,
			},
			expectedErr: nil,
		},
		// Failure case
		{
			description: "missing message type Validators",
			value:       Test{},
			expectedErr: ErrInvalidMsgTypeValidator,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			msgv, err := NewMsgTypeValidator(tc.value.m, tc.value.defaultValidator)
			assert.NotNil(msgv)
			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
				return
			}

			assert.Nil(err)
		})
	}
}

func testAlwaysInvalidMsg(t *testing.T) {
	assert := assert.New(t)
	msg := Message{}
	v := alwaysInvalidMsg()

	assert.NotNil(v)
	err := v(msg)

	assert.NotNil(err)
	assert.ErrorIs(err, ErrInvalidMsgType)

}

func TestHelperValidators(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{"alwaysInvalidMsg", testAlwaysInvalidMsg},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func TestMsgTypeValidator(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{"MsgTypeValidator validate", testMsgTypeValidatorValidate},
		{"MsgTypeValidator factory", testNewMsgTypeValidator},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}
