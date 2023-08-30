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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	gokithttp "github.com/go-kit/kit/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/wrp-go/v3"
)

var foo testContextKey = "foo"

func TestHandlerFunc(t *testing.T) {
	var (
		assert = assert.New(t)

		expectedResponse ResponseWriter = &entityResponseWriter{}
		expectedRequest                 = new(Request)

		called             = false
		hf     HandlerFunc = func(actualResponse ResponseWriter, actualRequest *Request) {
			called = true
			assert.Equal(expectedResponse, actualResponse)
			assert.Equal(expectedRequest, actualRequest)
		}
	)

	hf.ServeWRP(expectedResponse, expectedRequest)
	assert.True(called)
}

func testWithErrorEncoderDefault(t *testing.T) {
	var (
		assert = assert.New(t)
		wh     = new(wrpHandler)
	)

	WithErrorEncoder(nil)(wh)
	assert.NotNil(wh.errorEncoder)
}

func testWithErrorEncoderCustom(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		customCalled                        = false
		custom       gokithttp.ErrorEncoder = func(context.Context, error, http.ResponseWriter) {
			customCalled = true
		}

		wh = new(wrpHandler)
	)

	WithErrorEncoder(custom)(wh)
	require.NotNil(wh.errorEncoder)

	wh.errorEncoder(context.Background(), errors.New("expected"), httptest.NewRecorder())
	assert.True(customCalled)
}

func TestWithErrorEncoder(t *testing.T) {
	t.Run("Default", testWithErrorEncoderDefault)
	t.Run("Custom", testWithErrorEncoderCustom)
}

func testWithNewResponseWriterDefault(t *testing.T) {
	var (
		assert = assert.New(t)
		wh     = new(wrpHandler)
	)

	WithNewResponseWriter(nil)(wh)
	assert.NotNil(wh.newResponseWriter)
}

func testWithNewResponseWriterCustom(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expected                    = &entityResponseWriter{}
		custom   ResponseWriterFunc = func(http.ResponseWriter, *Request) (ResponseWriter, error) {
			return expected, nil
		}

		wh = new(wrpHandler)
	)

	WithNewResponseWriter(custom)(wh)
	require.NotNil(wh.newResponseWriter)

	actual, err := wh.newResponseWriter(httptest.NewRecorder(), new(Request))
	assert.Equal(expected, actual)
	assert.NoError(err)
}

func TestWithNewResponseWriter(t *testing.T) {
	t.Run("Default", testWithNewResponseWriterDefault)
	t.Run("Custom", testWithNewResponseWriterCustom)
}

func testWithDecoderDefault(t *testing.T) {
	var (
		assert = assert.New(t)
		wh     = new(wrpHandler)
	)

	WithDecoder(nil)(wh)
	assert.NotNil(wh.decoder)
}

func testWithDecoderCustom(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expected         = new(Entity)
		custom   Decoder = func(context.Context, *http.Request) (*Entity, error) {
			return expected, nil
		}

		wh = new(wrpHandler)
	)

	WithDecoder(custom)(wh)
	require.NotNil(wh.decoder)

	actual, err := wh.decoder(context.Background(), httptest.NewRequest("GET", "/", nil))
	assert.Equal(expected, actual)
	assert.NoError(err)
}

func TestWithDecoder(t *testing.T) {
	t.Run("Default", testWithDecoderDefault)
	t.Run("Custom", testWithDecoderCustom)
}

func TestWithBefore(t *testing.T) {
	testData := [][]MessageFunc{
		nil,
		[]MessageFunc{},
		[]MessageFunc{
			func(context.Context, *wrp.Message) context.Context { return nil },
		},
		[]MessageFunc{
			func(context.Context, *wrp.Message) context.Context { return nil },
			func(context.Context, *wrp.Message) context.Context { return nil },
			func(context.Context, *wrp.Message) context.Context { return nil },
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				wh     = new(wrpHandler)
			)

			WithBefore(record...)(wh)
			assert.Len(wh.before, len(record))
			WithBefore(record...)(wh)
			assert.Len(wh.before, 2*len(record))
		})
	}
}

