// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

// Encoder returns an Encoder for the given format.
func (f *Format) Encoder(output io.Writer) Encoder {
	return NewEncoder(output, *f)
}

// EncoderBytes returns an Encoder for the given format.
func (f *Format) EncoderBytes(output *[]byte) Encoder {
	return NewEncoderBytes(output, *f)
}

// Decoder returns a Decoder for the given format.
func (f *Format) Decoder(input io.Reader) Decoder {
	return NewDecoder(input, *f)
}

// DecoderBytes returns a Decoder for the given format.
func (f *Format) DecoderBytes(input []byte) Decoder {
	return NewDecoderBytes(input, *f)
}

// Encoder is the interface for encoding WRP messages.
type Encoder interface {
	// Encode writes the WRP message to the output stream after validating it.
	// To skip validation, pass NoStandardValidation().  Custom validators can be
	// provided as additional arguments.
	Encode(src Union, validators ...Processor) error
}

// Decoder is the interface for decoding WRP messages.
type Decoder interface {
	// Decode reads the next WRP message from the input stream and stores it in the
	// provided Union.  The message is validated before being stored.  To skip
	// validation, pass NoStandardValidation().  Custom validators can be provided
	// as additional arguments.
	Decode(dest Union, validators ...Processor) error
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
		buffer := bytes.NewBuffer(*output)
		return &jsonEncoder{
			buffer: buffer,
			enc:    json.NewEncoder(buffer),
			output: output,
		}
	case Msgpack:
		return &msgpEncoder{bits: output}
	}

	return nil
}

type jsonEncoder struct {
	enc    *json.Encoder
	buffer *bytes.Buffer
	output *[]byte
}

func (e *jsonEncoder) Encode(msg Union, validators ...Processor) error {
	var m *Message
	var err error

	if msg == nil {
		return ErrMessageIsInvalid
	}

	if tmp, ok := msg.(*Message); ok {
		m = tmp
		err = validate(m, validators...)
	} else {
		m = new(Message)
		err = msg.To(m, validators...)
	}
	if err != nil {
		return err
	}

	err = e.enc.Encode(msg)
	if err == nil && e.buffer != nil && e.output != nil {
		*e.output = e.buffer.Bytes()
	}
	return err
}

type msgpEncoder struct {
	bits   *[]byte
	stream io.Writer
}

func (e *msgpEncoder) Encode(msg Union, validators ...Processor) error {
	var m *Message
	var err error

	if msg == nil {
		return ErrMessageIsInvalid
	}

	if tmp, ok := msg.(*Message); ok {
		m = tmp
		err = validate(m, validators...)
	} else {
		m = new(Message)
		err = msg.To(m, validators...)
	}
	if err != nil {
		return err
	}

	if e.stream != nil {
		got, err := m.marshalMsg(nil)
		if err == nil {
			_, err = e.stream.Write(got)
		}
		return err
	}

	*e.bits, err = m.marshalMsg(*e.bits)
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

func (d *jsonDecoder) Decode(msg Union, validators ...Processor) error {
	if msg == nil {
		return ErrMessageIsInvalid
	}

	var m Message
	err := d.dec.Decode(&m)
	if err != nil {
		return err
	}

	return msg.From(&m, validators...)
}

type msgpDecoder struct {
	bits   []byte
	stream io.Reader
}

func (d *msgpDecoder) Decode(msg Union, validators ...Processor) error {
	if msg == nil {
		return ErrMessageIsInvalid
	}

	var err error
	if d.stream != nil {
		d.bits, err = io.ReadAll(d.stream)
		if err != nil {
			return err
		}
	}
	var tmp Message
	if _, err := tmp.unmarshalMsg(d.bits); err != nil {
		return err
	}

	return msg.From(&tmp, validators...)
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

// Validate performs a set of validations on a message.  If no validators are
// provided, the default set of standard WRP validators is used.  If the
// NoStandardValidation() processor is provided, no standard validation is
// performed.  After standard validation (if applicable) is performed, any
// additional validators are executed in the order they are provided.  If any
// validator returns an error excluding ErrNotHandled, the iteration stops and
// the error is returned.  If a validator return ErrNotHandled, then the
// validation is considered successful.  Any combination of nil errors and
// ErrNotHandled is considered a successful validation.  All other errors are
// considered validation failures and the first encountered error is returned.
func Validate(msg Union, validators ...Processor) error {
	if msg == nil {
		return ErrMessageIsInvalid
	}

	if m, ok := msg.(*Message); ok {
		return validate(m, validators...)
	}

	return msg.To(&Message{}, validators...)
}

func validate(msg *Message, validators ...Processor) error {
	defaults := []Processor{
		StandardValidator(),
	}
	for _, v := range validators {
		if v == nil {
			continue
		}
		if _, ok := v.(noStandardValidation); ok {
			defaults = nil
			break
		}
	}

	validators = append(defaults, validators...)

	err := Processors(validators).ProcessWRP(context.Background(), *msg)
	if err == nil || errors.Is(err, ErrNotHandled) {
		return nil
	}

	return err
}
