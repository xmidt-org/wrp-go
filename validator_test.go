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
		m                 map[MessageType]Validator
		defaultValidators Validators
		msg               Message
	}

	tests := []struct {
		description string
		value       Test
		expectedErr error
	}{
		// Success case
		{
			description: "Found success",
			value: Test{
				m: map[MessageType]Validator{
					SimpleEventMessageType: Validators{AlwaysValid},
				},
				msg: Message{Type: SimpleEventMessageType},
			},
		},
		{
			description: "Unfound success",
			value: Test{
				m: map[MessageType]Validator{
					SimpleEventMessageType: Validators{AlwaysInvalid},
				},
				defaultValidators: Validators{AlwaysValid},
				msg:               Message{Type: CreateMessageType},
			},
		},
		// Failure case
		{
			description: "Found error",
			value: Test{
				m: map[MessageType]Validator{
					SimpleEventMessageType: Validators{AlwaysInvalid},
				},
				defaultValidators: Validators{AlwaysValid},
				msg:               Message{Type: SimpleEventMessageType},
			},
			expectedErr: ErrInvalidMsgType,
		},
		{
			description: "Unfound error",
			value: Test{
				m: map[MessageType]Validator{
					SimpleEventMessageType: Validators{AlwaysValid},
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
		m                 map[MessageType]Validator
		defaultValidators Validators
	}

	tests := []struct {
		description string
		value       Test
		expectedErr error
	}{
		// Success case
		{
			description: "Default Validators success",
			value: Test{
				m: map[MessageType]Validator{
					SimpleEventMessageType: AlwaysValid,
				},
				defaultValidators: Validators{AlwaysValid},
			},
			expectedErr: nil,
		},
		{
			description: "Empty map of Validators success",
			value: Test{
				m:                 map[MessageType]Validator{},
				defaultValidators: Validators{AlwaysValid},
			},
			expectedErr: nil,
		},
		{
			description: "Omit default Validators success",
			value: Test{
				m: map[MessageType]Validator{
					SimpleEventMessageType: Validators{AlwaysValid},
				},
			},
			expectedErr: nil,
		},
		// Failure case
		{
			description: "Nil list of default Validators error",
			value: Test{
				m: map[MessageType]Validator{
					SimpleEventMessageType: AlwaysValid,
				},
				defaultValidators: Validators{nil},
			},
			expectedErr: ErrInvalidTypeValidator,
		},
		{
			description: "Empty list of Validators error",
			value: Test{
				m: map[MessageType]Validator{
					SimpleEventMessageType: Validators{},
				},
				defaultValidators: Validators{AlwaysValid},
			},
			expectedErr: ErrInvalidTypeValidator,
		},
		{
			description: "Nil Validator error",
			value: Test{
				m: map[MessageType]Validator{
					SimpleEventMessageType: nil,
				},
				defaultValidators: Validators{AlwaysValid},
			},
			expectedErr: ErrInvalidTypeValidator,
		},
		{
			description: "Nil list of Validators error",
			value: Test{
				m: map[MessageType]Validator{
					SimpleEventMessageType: Validators{nil},
				},
				defaultValidators: Validators{AlwaysValid},
			},
			expectedErr: ErrInvalidTypeValidator,
		},
		{
			description: "Nil map of Validators error",
			value: Test{
				m:                 nil,
				defaultValidators: Validators{AlwaysValid},
			},
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

func testAlwaysValid(t *testing.T) {
	assert := assert.New(t)
	msg := Message{}
	err := AlwaysValid(msg)
	assert.NoError(err)
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
		{"AlwaysValid", testAlwaysValid},
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
	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[MessageType]Validator{SimpleEventMessageType: Validators{AlwaysValid}},
		// Validates unfound msg types
		AlwaysInvalid)
	fmt.Printf("%v %T", err == nil, msgv)
	// Output: true wrp.TypeValidator
}

func ExampleTypeValidator_Validate() {
	msgv, err := NewTypeValidator(
		// Validates found msg types
		map[MessageType]Validator{SimpleEventMessageType: Validators{AlwaysValid}},
		// Validates unfound msg types
		AlwaysInvalid)
	if err != nil {
		return
	}

	foundErr := msgv.Validate(Message{Type: SimpleEventMessageType}) // Found success
	unfoundErr := msgv.Validate(Message{Type: CreateMessageType})    // Unfound error
	fmt.Println(foundErr == nil, unfoundErr == nil)
	// Output: true false
}
