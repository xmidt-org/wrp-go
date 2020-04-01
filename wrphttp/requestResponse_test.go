package wrphttp

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/wrp-go/v3"
)

func testRequestContextDefault(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(context.Background(), new(Request).Context())
}

func testRequestContextCustom(t *testing.T) {
	var (
		assert = assert.New(t)
		ctx    = context.WithValue(context.Background(), "asdf", "poiuy")
		r      = Request{ctx: ctx}
	)

	assert.Equal(ctx, r.Context())
}

func testRequestWithContextNil(t *testing.T) {
	assert := assert.New(t)
	assert.Panics(func() {
		new(Request).WithContext(nil)
	})
}

func testRequestWithContextCustom(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		ctx = context.WithValue(context.Background(), "homer", "simpson")
		r   = &Request{
			Original: httptest.NewRequest("GET", "/", nil),
			Entity:   new(Entity),
		}

		c = r.WithContext(ctx)
	)

	require.NotNil(c)
	assert.False(r == c)
	assert.True(r.Entity == c.Entity)
	assert.Equal(r.Original, c.Original)
	assert.Equal(ctx, c.Context())
}

func TestRequest(t *testing.T) {
	t.Run("Context", func(t *testing.T) {
		t.Run("Default", testRequestContextDefault)
		t.Run("Custom", testRequestContextCustom)
	})

	t.Run("WithContext", func(t *testing.T) {
		t.Run("Nil", testRequestWithContextNil)
		t.Run("Custom", testRequestWithContextCustom)
	})
}

func testEntityResponseWriterInvalidAccept(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		erw          = NewEntityResponseWriter(wrp.Msgpack)
		httpResponse = httptest.NewRecorder()
		wrpRequest   = &Request{
			Original: httptest.NewRequest("POST", "/", nil),
		}
	)

	require.NotNil(erw)
	wrpRequest.Original.Header.Set("Accept", "asd;lfkjasdfkjasdfkjasdf")

	wrpResponse, err := erw(httpResponse, wrpRequest)
	assert.Nil(wrpResponse)
	assert.Error(err)
}

func testWriteWRPSuccess(t *testing.T, includeBytes bool, defaultFormat, expectedFormat wrp.Format, accept string) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		erw          = NewEntityResponseWriter(defaultFormat)
		httpResponse = httptest.NewRecorder()
		wrpRequest   = &Request{
			Original: httptest.NewRequest("POST", "/", nil),
		}

		expected = &wrp.Message{
			Type:        wrp.SimpleRequestResponseMessageType,
			ContentType: "text/plain",
			Payload:     []byte("hi there"),
		}
	)

	require.NotNil(erw)
	wrpRequest.Original.Header.Set("Accept", accept)

	wrpResponse, err := erw(httpResponse, wrpRequest)
	require.NoError(err)
	require.NotNil(wrpResponse)

	e := &Entity{
		Message: *expected,
		Format:  expectedFormat,
	}

	if includeBytes {
		e.Bytes = wrp.MustEncode(e.Message, expectedFormat)
	}

	count, err := wrpResponse.WriteWRP(e)
	require.NoError(err)
	assert.True(count > 0)

	actual := new(wrp.Message)
	assert.NoError(wrp.NewDecoder(httpResponse.Body, expectedFormat).Decode(actual))
	assert.Equal(*expected, *actual)
}

func testWriteWRPSuccessDespiteWrongBytesFormat(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		erw          = NewEntityResponseWriter(wrp.JSON)
		httpResponse = httptest.NewRecorder()
		wrpRequest   = &Request{
			Original: httptest.NewRequest("POST", "/", nil),
		}

		expected = &wrp.Message{
			Type:        wrp.SimpleRequestResponseMessageType,
			ContentType: "text/plain",
			Payload:     []byte("hi there"),
		}
	)

	require.NotNil(erw)
	wrpRequest.Original.Header.Set("Accept", wrp.Msgpack.ContentType())

	wrpResponse, err := erw(httpResponse, wrpRequest)
	require.NoError(err)
	require.NotNil(wrpResponse)

	e := &Entity{
		Message: *expected,
		Format:  wrp.JSON,
		Bytes:   wrp.MustEncode(*expected, wrp.JSON),
	}

	count, err := wrpResponse.WriteWRP(e)
	require.NoError(err)
	assert.True(count > 0)

	actual := new(wrp.Message)
	assert.NoError(wrp.NewDecoder(httpResponse.Body, wrp.Msgpack).Decode(actual))
	assert.Equal(*expected, *actual)
}

