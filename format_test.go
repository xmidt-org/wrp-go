// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	// "github.com/k0kubun/pp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// "zappem.net/pub/debug/xxd"
)

func testPayload(t *testing.T, payload []byte) {
	var (
		assert   = assert.New(t)
		require  = require.New(t)
		original = Message{
			Type:        SimpleEventMessageType,
			Source:      "dns:source",
			Destination: "dns:destination",
			Payload:     payload,
		}

		decoded Message

		output bytes.Buffer
	)

	encoder := NewEncoder(&output, Msgpack)
	require.NoError(encoder.Encode(&original))

	decoder := NewDecoder(&output, Msgpack)
	require.NoError(decoder.Decode(&decoded))

	// don't output the payload if it's a ridiculous size
	if testing.Verbose() && len(payload) < 1024 {
		fmt.Println(hex.Dump(output.Bytes()))
		t.Logf("original.Payload=%s", original.Payload)
		t.Logf("decoded.Payload=%s", decoded.Payload)
	}

	assert.Equal(payload, decoded.Payload)
}

func TestPayload(t *testing.T) {
	t.Run("UTF8", func(t *testing.T) {
		testPayload(t, []byte("this is clearly a UTF8 string"))
	})

	t.Run("Binary", func(t *testing.T) {
		testPayload(t, []byte{0x00, 0x06, 0xFF, 0xF0})
	})

	t.Run("LargePayload", func(t *testing.T) {
		// generate a very large random payload
		payload := make([]byte, 70*1024)
		rand.Read(payload)
		testPayload(t, payload)
	})
}

func TestSampleMsgpack(t *testing.T) {
	var (
		sampleEncoded = []byte{
			0x85, 0xa8, 0x6d, 0x73, 0x67, 0x5f, 0x74, 0x79,
			0x70, 0x65, 0x03, 0xb0, 0x74, 0x72, 0x61, 0x6e,
			0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
			0x75, 0x75, 0x69, 0x64, 0xd9, 0x24, 0x39, 0x34,
			0x34, 0x37, 0x32, 0x34, 0x31, 0x63, 0x2d, 0x35,
			0x32, 0x33, 0x38, 0x2d, 0x34, 0x63, 0x62, 0x39,
			0x2d, 0x39, 0x62, 0x61, 0x61, 0x2d, 0x37, 0x30,
			0x37, 0x36, 0x65, 0x33, 0x32, 0x33, 0x32, 0x38,
			0x39, 0x39, 0xa6, 0x73, 0x6f, 0x75, 0x72, 0x63,
			0x65, 0xd9, 0x26, 0x64, 0x6e, 0x73, 0x3a, 0x77,
			0x65, 0x62, 0x70, 0x61, 0x2e, 0x63, 0x6f, 0x6d,
			0x63, 0x61, 0x73, 0x74, 0x2e, 0x63, 0x6f, 0x6d,
			0x2f, 0x76, 0x32, 0x2d, 0x64, 0x65, 0x76, 0x69,
			0x63, 0x65, 0x2d, 0x63, 0x6f, 0x6e, 0x66, 0x69,
			0x67, 0xa4, 0x64, 0x65, 0x73, 0x74, 0xb2, 0x73,
			0x65, 0x72, 0x69, 0x61, 0x6c, 0x3a, 0x31, 0x32,
			0x33, 0x34, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
			0x67, 0xa7, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61,
			0x64, 0xc4, 0x45, 0x7b, 0x20, 0x22, 0x6e, 0x61,
			0x6d, 0x65, 0x73, 0x22, 0x3a, 0x20, 0x5b, 0x20,
			0x22, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e,
			0x58, 0x5f, 0x43, 0x49, 0x53, 0x43, 0x4f, 0x5f,
			0x43, 0x4f, 0x4d, 0x5f, 0x53, 0x65, 0x63, 0x75,
			0x72, 0x69, 0x74, 0x79, 0x2e, 0x46, 0x69, 0x72,
			0x65, 0x77, 0x61, 0x6c, 0x6c, 0x2e, 0x46, 0x69,
			0x72, 0x65, 0x77, 0x61, 0x6c, 0x6c, 0x4c, 0x65,
			0x76, 0x65, 0x6c, 0x22, 0x20, 0x5d, 0x20, 0x7d,
		}

		sampleMessage = Message{
			Type:            SimpleRequestResponseMessageType,
			Source:          "dns:webpa.comcast.com/v2-device-config",
			Destination:     "serial:1234/config",
			TransactionUUID: "9447241c-5238-4cb9-9baa-7076e3232899",
			Payload: []byte(
				`{ "names": [ "Device.X_CISCO_COM_Security.Firewall.FirewallLevel" ] }`,
			),
		}
	)

	t.Run("Encode", func(t *testing.T) {
		var (
			assert        = assert.New(t)
			buffer        bytes.Buffer
			encoder       = NewEncoder(&buffer, Msgpack)
			decoder       = NewDecoder(&buffer, Msgpack)
			actualMessage Message
		)

		assert.NoError(encoder.Encode(&sampleMessage))
		assert.NoError(decoder.Decode(&actualMessage))
		assert.Equal(sampleMessage, actualMessage)
	})

	t.Run("Decode", func(t *testing.T) {
		var (
			assert        = assert.New(t)
			decoder       = NewDecoder(bytes.NewBuffer(sampleEncoded), Msgpack)
			actualMessage Message
		)

		assert.NoError(decoder.Decode(&actualMessage))
		assert.Equal(sampleMessage, actualMessage)
	})

	t.Run("DecodeBytes", func(t *testing.T) {
		var (
			assert        = assert.New(t)
			decoder       = NewDecoderBytes(sampleEncoded, Msgpack)
			actualMessage Message
		)

		assert.NoError(decoder.Decode(&actualMessage))
		assert.Equal(sampleMessage, actualMessage)
	})
}

