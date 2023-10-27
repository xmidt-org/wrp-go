// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrphttp

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/xmidt-org/wrp-go/v3"
	"github.com/xmidt-org/wrp-go/v3/wrpcontext"
)

var defaultDecoder Decoder = DecodeEntity(wrp.Msgpack)

// Decoder turns an HTTP request into a WRP entity.
type Decoder func(context.Context, *http.Request) (*Entity, error)

// MessageFunc is a strategy for post-processing a WRP message, adding things to the
// context or performing other processing on the message itself.
type MessageFunc func(context.Context, *wrp.Message) context.Context

func DefaultDecoder() Decoder {
	return defaultDecoder
}

func DecodeEntity(defaultFormat wrp.Format) Decoder {
	return func(ctx context.Context, original *http.Request) (*Entity, error) {
		format, err := DetermineFormat(defaultFormat, original.Header, "Content-Type")
		if err != nil {
			return nil, fmt.Errorf("failed to determine format of Content-Type header: %v", err)
		}

		_, err = DetermineFormat(defaultFormat, original.Header, "Accept")
		if err != nil {
			return nil, fmt.Errorf("failed to determine format of Accept header: %v", err)
		}

		entity := &Entity{Format: format}
		if contents, ok := wrpcontext.GetContents(original.Context()); ok {
			entity.Bytes = contents
		} else {
			contents, err := io.ReadAll(original.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read request body: %v", err)
			}

			entity.Bytes = contents
		}

		if msg, ok := wrpcontext.GetMessage(original.Context()); ok {
			entity.Message = *msg
		} else {
			err = wrp.NewDecoderBytes(entity.Bytes, format).Decode(&entity.Message)
			if err != nil {
				return nil, fmt.Errorf("failed to decode wrp: %v", err)
			}
		}

		return entity, err
	}
}

func DecodeEntityFromSources(defaultFormat wrp.Format, allowHeaderSource bool) Decoder {
	return func(ctx context.Context, original *http.Request) (*Entity, error) {
		if allowHeaderSource && (original.Header.Get(MessageTypeHeader) != "" || original.Header.Get(msgTypeHeader) != "") {
			return DecodeRequestHeaders(ctx, original)
		}
		return DecodeEntity(defaultFormat)(ctx, original)
	}
}

// DecodeRequestHeaders is a Decoder that uses the HTTP headers as fields of a WRP message.
// The HTTP entity, if specified, is used as the payload of the WRP message.
func DecodeRequestHeaders(ctx context.Context, original *http.Request) (*Entity, error) {
	entity := &Entity{
		Format: wrp.Msgpack,
	}

	err := SetMessageFromHeaders(original.Header, &entity.Message)
	if err != nil {
		return nil, err
	}

	_, err = ReadPayload(original.Header, original.Body, &entity.Message)

	if err != nil {
		return entity, err
	}

	err = wrp.NewEncoderBytes(&entity.Bytes, entity.Format).Encode(entity.Message)
	return entity, err
}

// DecodeRequest is a Decoder that provides lower-level way of decoding an *http.Request
// Can work for servers that don't use a wrp.Handler
func DecodeRequest(r *http.Request, msg any) (*http.Request, error) {

	if _, ok := wrpcontext.GetMessage(r.Context()); ok {
		// Context already contains a message, so just return the original request
		return r, nil
	}

	format, err := DetermineFormat(wrp.JSON, r.Header, "Content-Type")
	if err != nil {
		return nil, fmt.Errorf("failed to determine format of Content-Type header: %v", err)
	}

	var decodedMessage wrp.Message

	contents, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}

	// Try to decode the message using the HTTP Request headers
	// If this doesn't work, decode the message as Msgpack or JSON format
	if err = SetMessageFromHeaders(r.Header, &decodedMessage); err != nil {
		// Msgpack or JSON Format
		err = wrp.NewDecoderBytes(contents, format).Decode(&decodedMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to decode wrp message: %v", err)
		}
	}

	ctx := wrpcontext.SetMessage(r.Context(), &decodedMessage)
	ctx = wrpcontext.SetContents(ctx, contents)

	// Return a new request with the new context, containing the decoded message
	return r.WithContext(ctx), nil
}