func TestWithAfter(t *testing.T) {
	testData := [][]MessageFunc{
		nil,
		[]MessageFunc{},
		[]MessageFunc{
			func(context.Context, *wrp.Message) context.Context { return nil },
		},
		[]MessageFunc{
			func(context.Context, *wrp.Message) context.Context { return nil },
			func(context.Context, *wrp.Message) context.Context { return nil },
			func(context.Context, *wrp.Message) context.Context { return nil },
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				wh     = new(wrpHandler)
			)

			WithAfter(record...)(wh)
			assert.Len(wh.after, len(record))
			WithAfter(record...)(wh)
			assert.Len(wh.after, 2*len(record))
		})
	}
}

func TestNewHTTPHandler(t *testing.T) {
	assert := assert.New(t)
	assert.Panics(func() {
		NewHTTPHandler(nil)
	})
}

func testWRPHandlerDecodeError(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expectedCtx            = context.WithValue(context.Background(), foo, "bar")
		expectedErr            = errors.New("expected")
		expectedHTTPStatusCode = http.StatusBadRequest

		decoder = func(actualCtx context.Context, httpRequest *http.Request) (*Entity, error) {
			assert.Equal(expectedCtx, actualCtx)
			return nil, expectedErr
		}

		errorEncoderCalled = false
		errorEncoder       = func(actualCtx context.Context, actualErr error, _ http.ResponseWriter) {
			errorEncoderCalled = true
			assert.Equal(expectedCtx, actualCtx)
			assert.ErrorIs(actualErr, expectedErr,
				fmt.Errorf("error [%v] doesn't contain error [%v] in its err chain",
					actualErr, expectedErr))

			var actualErrorHTTP httpError
			if assert.ErrorAs(actualErr, &actualErrorHTTP,
				fmt.Errorf("error [%v] doesn't contain error [%v] in its err chain",
					actualErr, actualErrorHTTP)) {
				assert.Equal(expectedHTTPStatusCode, actualErrorHTTP.StatusCode())
			}
		}

		wrpHandler  = new(MockHandler)
		httpHandler = NewHTTPHandler(wrpHandler, WithDecoder(decoder), WithErrorEncoder(errorEncoder))

		httpResponse = httptest.NewRecorder()
		httpRequest  = httptest.NewRequest("POST", "/", nil).WithContext(expectedCtx)
	)

	require.NotNil(httpHandler)
	httpHandler.ServeHTTP(httpResponse, httpRequest)

	assert.True(errorEncoderCalled)
	wrpHandler.AssertExpectations(t)
}

func testTransactionUUIDError(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		msg = wrp.Message{
			Type: wrp.SimpleRequestResponseMessageType,
		}
		msgBytes, _ = json.Marshal(msg)
		entity      = &Entity{
			Message: msg,
			Bytes:   msgBytes,
		}

		expectedError = httpError{
			err:  fmt.Errorf("%s", string(entity.Bytes)),
			code: http.StatusBadRequest,
		}
		decoder = func(_ context.Context, _ *http.Request) (*Entity, error) {
			return entity, nil
		}

		errorEncoderCalled = false
		httpResponse       = httptest.NewRecorder()
	)

	httpRequest := httptest.NewRequest("POST", "/", nil)

	errorEncoder := func(_ context.Context, actualErr error, _ http.ResponseWriter) {
		errorEncoderCalled = true
		assert.Equal(expectedError, actualErr)

		var actualErrorHTTP httpError
		if assert.ErrorAs(actualErr, &actualErrorHTTP,
			fmt.Errorf("error [%v] doesn't contain error [%v] in its err chain",
				actualErr, actualErrorHTTP)) {
			assert.Equal(expectedError.code, actualErrorHTTP.code)
		}
	}

	wrpHandler := new(MockHandler)
	httpHandler := NewHTTPHandler(wrpHandler,
		WithDecoder(decoder),
		WithErrorEncoder(errorEncoder))

	require.NotNil(httpHandler)
	httpHandler.ServeHTTP(httpResponse, httpRequest)

	assert.True(errorEncoderCalled)

}

