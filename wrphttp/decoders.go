// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrphttp

import (
	"context"
	"encoding/json"
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

		// Check if the context already contains a message
		// If so, return the original request's message as an entity
		msg, ok := wrpcontext.Get[*wrp.Message](original.Context())
		if ok {
			jsonBytes, err := json.Marshal(msg)
			if err != nil {
				return nil, err
			}
			entity := &Entity{
				Message: *msg,
				Format:  format,
				Bytes:   jsonBytes,
			}
			return entity, nil
		}

		contents, err := io.ReadAll(original.Body)
		if err != nil {
			return nil, err
		}

		entity := &Entity{
			Format: format,
			Bytes:  contents,
		}

		err = wrp.NewDecoderBytes(contents, format).Decode(&entity.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to decode wrp: %v", err)
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

	if _, ok := wrpcontext.Get[*wrp.Message](r.Context()); ok {
		// Context already contains a message, so just return the original request
		return r, nil
	}

	format, err := DetermineFormat(wrp.JSON, r.Header, "Content-Type")
	if err != nil {
		return nil, fmt.Errorf("failed to determine format of Content-Type header: %v", err)
	}

	var decodedMessage wrp.Message

	// Try to decode the message using the HTTP Request headers
	// If this doesn't work, decode the message as Msgpack or JSON format
	if err = SetMessageFromHeaders(r.Header, &decodedMessage); err != nil {
		// Msgpack or JSON Format
		bodyReader := r.Body
		err = wrp.NewDecoder(bodyReader, format).Decode(&decodedMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to decode wrp message: %v", err)
		}
	}

	ctx := wrpcontext.Set(r.Context(), &decodedMessage)

	// Return a new request with the new context, containing the decoded message
	return r.WithContext(ctx), nil
}
