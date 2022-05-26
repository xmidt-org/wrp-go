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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTypeValidatorValidate(t *testing.T) {
	type Test struct {
		m                 map[MessageType]Validators
		defaultValidators Validators
		msg               Message
	}

	var alwaysValid ValidatorFunc = func(msg Message) error { return nil }
	tests := []struct {
		description string
		value       Test
		expectedErr error
	}{
		// Success case
		{
			description: "Found success",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {alwaysValid},
				},
				msg: Message{Type: SimpleEventMessageType},
			},
		},
		{
			description: "Unfound success",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {AlwaysInvalid},
				},
				defaultValidators: Validators{alwaysValid},
				msg:               Message{Type: CreateMessageType},
			},
		},
		// Failure case
		{
			description: "Found error",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {AlwaysInvalid},
				},
				defaultValidators: Validators{alwaysValid},
				msg:               Message{Type: SimpleEventMessageType},
			},
			expectedErr: ErrInvalidMsgType,
		},
		{
			description: "Unfound error",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {alwaysValid},
				},
				msg: Message{Type: CreateMessageType},
			},
			expectedErr: ErrInvalidMsgType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			msgv, err := NewTypeValidator(tc.value.m, tc.value.defaultValidators...)
			require.NotNil(msgv)
			require.NoError(err)
			err = msgv.Validate(tc.value.msg)
			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
				return
			}

			assert.NoError(err)
		})
	}
}

func testNewTypeValidator(t *testing.T) {
	type Test struct {
		m                 map[MessageType]Validators
		defaultValidators Validators
	}

	var alwaysValid ValidatorFunc = func(msg Message) error { return nil }
	tests := []struct {
		description string
		value       Test
		expectedErr error
	}{
		// Success case
		{
			description: "Default Validators success",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {alwaysValid},
				},
				defaultValidators: Validators{alwaysValid},
			},
			expectedErr: nil,
		},
		{
			description: "Empty map of Validators success",
			value: Test{
				m:                 map[MessageType]Validators{},
				defaultValidators: Validators{alwaysValid},
			},
			expectedErr: nil,
		},
		{
			description: "Omit default Validators success",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {alwaysValid},
				},
			},
			expectedErr: nil,
		},
		// Failure case
		{
			description: "Nil default Validators",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {alwaysValid},
				},
				defaultValidators: Validators{nil},
			},
			expectedErr: ErrInvalidTypeValidator,
		},
		{
			description: "Empty list of Validators error",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {},
				},
				defaultValidators: Validators{alwaysValid},
			},
			expectedErr: ErrInvalidTypeValidator,
		},
		{
			description: "Nil Validators error",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: nil,
				},
				defaultValidators: Validators{alwaysValid},
			},
			expectedErr: ErrInvalidTypeValidator,
		},
		{
			description: "Nil list of Validators error",
			value: Test{
				m: map[MessageType]Validators{
					SimpleEventMessageType: {nil},
				},
				defaultValidators: Validators{alwaysValid},
			},
			expectedErr: ErrInvalidTypeValidator,
		},
		{
			description: "Empty map of Validators error",
			value:       Test{},
			expectedErr: ErrInvalidTypeValidator,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			msgv, err := NewTypeValidator(tc.value.m, tc.value.defaultValidators...)
			assert.NotNil(msgv)
			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
				return
			}

			assert.NoError(err)
		})
	}
}

func testAlwaysInvalid(t *testing.T) {
	assert := assert.New(t)
	msg := Message{}
	err := AlwaysInvalid(msg)

	assert.ErrorIs(err, ErrInvalidMsgType)

}

func TestHelperValidators(t *testing.T) {
	tests := []struct {
		description string
		test        func(*testing.T)
	}{
		{"AlwaysInvalid", testAlwaysInvalid},
	}

	for _, tc := range tests {
		t.Run(tc.description, tc.test)
	}
}

func TestTypeValidator(t *testing.T) {
	tests := []struct {
		description string
		test        func(*testing.T)
	}{
		{"TypeValidator validate", testTypeValidatorValidate},
		{"TypeValidator factory", testNewTypeValidator},
	}

	for _, tc := range tests {
		t.Run(tc.description, tc.test)
	}
}

func ExampleNewTypeValidator() {
	var alwaysValid ValidatorFunc = func(msg Message) error { return nil }
	msgv, err := NewTypeValidator(
		// Validates known msg types
		map[MessageType]Validators{SimpleEventMessageType: {alwaysValid}},
		// Validates unknown msg types
		AlwaysInvalid)

	fmt.Printf("%v, %T", err == nil, msgv)
	// Output: true, wrp.TypeValidator

}

func ExampleTypeValidator_Validate() {
	var alwaysValid ValidatorFunc = func(msg Message) error { return nil }
	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[MessageType]Validators{SimpleEventMessageType: {alwaysValid}},
		// Validates unfound msg types
		AlwaysInvalid)
	if err != nil {
		return
	}
	foundErr := msgv.Validate(Message{Type: SimpleEventMessageType}) // Found success
	unfoundErr := msgv.Validate(Message{Type: CreateMessageType})    // Unfound error
	fmt.Printf("foundErr is nil: %v, unfoundErr is nil: %v", foundErr == nil, unfoundErr == nil)
	// Output: foundErr is nil: true, unfoundErr is nil: false
}
