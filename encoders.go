// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
)

// Encoder is the interface for encoding WRP messages.
type Encoder interface {
	// Encode writes the WRP message to the output stream after validating it.
	// To skip validation, pass NoStandardValidation().  Custom validators can be
	// provided as additional arguments.
	Encode(src Union, validators ...Processor) error
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

	if msg == nil || reflect.ValueOf(msg).IsNil() {
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

	err = e.enc.Encode(m)
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

	if msg == nil || reflect.ValueOf(msg).IsNil() {
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

// MustEncode is a convenience function that attempts to encode a given message.  A panic
// is raised on any error.  This function is handy for package initialization.
func MustEncode(message *Message, f Format) []byte {
	var out bytes.Buffer

	if err := f.Encoder(&out).Encode(message); err != nil {
		panic(err)
	}

	return out.Bytes()
}
