// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// allFormats enumerates all of the supported formats to use in testing
	allFormats = []Format{JSON, Msgpack}
)

func TestFindStringSubMatch(t *testing.T) {
	events := []string{
		"event:iot",
		"mac:112233445566/event/iot",
		"event:unknown",
		"mac:112233445566/event/test/extra-stuff",
		"event:wrp",
	}

	expected := []string{
		"iot",
		"unknown",
		"unknown",
		"unknown",
		"wrp",
	}

	var result string
	for i := 0; i < len(events); i++ {
		result = findEventStringSubMatch(events[i])
		if result != expected[i] {
			t.Errorf("\ntesting %v:\ninput: %v\nexpected: %v\ngot: %v\n\n", i, spew.Sprintf(events[i]), spew.Sprintf(expected[i]), spew.Sprintf(result))
		}
	}
}

func testMessageSetStatus(t *testing.T) {
	var (
		assert  = assert.New(t)
		message Message
	)

	assert.Nil(message.Status)
	assert.True(&message == message.SetStatus(72))
	assert.NotNil(message.Status)
	assert.Equal(int64(72), *message.Status)
	assert.True(&message == message.SetStatus(6))
	assert.NotNil(message.Status)
	assert.Equal(int64(6), *message.Status)
}

func testMessageSetRequestDeliveryResponse(t *testing.T) {
	var (
		assert  = assert.New(t)
		message Message
	)

	assert.Nil(message.RequestDeliveryResponse)
	assert.True(&message == message.SetRequestDeliveryResponse(14))
	assert.NotNil(message.RequestDeliveryResponse)
	assert.Equal(int64(14), *message.RequestDeliveryResponse)
	assert.True(&message == message.SetRequestDeliveryResponse(456))
	assert.NotNil(message.RequestDeliveryResponse)
	assert.Equal(int64(456), *message.RequestDeliveryResponse)
}

func testMessageEncode(t *testing.T, f Format, original Message) {
	var (
		assert  = assert.New(t)
		decoded Message

		buffer  bytes.Buffer
		encoder = NewEncoder(&buffer, f)
		decoder = NewDecoder(&buffer, f)
	)

	assert.NoError(encoder.Encode(&original))
	assert.True(buffer.Len() > 0)
	assert.NoError(decoder.Decode(&decoded))
	assert.Equal(original, decoded)
}

func TestMessage(t *testing.T) {
	t.Run("SetStatus", testMessageSetStatus)
	t.Run("SetRequestDeliveryResponse", testMessageSetRequestDeliveryResponse)

	var (
		expectedStatus                  int64 = 3471
		expectedRequestDeliveryResponse int64 = 34

		messages = []Message{
			{},
			{
				Type:             SimpleEventMessageType,
				Source:           "mac:121234345656",
				Destination:      "foobar.com/service",
				TransactionUUID:  "a unique identifier",
				QualityOfService: 24,
			},
			{
				Type:                    SimpleRequestResponseMessageType,
				Source:                  "somewhere.comcast.net:9090/something",
				Destination:             "serial:1234/blergh",
				TransactionUUID:         "123-123-123",
				Status:                  &expectedStatus,
				RequestDeliveryResponse: &expectedRequestDeliveryResponse,
			},
			{
				Type:            SimpleRequestResponseMessageType,
				Source:          "external.com",
				Destination:     "mac:FFEEAADD44443333",
				TransactionUUID: "DEADBEEF",
				Headers:         []string{"Header1", "Header2"},
				Metadata:        map[string]string{"name": "value"},
				Payload:         []byte{1, 2, 3, 4, 0xff, 0xce},
				PartnerIDs:      []string{"foo"},
			},
			{
				Type:        CreateMessageType,
				Source:      "wherever.webpa.comcast.net/glorious",
				Destination: "uuid:1111-11-111111-11111",
				Path:        "/some/where/over/the/rainbow",
				Payload:     []byte{1, 2, 3, 4, 0xff, 0xce},
				PartnerIDs:  []string{"foo", "bar"},
			},
		}
	)

	for _, source := range allFormats {
		t.Run(fmt.Sprintf("Encode%s", source), func(t *testing.T) {
			for _, message := range messages {
				testMessageEncode(t, source, message)
			}
		})
	}
}

