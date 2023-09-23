// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrphttp

import (
	"bytes"
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/wrp-go/v3"
	"github.com/xmidt-org/wrp-go/v3/wrpcontext"
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
		t.Run(record.name, func(t *testing.T) {
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

func testDecodeRequestSuccess(t *testing.T) {
	testData := []struct {
		bodyFormat       wrp.Format
		contentType      string
		accept           string
		msgType          wrp.MessageType
		msgTypeString    string
		httpHeaderFormat bool
	}{
		{ // Msgpack
			bodyFormat:       wrp.Msgpack,
			contentType:      wrp.Msgpack.ContentType(),
			accept:           wrp.Msgpack.ContentType(),
			httpHeaderFormat: false,
		},
		{ // JSON
			bodyFormat:       wrp.JSON,
			contentType:      wrp.JSON.ContentType(),
			accept:           wrp.JSON.ContentType(),
			httpHeaderFormat: false,
		},
		{ // HTTP Header Format
			bodyFormat:       wrp.JSON,
			contentType:      wrp.JSON.ContentType(),
			accept:           wrp.JSON.ContentType(),
			msgType:          wrp.SimpleEventMessageType,
			msgTypeString:    "SimpleEvent",
			httpHeaderFormat: true,
		},
		{ // HTTP Header Format, Simple Request Response
			bodyFormat:       wrp.JSON,
			contentType:      wrp.JSON.ContentType(),
			accept:           wrp.JSON.ContentType(),
			msgType:          wrp.SimpleRequestResponseMessageType,
			msgTypeString:    "SimpleRequestResponse",
			httpHeaderFormat: true,
		},
	}

	for _, record := range testData {
		var (
			assert   = assert.New(t)
			require  = require.New(t)
			expected = &wrp.Message{
				ContentType: record.contentType,
			}
			body []byte
		)

		require.NoError(
			wrp.NewEncoderBytes(&body, record.bodyFormat).Encode(&expected),
		)

		request := httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", record.contentType)
		request.Header.Set("Accept", record.accept)

		if record.httpHeaderFormat {
			request.Header.Set(MessageTypeHeader, record.msgTypeString)
			expected.Type = record.msgType
		}

		var msg wrp.Message
		actual, err := DecodeRequest(request, &msg)
		msg, ok := wrpcontext.Get[wrp.Message](actual.Context())

		assert.True(ok)
		assert.Nil(err)
		require.NotNil(actual)
		require.NotNil(actual.Context())
		assert.Equal(expected, &msg)

	}
}

func testDecodeRequestInvalid(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		testData = []struct {
			bodyFormat  wrp.Format
			contentType string
			accept      string
			decodeError bool
		}{
			{
				bodyFormat:  wrp.JSON,
				contentType: "BAD CONTENT TYPE",
				accept:      wrp.JSON.ContentType(),
			},
			{
				bodyFormat:  wrp.JSON,
				contentType: wrp.JSON.ContentType(),
				accept:      wrp.JSON.ContentType(),
				decodeError: true,
			},
		}
	)

	for _, record := range testData {

		var body []byte
		if !record.decodeError {
			expected := &wrp.Message{}
			require.NoError(
				wrp.NewEncoderBytes(&body, record.bodyFormat).Encode(&expected),
			)

		}

		request := httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", record.contentType)
		request.Header.Set("Accept", record.accept)

		var msg wrp.Message
		actual, err := DecodeRequest(request, &msg)

		assert.Error(err)
		require.Nil(actual)

	}
}

func TestDecodeRequest(t *testing.T) {
	t.Run("Success", testDecodeRequestSuccess)
	t.Run("Invalid", testDecodeRequestInvalid)
}
