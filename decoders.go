// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"bytes"
	"encoding/json"
	"io"
)

// Decoder is the interface for decoding WRP messages.
type Decoder interface {
	// Decode reads the next WRP message from the input stream and stores it in the
	// provided Union.  The message is validated before being stored.  To skip
	// validation, pass NoStandardValidation().  Custom validators can be provided
	// as additional arguments.
	Decode(dest Union, validators ...Processor) error
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

	if d.stream != nil {
		var err error
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
