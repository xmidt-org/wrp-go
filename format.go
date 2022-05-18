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

package wrp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ugorji/go/codec"
)

//go:generate go install golang.org/x/tools/cmd/stringer@latest
//go:generate stringer -type=Format

// Format indicates which format is desired.
// The zero value indicates Msgpack, which means by default other
// infrastructure can assume msgpack-formatted data.
type Format int

const (
	Msgpack Format = iota
	JSON
	lastFormat

	MimeTypeMsgpack     = "application/msgpack"
	MimeTypeJson        = "application/json"
	MimeTypeOctetStream = "application/octet-stream"

	// Deprecated: This constant should only be used for backwards compatibility
	// matching.  Use MimeTypeMsgpack instead.
	MimeTypeWrp = "application/wrp"
)

// AllFormats returns a distinct slice of all supported formats.
func AllFormats() []Format {
	return []Format{Msgpack, JSON}
}

var (
	jsonHandle = codec.JsonHandle{
		BasicHandle: codec.BasicHandle{
			TypeInfos: codec.NewTypeInfos([]string{"json"}),
		},
		IntegerAsString: 'L',
	}

	// msgpackHandle uses the configuration required for the updated msgpack spec.
	// this is what's required to ensure that the Payload field is encoded and decoded properly.
	// See: http://ugorji.net/blog/go-codec-primer#format-specific-runtime-configuration
	msgpackHandle = codec.MsgpackHandle{
		WriteExt: true,
		BasicHandle: codec.BasicHandle{
			TypeInfos: codec.NewTypeInfos([]string{"json"}),
		},
	}
)

// ContentType returns the MIME type associated with this format
func (f Format) ContentType() string {
	switch f {
	case Msgpack:
		return MimeTypeMsgpack
	case JSON:
		return MimeTypeJson
	default:
		return MimeTypeOctetStream
	}
}

// FormatFromContentType examines the Content-Type value and returns
// the appropriate Format.  This function returns an error if the given
// Content-Type did not map to a WRP format.
//
// The optional fallback is used if contentType is the empty string.  Only
// the first fallback value is used.  The rest are ignored.  This approach allows
// simple usages such as:
//
//   FormatFromContentType(header.Get("Content-Type"), wrp.Msgpack)
func FormatFromContentType(contentType string, fallback ...Format) (Format, error) {
	if len(contentType) == 0 {
		if len(fallback) > 0 {
			return fallback[0], nil
		}

		return Format(-1), errors.New("Missing content type")
	}

	if strings.Contains(contentType, "json") {
		return JSON, nil
	} else if strings.Contains(contentType, "msgpack") {
		return Msgpack, nil
	}

	return Format(-1), fmt.Errorf("invalid WRP content type: %s", contentType)
}

// handle looks up the appropriate codec.Handle for this format constant.
// This method panics if the format is not a valid value.
func (f Format) handle() codec.Handle {
	switch f {
	case Msgpack:
		return &msgpackHandle
	case JSON:
		return &jsonHandle
	}

	panic(fmt.Errorf("Invalid format constant: %d", f))
}

// EncodeListener can be implemented on any type passed to an Encoder in order
// to get notified when an encoding happens.  This interface is useful to set
// mandatory fields, such as message type.
type EncodeListener interface {
	BeforeEncode() error
}

// Encoder represents the underlying ugorji behavior that WRP supports
type Encoder interface {
	Encode(interface{}) error
	Reset(io.Writer)
	ResetBytes(*[]byte)
}

// encoderDecorator wraps a ugorji Encoder and implements the wrp.Encoder interface.
type encoderDecorator struct {
	*codec.Encoder
}

// Encode checks to see if value implements EncoderTo and if it does, uses the
// value.EncodeTo() method.  Otherwise, the value is passed as is to the decorated
// ugorji Encoder.
func (ed *encoderDecorator) Encode(value interface{}) error {
	if listener, ok := value.(EncodeListener); ok {
		if err := listener.BeforeEncode(); err != nil {
			return err
		}
	}

	return ed.Encoder.Encode(value)
}

// Decoder represents the underlying ugorji behavior that WRP supports
type Decoder interface {
	Decode(interface{}) error
	Reset(io.Reader)
	ResetBytes([]byte)
}

// NewEncoder produces a ugorji Encoder using the appropriate WRP configuration
// for the given format
func NewEncoder(output io.Writer, f Format) Encoder {
	return &encoderDecorator{
		codec.NewEncoder(output, f.handle()),
	}
}

// NewEncoderBytes produces a ugorji Encoder using the appropriate WRP configuration
// for the given format
func NewEncoderBytes(output *[]byte, f Format) Encoder {
	return &encoderDecorator{
		codec.NewEncoderBytes(output, f.handle()),
	}
}

// NewDecoder produces a ugorji Decoder using the appropriate WRP configuration
// for the given format
func NewDecoder(input io.Reader, f Format) Decoder {
	return codec.NewDecoder(input, f.handle())
}

// NewDecoderBytes produces a ugorji Decoder using the appropriate WRP configuration
// for the given format
func NewDecoderBytes(input []byte, f Format) Decoder {
	return codec.NewDecoderBytes(input, f.handle())
}

// TranscodeMessage converts a WRP message of any type from one format into another,
// e.g. from JSON into Msgpack.  The intermediate, generic Message used to hold decoded
// values is returned in addition to any error.  If a decode error occurs, this function
// will not perform the encoding step.
func TranscodeMessage(target Encoder, source Decoder) (msg *Message, err error) {
	msg = new(Message)
	if err = source.Decode(msg); err == nil {
		err = target.Encode(msg)
	}

	return
}

// MustEncode is a convenience function that attempts to encode a given message.  A panic
// is raised on any error.  This function is handy for package initialization.
func MustEncode(message interface{}, f Format) []byte {
	var (
		output  bytes.Buffer
		encoder = NewEncoder(&output, f)
	)

	if err := encoder.Encode(message); err != nil {
		panic(err)
	}

	return output.Bytes()
}