func TestIsQOSAckPart(t *testing.T) {
	tests := []struct {
		description string
		msg         Message
		ack         bool
	}{
		// Ack case
		{
			description: "SimpleEventMessageType QOSMediumValue ack",
			msg:         Message{Type: SimpleEventMessageType, QualityOfService: QOSMediumValue},
			ack:         true,
		},
		{
			description: "SimpleEventMessageType QOSHighValue ack",
			msg:         Message{Type: SimpleEventMessageType, QualityOfService: QOSHighValue},
			ack:         true,
		},
		{
			description: "SimpleEventMessageType QOSCriticalValue ack",
			msg:         Message{Type: SimpleEventMessageType, QualityOfService: QOSCriticalValue},
			ack:         true,
		},
		{
			description: "SimpleEventMessageType above QOS range ack",
			msg:         Message{Type: SimpleEventMessageType, QualityOfService: QOSCriticalValue + 1},
			ack:         true,
		},
		{
			description: "SimpleRequestResponseMessageType ack",
			msg:         Message{Type: SimpleRequestResponseMessageType, QualityOfService: QOSCriticalValue},
			ack:         true,
		},
		{
			description: "CreateMessageType ack",
			msg:         Message{Type: CreateMessageType, QualityOfService: QOSCriticalValue},
			ack:         true,
		},
		{
			description: "RetrieveMessageType ack",
			msg:         Message{Type: RetrieveMessageType, QualityOfService: QOSCriticalValue},
			ack:         true,
		},
		{
			description: "UpdateMessageType ack",
			msg:         Message{Type: UpdateMessageType, QualityOfService: QOSCriticalValue},
			ack:         true,
		},
		{
			description: "DeleteMessageType ack",
			msg:         Message{Type: DeleteMessageType, QualityOfService: QOSCriticalValue},
			ack:         true,
		},

		// No ack case
		{
			description: "SimpleEventMessageType below QOS range no ack",
			msg:         Message{Type: SimpleEventMessageType, QualityOfService: QOSLowValue - 1},
		},
		{
			description: "SimpleEventMessageType QOSLowValue no ack",
			msg:         Message{Type: SimpleEventMessageType, QualityOfService: QOSLowValue},
		},
		{
			description: "ServiceRegistrationMessageType no ack",
			msg:         Message{Type: ServiceRegistrationMessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "Invalid0MessageType no ack",
			msg:         Message{Type: Invalid0MessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "ServiceAliveMessageType no ack",
			msg:         Message{Type: ServiceAliveMessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "UnknownMessageType no ack",
			msg:         Message{Type: UnknownMessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "AuthorizationMessageType no ack",
			msg:         Message{Type: AuthorizationMessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "Invalid0MessageType no ack",
			msg:         Message{Type: Invalid0MessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "Invalid1MessageType no ack",
			msg:         Message{Type: Invalid1MessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "lastMessageType no ack",
			msg:         Message{Type: LastMessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "Nonexistent negative MessageType no ack",
			msg:         Message{Type: -10, QualityOfService: QOSCriticalValue},
		},
		{
			description: "Nonexistent positive MessageType no ack",
			msg:         Message{Type: LastMessageType + 1, QualityOfService: QOSCriticalValue},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			if tc.ack {
				assert.True(tc.msg.IsQOSAckPart())
				return
			}

			assert.False(tc.msg.IsQOSAckPart())
		})
	}
}

func TestMessage_TrimmedPartnerIDs(t *testing.T) {
	tests := []struct {
		description string
		partners    []string
		want        []string
	}{
		{
			description: "empty partner list",
			partners:    []string{},
			want:        []string{},
		}, {
			description: "normal partner list",
			partners:    []string{"foo", "bar", "baz"},
			want:        []string{"foo", "bar", "baz"},
		}, {
			description: "partner list with empty strings",
			partners:    []string{"", "foo", "", "bar", "", "baz", ""},
			want:        []string{"foo", "bar", "baz"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			msg := &Message{
				PartnerIDs: tc.partners,
			}
			assert.Equal(tc.want, msg.TrimmedPartnerIDs())
		})
	}
}

func mapToEnviron(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

func TestEnviron_Message(t *testing.T) {
	tests := []struct {
		description string
		want        Message
		err         error
	}{
		{
			description: "empty",
		}, {
			description: "simple",
			want: Message{
				Source:           "source",
				Destination:      "destination",
				TransactionUUID:  "transaction_uuid",
				QualityOfService: 24,
				PartnerIDs:       []string{"foo", "bar", "baz"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			msg := tc.want
			m := msg.ToEnvironForm()
			assert.NotNil(t, m)

			got, err := NewMessageFromEnviron(mapToEnviron(m))

			if tc.err != nil {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, &msg, got)
		})
	}
}

func TestHeaders_Message(t *testing.T) {
	tests := []struct {
		description string
		want        Message
		err         error
	}{
		{
			description: "simple",
			want: Message{
				Type:             SimpleEventMessageType,
				Source:           "source",
				Destination:      "destination",
				TransactionUUID:  "transaction_uuid",
				QualityOfService: 24,
				PartnerIDs:       []string{"foo", "bar", "baz"},
			},
		}, {
			description: "simple with payload",
			want: Message{
				Type:             SimpleEventMessageType,
				Source:           "source",
				Destination:      "destination",
				TransactionUUID:  "transaction_uuid",
				QualityOfService: 24,
				PartnerIDs:       []string{"foo", "bar", "baz"},
				Payload:          []byte("payload"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			msg := tc.want
			headers, payload := msg.ToHeaderForm()
			assert.NotNil(t, headers)
			if tc.want.Payload != nil {
				assert.NotNil(t, payload)
				assert.Equal(t, tc.want.Payload, payload)
			}

			got, err := NewMessageFromHeaders(headers, payload)

			if tc.err != nil {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)

			// The content type is set to application/octet-stream if the payload
			// is not empty and the content type is not set.
			if got.ContentType != "" && tc.want.ContentType == "" {
				assert.Equal(t, "application/octet-stream", got.ContentType)
				got.ContentType = ""
			}
			assert.Equal(t, &msg, got)
		})
	}
}