func testFormatString(t *testing.T) {
	assert := assert.New(t)

	assert.NotEmpty(JSON.String())
	assert.NotEmpty(Msgpack.String())
	assert.NotEmpty(Format(-1).String())
	assert.NotEqual(JSON.String(), Msgpack.String())
}

func TestFormat(t *testing.T) {
	t.Run("String", testFormatString)
}

// testTranscodeMessage expects a nonpointer reference to a WRP message struct as the original parameter
func testTranscodeMessage(t *testing.T, target, source Format, original Message) {
	assert := assert.New(t)
	require := require.New(t)

	var (
		sourceBuffer  bytes.Buffer
		sourceEncoder = NewEncoder(&sourceBuffer, source)
		sourceDecoder = NewDecoder(&sourceBuffer, source)

		targetBuffer  bytes.Buffer
		targetEncoder = NewEncoder(&targetBuffer, target)
		targetDecoder = NewDecoder(&targetBuffer, target)
	)

	// create the input first
	require.NoError(sourceEncoder.Encode(&original))

	// now we can attempt the transcode
	message, err := TranscodeMessage(targetEncoder, sourceDecoder)
	assert.NotNil(message)
	assert.NoError(err)

	var got Message
	assert.NoError(targetDecoder.Decode(&got))
	assert.Equal(original, got)
}

func TestTranscodeMessage(t *testing.T) {
	var (
		expectedStatus                  int64 = 123
		expectedRequestDeliveryResponse int64 = -1234

		messages = []Message{
			{
				Type:        SimpleEventMessageType,
				Source:      "dns:foobar.com",
				Destination: "mac:FFEEDDCCBBAA",
				Payload:     []byte("hi!"),
			}, {
				Type:                    SimpleRequestResponseMessageType,
				Source:                  "dns:foobar.com",
				Destination:             "mac:FFEEDDCCBBAA",
				TransactionUUID:         "60dfdf5b-98c5-4e91-95fd-1fa6cb114cf5",
				ContentType:             "application/msgpack",
				Accept:                  "application/msgpack",
				Status:                  &expectedStatus,
				RequestDeliveryResponse: &expectedRequestDeliveryResponse,
				Headers:                 []string{"X-Header-1", "X-Header-2"},
				Metadata:                map[string]string{"hi": "there"},
				Payload:                 []byte("hi!"),
			},
			{
				Type:        SimpleEventMessageType,
				Source:      "dns:foobar.com",
				Destination: "mac:FFEEDDCCBBAA",
				Payload:     []byte("hi!"),
			},
			{
				Type:                    SimpleRequestResponseMessageType,
				Source:                  "dns:foobar.com",
				Destination:             "mac:FFEEDDCCBBAA",
				TransactionUUID:         "60dfdf5b-98c5-4e91-95fd-1fa6cb114cf5",
				ContentType:             "text/plain",
				Accept:                  "text/plain",
				Status:                  &expectedStatus,
				RequestDeliveryResponse: &expectedRequestDeliveryResponse,
				Headers:                 []string{"X-Header-1", "X-Header-2"},
				Metadata:                map[string]string{"hi": "there"},
				Payload:                 []byte("hi!"),
			},
		}
	)

	for _, target := range allFormats {
		for _, source := range allFormats {
			t.Run(fmt.Sprintf("%sTo%s", source, target), func(t *testing.T) {
				for _, original := range messages {
					testTranscodeMessage(t, target, source, original)
				}
			})
		}
	}
}

