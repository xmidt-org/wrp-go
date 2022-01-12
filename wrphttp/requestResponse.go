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
	"errors"
	"net/http"

	"github.com/xmidt-org/wrp-go/v3"
)

// Entity represents a WRP message in transit. Useful for
// request/response WRP handler exchanges.
type Entity struct {
	// Message holds the WRP message.
	Message wrp.Message

	// Format is the format of the WRP message encoded into the Bytes
	// field.
	// (Optional)
	Format wrp.Format

	// Bytes is the Format-encoded version of the WRP Message.
	// It serves as an optimization strategy for WRP handlers
	// to avoid unnecessary re-encodings.
	// (Optional)
	Bytes []byte
}

var (
	ErrEmptyWRPBytes              = errors.New("Encoded WRP bytes were empty.")
	ErrContentNegotiationMismatch = errors.New("Given format violates content negotiation")
)

// DetermineFormat examines zero or more headers to determine which WRP format is to be used, either
// for decoding or encoding.  The headers are tried in order, and the first non-empty value that maps
// to a WRP format is returned.  Any non-empty header that is invalid results in an error.  If none of
// the headers are present, this function returns the defaultFormat.
//
// This function can be used with a single header, e.g. DetermineFormat(wrp.Msgpack, header, "Content-Type").
// It can also be used for simple content negotiation, e.g. DetermineFormat(wrp.Msgpack, header, "Accept", "Content-Type").
func DetermineFormat(defaultFormat wrp.Format, h http.Header, names ...string) (wrp.Format, error) {
	for _, n := range names {
		v := h.Get(n)
		if len(v) > 0 {
			return wrp.FormatFromContentType(v)
		}
	}

	return defaultFormat, nil
}

// Request wraps an original http.Request and contains WRP message information.  Context handling
// mimics http.Request.
type Request struct {
	// Original is the HTTP request which corresponds to this WRP request.  The request body will have
	// already been read to produce the entity.
	Original *http.Request

	// Entity is the decoded WRP message
	Entity *Entity

	// ctx is the constructed context, not necessarily the same as that in the original request
	ctx context.Context
}

// Context returns the context associated with this WRP Request, which is not necessarily the
// same as the context returned by r.Original.Context().  Use this method instead of the original
// request.
func (r *Request) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}

	return context.Background()
}

// WithContext returns a shallow copy of this WRP Request using the supplied context.
// The semantics of this method are the same as http.Request.WithContext.  Note that the
// original request's context is not updated via this method.
func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("the context cannot be nil")
	}

	copy := new(Request)
	*copy = *r
	copy.ctx = ctx
	return copy
}

// ResponseWriter extends http.ResponseWriter with some WRP behavior.
type ResponseWriter interface {
	http.ResponseWriter

	// WriteWRP writes a WRP message to the underlying response.  The format used is determined
	// by the configuration of the underlying implementation.  This method is idempotent, and returns
	// an error if called multiple times for the same instance.
	WriteWRP(e *Entity) (int, error)

	// WriteWRPBytes writes a WRP message to the underlying response. The byte array input is assumed
	// to be the WRP message in the given format. This method is idempotent, and behaves
	// similarly as WriteWRP.
	WriteWRPBytes(wrp.Format, []byte) (int, error)

	// WRPFormat returns the format in which the WRP message response will be written. It is a convenience
	// function to verify the right format is called with WriteWRPBytes
	WRPFormat() wrp.Format
}

type ResponseWriterFunc func(http.ResponseWriter, *Request) (ResponseWriter, error)

var defaultResponseWriterFunc ResponseWriterFunc = NewEntityResponseWriter(wrp.Msgpack)

func DefaultResponseWriterFunc() ResponseWriterFunc {
	return defaultResponseWriterFunc
}

// NewEntityResponseWriter creates a ResponseWriterFunc that returns an entity-based ResponseWriter.  The returned
// ResponseWriter writes WRP messages to the response body, using content negotiation with a fallback to the supplied
// default format.
func NewEntityResponseWriter(defaultFormat wrp.Format) ResponseWriterFunc {
	return func(httpResponse http.ResponseWriter, wrpRequest *Request) (ResponseWriter, error) {
		format, err := DetermineFormat(defaultFormat, wrpRequest.Original.Header, "Accept")
		if err != nil {
			return nil, err
		}

		return &entityResponseWriter{
			ResponseWriter: httpResponse,
			f:              format,
		}, nil
	}
}

// entityResponseWriter provides ResponseWriter behavior that marshals WRP messages into the HTTP entity (body)
type entityResponseWriter struct {
	http.ResponseWriter
	f wrp.Format
}

func (erw *entityResponseWriter) WriteWRP(e *Entity) (int, error) {
	var output []byte

	if len(e.Bytes) > 0 && e.Format == erw.f {
		output = e.Bytes
	} else {
		encoder := wrp.NewEncoderBytes(&output, erw.f)

		if err := encoder.Encode(e.Message); err != nil {
			return 0, err
		}
	}

	erw.ResponseWriter.Header().Set("Content-Type", erw.f.ContentType())
	return erw.ResponseWriter.Write(output)
}

func (erw *entityResponseWriter) WriteWRPBytes(f wrp.Format, encodedWRP []byte) (int, error) {
	if encodedWRP == nil {
		return 0, ErrEmptyWRPBytes
	}
	if f != erw.f {
		return 0, ErrContentNegotiationMismatch
	}
	erw.ResponseWriter.Header().Set("Content-Type", f.ContentType())
	return erw.ResponseWriter.Write(encodedWRP)
}

func (erw *entityResponseWriter) WRPFormat() wrp.Format {
	return erw.f
}