func testWRPHandlerResponseWriterError(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expectedCtx    = context.WithValue(context.Background(), foo, "bar")
		expectedErr    = errors.New("expected")
		expectedEntity = &Entity{
			Message: wrp.Message{
				Type: wrp.SimpleEventMessageType,
			},
		}

		before = func(ctx context.Context, m *wrp.Message) context.Context {
			m.ContentType = "something"
			return ctx
		}

		decoder = func(actualCtx context.Context, httpRequest *http.Request) (*Entity, error) {
			assert.Equal(expectedCtx, actualCtx)
			return expectedEntity, nil
		}

		newResponseWriterCalled = false
		newResponseWriter       = func(_ http.ResponseWriter, wrpRequest *Request) (ResponseWriter, error) {
			newResponseWriterCalled = true
			assert.Equal(
				wrp.Message{
					Type:        wrp.SimpleEventMessageType,
					ContentType: "something",
				},
				wrpRequest.Entity.Message,
			)

			return nil, expectedErr
		}

		errorEncoderCalled = false
		errorEncoder       = func(actualCtx context.Context, actualErr error, _ http.ResponseWriter) {
			errorEncoderCalled = true
			assert.Equal(expectedCtx, actualCtx)
			assert.Equal(expectedErr, actualErr)
		}

		wrpHandler  = new(MockHandler)
		httpHandler = NewHTTPHandler(wrpHandler,
			WithBefore(before),
			WithDecoder(decoder),
			WithNewResponseWriter(newResponseWriter),
			WithErrorEncoder(errorEncoder),
		)

		httpResponse = httptest.NewRecorder()
		httpRequest  = httptest.NewRequest("POST", "/", nil).WithContext(expectedCtx)
	)

	require.NotNil(httpHandler)
	httpHandler.ServeHTTP(httpResponse, httpRequest)

	assert.True(newResponseWriterCalled)
	assert.True(errorEncoderCalled)
	wrpHandler.AssertExpectations(t)
}

func testWRPHandlerSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expectedCtx    = context.WithValue(context.Background(), foo, "bar")
		expectedEntity = &Entity{
			Message: wrp.Message{
				Type: wrp.SimpleEventMessageType,
			},
		}

		before = func(ctx context.Context, m *wrp.Message) context.Context {
			m.ContentType = "something"
			return ctx
		}

		decoder = func(actualCtx context.Context, httpRequest *http.Request) (*Entity, error) {
			assert.Equal(expectedCtx, actualCtx)
			return expectedEntity, nil
		}

		wrpHandler  = new(MockHandler)
		httpHandler = NewHTTPHandler(wrpHandler,
			WithBefore(before),
			WithDecoder(decoder),
			WithNewResponseWriter(NewEntityResponseWriter(wrp.Msgpack)),
		)

		httpResponse = httptest.NewRecorder()
		httpRequest  = httptest.NewRequest("POST", "/", nil).WithContext(expectedCtx)
	)

	require.NotNil(httpHandler)
	wrpHandler.On("ServeWRP",
		mock.MatchedBy(func(r ResponseWriter) bool {
			return r != nil
		}),
		mock.MatchedBy(func(r *Request) bool {
			return assert.Equal(wrp.Message{Type: wrp.SimpleEventMessageType, ContentType: "something"}, r.Entity.Message)
		}),
	).Once()

	httpHandler.ServeHTTP(httpResponse, httpRequest)
	wrpHandler.AssertExpectations(t)
}

func TestWRPHandler(t *testing.T) {
	t.Run("ServeHTTP", func(t *testing.T) {
		t.Run("DecodeError", testWRPHandlerDecodeError)
		t.Run("ResponseWriterError", testWRPHandlerResponseWriterError)
		t.Run("Success", testWRPHandlerSuccess)
		t.Run("TransactionUUIDError", testTransactionUUIDError)
	})
}