func TestAllFormats(t *testing.T) {
	require.NotNil(t, AllFormats())
	assert.Contains(t, AllFormats(), JSON)
	assert.Contains(t, AllFormats(), Msgpack)
}

func TestCodecEndToEnd(t *testing.T) {

	for _, format := range append(AllFormats(), Format(-1)) {
		if format == Format(-1) {
			assert.Nil(t, NewEncoder(nil, format))
			assert.Nil(t, NewDecoder(nil, format))
			assert.Nil(t, NewEncoderBytes(nil, format))
			assert.Nil(t, NewDecoderBytes(nil, format))
			continue
		}

		t.Run(format.String(), func(t *testing.T) {
			buf := new(bytes.Buffer)
			var bts []byte
			encoder := NewEncoder(buf, format)
			require.NotNil(t, encoder)

			encoderBytes := NewEncoderBytes(&bts, format)
			require.NotNil(t, encoderBytes)

			original := Message{Type: UnknownMessageType}

			require.NoError(t, encoder.Encode(&original))
			require.NoError(t, encoderBytes.Encode(&original))

			assert.Equal(t, buf.Bytes(), bts)

			decoder := NewDecoder(buf, format)
			require.NotNil(t, decoder)

			decoderBytes := NewDecoderBytes(bts, format)
			require.NotNil(t, decoderBytes)

			var decoded Message
			require.NoError(t, decoder.Decode(&decoded))
			assert.Equal(t, original, decoded)

			var decodedBytes Message
			require.NoError(t, decoderBytes.Decode(&decodedBytes))
			assert.Equal(t, original, decodedBytes)
		})
	}
}

func TestJSONDecode(t *testing.T) {
	// JSON encoded message
	goodJSON := []byte(`{ "msg_type": 11 }`)
	invalidJSON := []byte(`{ "msg_type": 11, }`)

	decoder := NewDecoderBytes(goodJSON, JSON)
	require.NotNil(t, decoder)

	var decoded Message
	err := decoder.Decode(&decoded)
	require.NoError(t, err)
	assert.Equal(t, UnknownMessageType, decoded.Type)

	decoder = NewDecoderBytes(invalidJSON, JSON)
	require.NotNil(t, decoder)

	err = decoder.Decode(&decoded)
	require.Error(t, err)
}

func TestMsgPackDecode(t *testing.T) {
	// JSON encoded message
	goodMsgpack := []byte{
		0x81,                                               // 1 map
		0xa8, 'm', 's', 'g', '_', 't', 'y', 'p', 'e', 0x0b, // "msg_type": 11
	}

	decoder := NewDecoderBytes(goodMsgpack, Msgpack)
	require.NotNil(t, decoder)

	var decoded Message
	err := decoder.Decode(&decoded)
	require.NoError(t, err)
	assert.Equal(t, UnknownMessageType, decoded.Type)

	decoder = NewDecoderBytes(goodMsgpack[:len(goodMsgpack)-2], Msgpack)
	require.NotNil(t, decoder)

	err = decoder.Decode(&decoded)
	require.Error(t, err)
}

func TestMustEncode(t *testing.T) {
	assert.NotNil(t, MustEncode(&Message{Type: UnknownMessageType}, JSON))
	assert.NotNil(t, MustEncode(&Message{Type: UnknownMessageType}, Msgpack))
}
