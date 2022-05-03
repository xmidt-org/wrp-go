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

package wrphttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/wrp-go/v3"
)

func TestDefaultDecoder(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(DefaultDecoder())
}

func testDecodeEntitySuccess(t *testing.T) {
	testData := []struct {
		defaultFormat wrp.Format
		bodyFormat    wrp.Format
		contentType   string
		accept        string
		name          string
	}{
		// Default test cases
		{wrp.Msgpack, wrp.Msgpack, "", "", "default WRP headers"},
		{wrp.JSON, wrp.JSON, "", "", "default JSON headers"},
		// Accept header test cases
		{wrp.JSON, wrp.JSON, "", wrp.JSON.ContentType(), "JSON Accept header"},
		{wrp.JSON, wrp.JSON, "", wrp.Msgpack.ContentType(), "WRP Accept header"},
		// Content-Type header test cases
		{wrp.Msgpack, wrp.JSON, wrp.JSON.ContentType(), "", "JSON Content-Type header"},
		{wrp.JSON, wrp.Msgpack, wrp.Msgpack.ContentType(), "", "WRP Content-Type header"},
	}

	for _, record := range testData {
		t.Run(record.name, func(t *testing.T) {
			var (
				assert  = assert.New(t)
				require = require.New(t)

				expected = wrp.Message{
					Source:      "foo",
					Destination: "bar",
				}

				body    []byte
				decoder = DecodeEntity(record.defaultFormat)
			)

			require.NotNil(decoder)

			require.NoError(
				wrp.NewEncoderBytes(&body, record.bodyFormat).Encode(&expected),
			)

			request := httptest.NewRequest("POST", "/", bytes.NewBuffer(body))

			request.Header.Set("Content-Type", record.contentType)
			request.Header.Set("Accept", record.accept)
			entity, err := decoder(context.Background(), request)
			assert.NoError(err)
			require.NotNil(entity)

			assert.Equal(expected, entity.Message)
			assert.Equal(record.bodyFormat, entity.Format)
			assert.Equal(body, entity.Bytes)
		})
	}
}

func testDecodeEntityInvalidContentType(t *testing.T) {
	testData := []struct {
		contentType string
		accept      string
		name        string
	}{
		// Content-Type header test cases
		{"invalid", "", "Content-Type Header"},
		// Accept Header test cases
		{"", "invalid", "Accept Header"},
	}

	for _, record := range testData {
		t.Run(fmt.Sprintf("test case: %v", record.name), func(t *testing.T) {
			var (
				assert  = assert.New(t)
				require = require.New(t)

				decoder = DecodeEntity(wrp.Msgpack)
				request = httptest.NewRequest("GET", "/", nil)
			)
			require.NotNil(decoder)
			request.Header.Set("Content-Type", record.contentType)
			request.Header.Set("Accept", record.accept)
			entity, err := decoder(context.Background(), request)
			assert.Nil(entity)
			assert.Error(err)
		})
	}
}

func testDecodeEntityBodyError(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expectedError = errors.New("failed to decode wrp: EOF")
		decoder       = DecodeEntity(wrp.Msgpack)
		body          = bytes.NewReader(nil)
		request       = httptest.NewRequest("GET", "/", body)
	)

	require.NotNil(decoder)

	entity, err := decoder(context.Background(), request)

	assert.Nil(entity)
	assert.Equal(expectedError, err)
}

func TestDecodeEntity(t *testing.T) {
	t.Run("Success", testDecodeEntitySuccess)
	t.Run("InvalidContentType", testDecodeEntityInvalidContentType)
	t.Run("BodyError", testDecodeEntityBodyError)
}

func testDecodeRequestHeadersSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expected = wrp.Message{
			Type:            wrp.SimpleEventMessageType,
			Source:          "foo",
			Destination:     "bar",
			ContentType:     wrp.MimeTypeOctetStream,
			Payload:         []byte{1, 2, 3},
			TransactionUUID: "testytest",
		}
		expectedBytes []byte
		body          bytes.Buffer
		request       = httptest.NewRequest("POST", "/", &body)
	)

	require.NoError(
		wrp.NewEncoderBytes(&expectedBytes, wrp.Msgpack).Encode(&expected),
	)

	body.Write([]byte{1, 2, 3})
	request.Header.Set(MessageTypeHeader, "event")
	request.Header.Set(SourceHeader, "foo")
	request.Header.Set(DestinationHeader, "bar")
	request.Header.Set(TransactionUuidHeader, "testytest")
	entity, err := DecodeRequestHeaders(context.Background(), request)
	assert.NoError(err)
	require.NotNil(entity)
	assert.Equal(expected, entity.Message)
	assert.Equal(wrp.Msgpack, entity.Format)
	assert.Equal(expectedBytes, entity.Bytes)
}

func testDecodeRequestHeadersInvalid(t *testing.T) {
	var (
		assert  = assert.New(t)
		request = httptest.NewRequest("POST", "/", nil)
	)

	request.Header.Set(MessageTypeHeader, "askdjfa;skdjfasdf")
	entity, err := DecodeRequestHeaders(context.Background(), request)
	assert.Nil(entity)
	assert.Error(err)
}

func TestDecodeRequestHeaders(t *testing.T) {
	t.Run("Success", testDecodeRequestHeadersSuccess)
	t.Run("Invalid", testDecodeRequestHeadersInvalid)
}
