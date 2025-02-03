// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
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
)

const (
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
//	FormatFromContentType(header.Get("Content-Type"), wrp.Msgpack)
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

// Encoder represents the underlying ugorji behavior that WRP supports
type Encoder interface {
	Encode(*Message) error
}

// Decoder represents the underlying ugorji behavior that WRP supports
type Decoder interface {
	Decode(*Message) error
}

// NewEncoder produces an Encoder using the appropriate WRP configuration for
// the given format.
func NewEncoder(output io.Writer, f Format) Encoder {
	switch f {
	case JSON:
		return &jsonEncoder{enc: json.NewEncoder(output)}
	case Msgpack:
		return &msgpEncoder{stream: output}
	}

	return nil
}

// NewEncoderBytes produces an Encoder using the appropriate WRP configuration
// for the given format.
func NewEncoderBytes(output *[]byte, f Format) Encoder {
	switch f {
	case JSON:
		return &jsonEncoder{enc: json.NewEncoder(bytes.NewBuffer(*output))}
	case Msgpack:
		return &msgpEncoder{bits: output}
	}

	return nil
}

type jsonEncoder struct {
	enc *json.Encoder
}

func (e *jsonEncoder) Encode(msg *Message) error {
	return e.enc.Encode(msg)
}

type msgpEncoder struct {
	bits   *[]byte
	stream io.Writer
}

func (e *msgpEncoder) Encode(msg *Message) error {
	if e.stream != nil {
		got, err := msg.MarshalMsg(nil)
		if err != nil {
			return err
		}
		_, err = e.stream.Write(got)
		return err
	}

	_, err := msg.MarshalMsg(*e.bits)
	return err
}

// NewDecoder produces a ugorji Decoder using the appropriate WRP configuration
// for the given format
func NewDecoder(input io.Reader, f Format) Decoder {
	switch f {
	case JSON:
		d := json.NewDecoder(input)
		d.UseNumber()
		return &jsonDecoder{dec: d}
	case Msgpack:
		return &msgpDecoder{stream: input}
	}

	return nil
}

// NewDecoderBytes produces a ugorji Decoder using the appropriate WRP configuration
// for the given format
func NewDecoderBytes(input []byte, f Format) Decoder {
	switch f {
	case JSON:
		return &jsonDecoder{dec: json.NewDecoder(bytes.NewReader(input))}
	case Msgpack:
		return &msgpDecoder{bits: input}
	}

	return nil
}

type jsonDecoder struct {
	dec *json.Decoder
}

func (d *jsonDecoder) Decode(msg *Message) error {
	return d.dec.Decode(msg)
}

type msgpDecoder struct {
	bits   []byte
	stream io.Reader
}

func (d *msgpDecoder) Decode(msg *Message) error {
	var err error
	if d.stream != nil {
		d.bits, err = io.ReadAll(d.stream)
		if err != nil {
			return err
		}
	}
	_, err = msg.UnmarshalMsg(d.bits)

	return err
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
func MustEncode(message *Message, f Format) []byte {
	var (
		output  bytes.Buffer
		encoder = NewEncoder(&output, f)
	)

	if err := encoder.Encode(message); err != nil {
		panic(err)
	}

	return output.Bytes()
}
