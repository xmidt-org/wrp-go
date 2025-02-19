// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
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
func (f Format) Encoder(output io.Writer) Encoder {
	return NewEncoder(output, f)
}

// EncoderBytes returns an Encoder for the given format.
func (f Format) EncoderBytes(output *[]byte) Encoder {
	return NewEncoderBytes(output, f)
}

// Decoder returns a Decoder for the given format.
func (f Format) Decoder(input io.Reader) Decoder {
	return NewDecoder(input, f)
}

// DecoderBytes returns a Decoder for the given format.
func (f Format) DecoderBytes(input []byte) Decoder {
	return NewDecoderBytes(input, f)
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
