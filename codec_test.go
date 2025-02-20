// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type codecTest struct {
	desc     string
	in       Union
	encVador Processor
	encErr   bool
	skipEnc  bool
	skipDec  bool
	final    Union
	dec      []byte
	decVador Processor
	decErr   bool
}

var codecTests = []codecTest{
	{
		desc:     "nil msg and nil target",
		in:       nil,
		encVador: nil,
		encErr:   true,
		decErr:   true,
	}, {
		desc:    "invalid decode buffer",
		final:   &Message{},
		skipEnc: true,
		dec:     []byte{0x01},
		decErr:  true,
	}, {
		desc: "simple event",
		in: &Message{
			Type:        SimpleEventMessageType,
			Source:      "dns:example.com",
			Destination: "mac:FFEEDDCCBBAA",
		},
		final: &SimpleEvent{
			Source:      "dns:example.com",
			Destination: "mac:FFEEDDCCBBAA",
		},
	}, {
		desc: "specific simple event",
		in: &SimpleEvent{
			Source:      "dns:example.com",
			Destination: "mac:FFEEDDCCBBAA",
		},
		final: &Message{
			Type:        SimpleEventMessageType,
			Source:      "dns:example.com",
			Destination: "mac:FFEEDDCCBBAA",
		},
	}, {
		desc: "specific simple event, missing fields",
		in: &SimpleEvent{
			Source: "dns:example.com",
		},
		encErr:  true,
		skipDec: true,
	}, {
		desc: "service alive invalid",
		in: &Message{
			Type: UnknownMessageType,
		},
		final:  &ServiceAlive{},
		decErr: true,
	},
}

func TestCodecs(t *testing.T) {
	for _, test := range codecTests {
		for _, f := range AllFormats() {
			t.Run(test.desc+" "+f.String(), func(t *testing.T) {
				assert := assert.New(t)
				require := require.New(t)

				buf := bytes.Buffer{}
				if !test.skipEnc {
					err := f.Encoder(&buf).Encode(test.in, test.encVador)
					if test.encErr {
						assert.Error(err)
					} else {
						require.NoError(err)

						// Ensure both Encoder and EncoderBytes produce the same output
						var bts []byte
						err = f.EncoderBytes(&bts).Encode(test.in, test.encVador)
						assert.NoError(err)
						if f == JSON {
							assert.JSONEq(buf.String(), string(bts))
						} else {
							assert.Equal(buf.Bytes(), bts)
						}
					}
				}

				if test.skipDec {
					return
				}

				var dst Union

				if test.final != nil {
					// invoke a copy of the input type
					dst = reflect.New(reflect.TypeOf(test.final).Elem()).Interface().(Union)
				}

				dec := test.dec
				if dec == nil {
					dec = buf.Bytes()
				}

				//fmt.Printf("dec: %s\n", dec)

				err := f.DecoderBytes(dec).Decode(dst, test.decVador)
				if test.decErr {
					assert.Error(err)
					return
				}

				require.NoError(err)
			})
		}
	}
}

func TestMustEncode(t *testing.T) {
	assert.NotNil(t, MustEncode(&Message{Type: UnknownMessageType}, JSON))
	assert.NotNil(t, MustEncode(&Message{Type: UnknownMessageType}, Msgpack))

	assert.Panics(t, func() {
		MustEncode(nil, JSON)
	})
}
