// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllFormats(t *testing.T) {
	all := AllFormats()
	require.NotNil(t, all)

	for _, f := range all {
		var wantedStr string
		switch f {
		case JSON:
			wantedStr = "JSON"
		case Msgpack:
			wantedStr = "Msgpack"
		default:
			t.Errorf("unexpected format: %v", f)
		}

		assert.Equal(t, wantedStr, f.String())

		buf := bytes.Buffer{}
		require.NotNil(t, f.Encoder(&buf))
		require.NotNil(t, f.EncoderBytes(&[]byte{}))
		require.NotNil(t, f.Decoder(&buf))
		require.NotNil(t, f.DecoderBytes([]byte{}))
	}

	{
		invalid := Format(-1)
		assert.NotEmpty(t, invalid.String())

		buf := bytes.Buffer{}
		require.Nil(t, invalid.Encoder(&buf))
		require.Nil(t, invalid.EncoderBytes(&[]byte{}))
		require.Nil(t, invalid.Decoder(&buf))
		require.Nil(t, invalid.DecoderBytes([]byte{}))
	}
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
					tmp := original
					require.NoError(sourceEncoder.Encode(&tmp))

					// now we can attempt the transcode
					message, err := TranscodeMessage(targetEncoder, sourceDecoder)
					assert.NotNil(message)
					assert.NoError(err)

					var got Message
					assert.NoError(targetDecoder.Decode(&got))
					assert.Equal(original, got)
				}
			})
		}
	}
}
