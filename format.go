// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"bytes"
	"encoding/json"
	"io"
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

// AllFormats returns a distinct slice of all supported formats.
func AllFormats() []Format {
	return []Format{Msgpack, JSON}
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
		got, err := msg.marshalMsg(nil)
		if err != nil {
			return err
		}
		_, err = e.stream.Write(got)
		return err
	}

	_, err := msg.marshalMsg(*e.bits)
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
	err := d.dec.Decode(msg)
	if err != nil {
		return err
	}
	return msg.validate()
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
	_, err = msg.unmarshalMsg(d.bits)
	if err != nil {
		return err
	}

	return msg.validate()
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
