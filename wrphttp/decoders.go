package wrphttp

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/xmidt-org/wrp-go/v3"
)

// Decoder turns an HTTP request into a WRP entity.
type Decoder func(context.Context, *http.Request) (*Entity, error)

var defaultDecoder Decoder = DecodeEntity(wrp.Msgpack)

func DefaultDecoder() Decoder {
	return defaultDecoder
}

func DecodeEntity(defaultFormat wrp.Format) Decoder {
	return func(ctx context.Context, original *http.Request) (*Entity, error) {
		format, err := DetermineFormat(defaultFormat, original.Header, "Content-Type")
		if err != nil {
			return nil, err
		}

		contents, err := ioutil.ReadAll(original.Body)
		if err != nil {
			return nil, err
		}

		entity := &Entity{
			Format: format,
			Bytes:  contents,
		}

		err = wrp.NewDecoderBytes(contents, format).Decode(&entity.Message)
		return entity, err
	}
}

func DecodeEntityFromSources(defaultFormat wrp.Format, allowHeaderSource bool) Decoder {
	return func(ctx context.Context, original *http.Request) (*Entity, error) {
		if allowHeaderSource && original.Header.Get(MessageTypeHeader) != "" {
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

// MessageFunc is a strategy for post-processing a WRP message, adding things to the
// context or performing other processing on the message itself.
type MessageFunc func(context.Context, *wrp.Message) context.Context