func testWriteWRPFail(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		erw          = NewEntityResponseWriter(wrp.Msgpack)
		httpResponse = httptest.NewRecorder()
		wrpRequest   = &Request{
			Original: httptest.NewRequest("POST", "/", nil),
		}
	)

	wrpResponse, err := erw(httpResponse, wrpRequest)
	require.Nil(err)
	assert.Panics(func() {
		wrpResponse.WriteWRP(nil)
	})
}

func testWriteWRPBytesFail(t *testing.T, bytes []byte, f wrp.Format, expectedErr error) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		erw          = NewEntityResponseWriter(wrp.Msgpack)
		httpResponse = httptest.NewRecorder()
		wrpRequest   = &Request{
			Original: httptest.NewRequest("POST", "/", nil),
		}
	)
	wrpRequest.Original.Header.Set("Accept", wrp.Msgpack.ContentType())

	wrpResponse, err := erw(httpResponse, wrpRequest)
	require.Nil(err)
	i, e := wrpResponse.WriteWRPBytes(f, bytes)
	assert.Zero(i)
	assert.EqualValues(expectedErr, e)
}

func testWriteWRPFromBytesSuccess(t *testing.T, defaultFormat, expectedFormat wrp.Format, accept string) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		erw          = NewEntityResponseWriter(defaultFormat)
		httpResponse = httptest.NewRecorder()
		wrpRequest   = &Request{
			Original: httptest.NewRequest("POST", "/", nil),
		}

		expected = &wrp.Message{
			Type:        wrp.SimpleRequestResponseMessageType,
			ContentType: "text/plain",
			Payload:     []byte("hi there"),
		}
	)

	require.NotNil(erw)
	wrpRequest.Original.Header.Set("Accept", accept)

	wrpResponse, err := erw(httpResponse, wrpRequest)
	require.NoError(err)
	require.NotNil(wrpResponse)

	format := wrpResponse.WRPFormat()
	require.Equal(expectedFormat, format)

	count, err := wrpResponse.WriteWRPBytes(format, wrp.MustEncode(*expected, format))
	require.NoError(err)
	assert.True(count > 0)

	actual := new(wrp.Message)
	assert.NoError(wrp.NewDecoder(httpResponse.Body, expectedFormat).Decode(actual))
	assert.Equal(*expected, *actual)
}

func TestEntityResponseWriter(t *testing.T) {
	t.Run("InvalidAccept", testEntityResponseWriterInvalidAccept)

	t.Run("WriteWRPFail", testWriteWRPFail)

	t.Run("WriteWRPFromBytesFail", func(t *testing.T) {
		t.Run("EmptyBytes", func(t *testing.T) {
			testWriteWRPBytesFail(t, nil, wrp.Msgpack, ErrEmptyWRPBytes)
		})
		t.Run("ContentNegotation", func(t *testing.T) {
			testWriteWRPBytesFail(t, []byte("few-bytes"), wrp.JSON, ErrContentNegotiationMismatch)
		})
	})

	t.Run("WriteWRPSuccessDespiteWrongBytesFormat", testWriteWRPSuccessDespiteWrongBytesFormat)

	t.Run("WriteWRPSuccess", func(t *testing.T) {
		for _, defaultFormat := range wrp.AllFormats() {
			t.Run(defaultFormat.String(), func(t *testing.T) {
				t.Run("IncludeBytes", func(t *testing.T) {
					testWriteWRPSuccess(t, true, defaultFormat, defaultFormat, "")
				})
				t.Run("NoBytes", func(t *testing.T) {
					testWriteWRPSuccess(t, false, defaultFormat, defaultFormat, "")
				})
				for _, accept := range wrp.AllFormats() {
					t.Run(accept.String(), func(t *testing.T) {
						t.Run("IncludeBytes", func(t *testing.T) {
							testWriteWRPSuccess(t, true, defaultFormat, accept, accept.ContentType())
						})
						t.Run("NoBytes", func(t *testing.T) {
							testWriteWRPSuccess(t, false, defaultFormat, accept, accept.ContentType())
						})
					})
				}
			})
		}
	})

	t.Run("WriteWRPFromBytesSuccess", func(t *testing.T) {
		for _, defaultFormat := range wrp.AllFormats() {
			t.Run(defaultFormat.String(), func(t *testing.T) {
				testWriteWRPFromBytesSuccess(t, defaultFormat, defaultFormat, "")
				for _, accept := range wrp.AllFormats() {
					t.Run(accept.String(), func(t *testing.T) {
						testWriteWRPFromBytesSuccess(t, defaultFormat, accept, accept.ContentType())
					})
				}
			})
		}
	})
}
