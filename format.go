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

func (e *jsonEncoder) Encode(msg *Message) error {
	err := e.enc.Encode(msg)
	if err == nil && e.buffer != nil && e.output != nil {
		*e.output = e.buffer.Bytes()
	}
	return err
}

type msgpEncoder struct {
	bits   *[]byte
	stream io.Writer
}

func (e *msgpEncoder) Encode(msg *Message) error {
	if e.stream != nil {
		got, err := msg.marshalMsg(nil)
		if err == nil {
			_, err = e.stream.Write(got)
		}
		return err
	}

	var err error
	*e.bits, err = msg.marshalMsg(*e.bits)
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
	_, err = msg.unmarshalMsg(d.bits)
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

func Decode[T UnionTypes](r io.Reader, f Format) (*T, error) {
	return DecodeThenValidate[T](r, f, NoStandardValidation())
}

func DecodeBytes[T UnionTypes](buf []byte, f Format) (*T, error) {
	return Decode[T](bytes.NewReader(buf), f)
}

func DecodeThenValidateBytes[T UnionTypes](buf []byte, f Format, validators ...Processor) (*T, error) {
	return DecodeThenValidate[T](bytes.NewReader(buf), f, validators...)
}

func DecodeThenValidate[T UnionTypes](r io.Reader, f Format, validators ...Processor) (*T, error) {
	var msg Message
	if err := f.Decoder(r).Decode(&msg); err != nil {
		return nil, err
	}

	if err := Validate(&msg, validators...); err != nil {
		return nil, err
	}

	out := new(T)
	switch m := any(out).(type) {
	case *Message:
		*m = msg
		return out, nil
	case *Authorization, *SimpleRequestResponse, *SimpleEvent, *CRUD, *ServiceRegistration, *ServiceAlive, *Unknown:
		if err := m.(converter).From(&msg); err != nil {
			return nil, err
		}
		return out, nil
	}

	return nil, ErrNotHandled
}

func Encode[T UnionTypes](msg *T, w io.Writer, f Format) error {
	return EncodeAfterValidate(msg, w, f, NoStandardValidation())
}

func EncodeBytes[T UnionTypes](msg *T, f Format) ([]byte, error) {
	return EncodeAfterValidateBytes(msg, f, NoStandardValidation())
}

func EncodeAfterValidateBytes[T UnionTypes](msg *T, f Format, validators ...Processor) ([]byte, error) {
	var buf bytes.Buffer
	if err := EncodeAfterValidate(msg, &buf, f, validators...); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// EncodeBytes is a convenience function that encodes a given message into a
// byte slice.
func EncodeAfterValidate[T UnionTypes](msg *T, w io.Writer, f Format, validators ...Processor) error {
	if err := Validate(msg, validators...); err != nil {
		return err
	}

	var encoder = NewEncoder(w, f)

	var err error
	var base *Message
	switch m := any(msg).(type) {
	case *Message:
		base = m
	case *Authorization, *SimpleRequestResponse, *SimpleEvent, *CRUD, *ServiceRegistration, *ServiceAlive, *Unknown:
		err = m.(converter).To(base)
	}

	if err != nil {
		return err
	}

	if err := encoder.Encode(base); err != nil {
		return err
	}

	return nil
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
func Validate[T UnionTypes](msg *T, validators ...Processor) error {
	return validateTo(msg, nil, validators...)
}

func validateTo[T UnionTypes](msg *T, base *Message, validators ...Processor) error {
	defaults := []Processor{
		StdValidator(),
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

	switch m := any(msg).(type) {
	case *Message:
		base = m
	case *Authorization, *SimpleRequestResponse, *SimpleEvent, *CRUD, *ServiceRegistration, *ServiceAlive, *Unknown:
		m.(converter).to(base)
	}

	err := Processors(validators).ProcessWRP(context.Background(), *base)
	if err == nil || errors.Is(err, ErrNotHandled) {
		return nil
	}

	return err
}
