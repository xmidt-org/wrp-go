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
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestInvalidUtf8Decoding(t *testing.T) {
	assert := assert.New(t)

	/*
		"\x85"  - 5 name value pairs
			"\xa8""msg_type"         : "\x03" // 3
			"\xa4""dest"             : "\xac""\xed\xbf\xbft-address"
			"\xa7""payload"          : "\xc4""\x03" - len 3
											 "123"
			"\xa6""source"           : "\xae""source-address"
			"\xb0""transaction_uuid" : "\xd9\x24""c07ee5e1-70be-444c-a156-097c767ad8aa"
	*/
	invalid := []byte{
		0x85,
		0xa8, 'm', 's', 'g', '_', 't', 'y', 'p', 'e', 0x03,
		0xa4, 'd', 'e', 's', 't', 0xac /* \xed\xbf\xbf is invalid */, 0xed, 0xbf, 0xbf, 't', '-', 'a', 'd', 'd', 'r', 'e', 's', 's',
		0xa7, 'p', 'a', 'y', 'l', 'o', 'a', 'd', 0xc4, 0x03, '1', '2', '3',
		0xa6, 's', 'o', 'u', 'r', 'c', 'e', 0xae, 's', 'o', 'u', 'r', 'c', 'e', '-', 'a', 'd', 'd', 'r', 'e', 's', 's',
		0xb0, 't', 'r', 'a', 'n', 's', 'a', 'c', 't', 'i', 'o', 'n', '_', 'u', 'u', 'i', 'd', 0xd9, 0x24, 'c', '0', '7', 'e', 'e', '5', 'e', '1', '-', '7', '0', 'b', 'e', '-', '4', '4', '4', 'c', '-', 'a', '1', '5', '6', '-', '0', '9', '7', 'c', '7', '6', '7', 'a', 'd', '8', 'a', 'a',
	}

	decoder := NewDecoderBytes(invalid, Msgpack)
	msg := new(Message)
	err := decoder.Decode(msg)
	assert.Nil(err)
	assert.True(utf8.ValidString(msg.Source))

	assert.False(utf8.ValidString(msg.Destination))
	err = UTF8(msg)
	assert.ErrorIs(err, ErrNotUTF8)
}

func TestUTF8(t *testing.T) {
	type Test struct {
		unexported string
		Name       string
		Age        int
	}

	testVal := Test{
		unexported: "this shouldn't be output",
		Name:       "Joe Schmoe",
		Age:        415,
	}

	tests := []struct {
		description string
		value       interface{}
		expectedErr error
	}{
		{
			description: "Success",
			value:       testVal,
		},
		{
			description: "Pointer success",
			value:       &testVal,
		},
		{
			description: "Non struct error",
			value:       5,
			expectedErr: ErrUnexpectedKind,
		},
		{
			description: "UTF8 error",
			value: Test{
				Name: string([]byte{0xbf}),
			},
			expectedErr: ErrNotUTF8,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := UTF8(tc.value)
			if tc.expectedErr == nil {
				assert.NoError(err)
				return
			}
			assert.ErrorIs(err, tc.expectedErr)
		})
	}
}
