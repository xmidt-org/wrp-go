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

func testMessageSetIncludeSpans(t *testing.T) {
	var (
		assert  = assert.New(t)
		message Message
	)

	assert.Nil(message.IncludeSpans)
	assert.True(&message == message.SetIncludeSpans(true))
	assert.NotNil(message.IncludeSpans)
	assert.Equal(true, *message.IncludeSpans)
	assert.True(&message == message.SetIncludeSpans(false))
	assert.NotNil(message.IncludeSpans)
	assert.Equal(false, *message.IncludeSpans)
}

func testMessageRoutable(t *testing.T, original Message) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
	)

	assert.Equal(original.Type, original.MessageType())
	assert.Equal(original.Destination, original.To())
	assert.Equal(original.Source, original.From())
	assert.Equal(original.TransactionUUID, original.TransactionKey())
	assert.Equal(
		original.Type.RequiresTransaction() && len(original.TransactionUUID) > 0,
		original.IsTransactionPart(),
	)

	routable := original.Response("testMessageRoutable", 1234)
	require.NotNil(routable)
	response, ok := routable.(*Message)
	require.NotNil(response)
	require.True(ok)

	assert.Equal(original.Type, response.Type)
	assert.Equal(original.Source, response.Destination)
	assert.Equal("testMessageRoutable", response.Source)
	require.NotNil(response.RequestDeliveryResponse)
	assert.Equal(int64(1234), *response.RequestDeliveryResponse)
	assert.Nil(response.Payload)
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
	t.Run("SetIncludeSpans", testMessageSetIncludeSpans)

	var (
		expectedStatus                  int64 = 3471
		expectedRequestDeliveryResponse int64 = 34
		expectedIncludeSpans            bool  = true

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
				IncludeSpans:            &expectedIncludeSpans,
			},
			{
				Type:            SimpleRequestResponseMessageType,
				Source:          "external.com",
				Destination:     "mac:FFEEAADD44443333",
				TransactionUUID: "DEADBEEF",
				Headers:         []string{"Header1", "Header2"},
				Metadata:        map[string]string{"name": "value"},
				Spans:           [][]string{{"1", "2"}, {"3"}},
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

	t.Run("Routable", func(t *testing.T) {
		for _, message := range messages {
			testMessageRoutable(t, message)
		}
	})

	for _, source := range allFormats {
		t.Run(fmt.Sprintf("Encode%s", source), func(t *testing.T) {
			for _, message := range messages {
				testMessageEncode(t, source, message)
			}
		})
	}
}

func testSimpleRequestResponseSetStatus(t *testing.T) {
	var (
		assert  = assert.New(t)
		message SimpleRequestResponse
	)

	assert.Nil(message.Status)
	assert.True(&message == message.SetStatus(15))
	assert.NotNil(message.Status)
	assert.Equal(int64(15), *message.Status)
	assert.True(&message == message.SetStatus(2312))
	assert.NotNil(message.Status)
	assert.Equal(int64(2312), *message.Status)
}

func testSimpleRequestResponseSetRequestDeliveryResponse(t *testing.T) {
	var (
		assert  = assert.New(t)
		message SimpleRequestResponse
	)

	assert.Nil(message.RequestDeliveryResponse)
	assert.True(&message == message.SetRequestDeliveryResponse(2))
	assert.NotNil(message.RequestDeliveryResponse)
	assert.Equal(int64(2), *message.RequestDeliveryResponse)
	assert.True(&message == message.SetRequestDeliveryResponse(67))
	assert.NotNil(message.RequestDeliveryResponse)
	assert.Equal(int64(67), *message.RequestDeliveryResponse)
}

func testSimpleRequestResponseSetIncludeSpans(t *testing.T) {
	var (
		assert  = assert.New(t)
		message SimpleRequestResponse
	)

	assert.Nil(message.IncludeSpans)
	assert.True(&message == message.SetIncludeSpans(true))
	assert.NotNil(message.IncludeSpans)
	assert.Equal(true, *message.IncludeSpans)
	assert.True(&message == message.SetIncludeSpans(false))
	assert.NotNil(message.IncludeSpans)
	assert.Equal(false, *message.IncludeSpans)
}

func testSimpleRequestResponseRoutable(t *testing.T, original SimpleRequestResponse) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
	)

	assert.Equal(original.Type, original.MessageType())
	assert.Equal(original.Destination, original.To())
	assert.Equal(original.Source, original.From())
	assert.Equal(original.TransactionUUID, original.TransactionKey())
	assert.Equal(
		len(original.TransactionUUID) > 0,
		original.IsTransactionPart(),
	)

	routable := original.Response("testSimpleRequestResponseRoutable", 34734)
	require.NotNil(routable)
	response, ok := routable.(*SimpleRequestResponse)
	require.NotNil(response)
	require.True(ok)

	assert.Equal(original.Type, response.Type)
	assert.Equal(original.Source, response.Destination)
	assert.Equal("testSimpleRequestResponseRoutable", response.Source)
	require.NotNil(response.RequestDeliveryResponse)
	assert.Equal(int64(34734), *response.RequestDeliveryResponse)
	assert.Nil(response.Payload)
}

func testSimpleRequestResponseEncode(t *testing.T, f Format, original SimpleRequestResponse) {
	var (
		assert  = assert.New(t)
		decoded SimpleRequestResponse

		buffer  bytes.Buffer
		encoder = NewEncoder(&buffer, f)
		decoder = NewDecoder(&buffer, f)
	)

	assert.NoError(encoder.Encode(&original))
	assert.True(buffer.Len() > 0)
	assert.Equal(SimpleRequestResponseMessageType, original.Type)
	assert.NoError(decoder.Decode(&decoded))
	assert.Equal(original, decoded)
}

func TestSimpleRequestResponse(t *testing.T) {
	t.Run("SetStatus", testSimpleRequestResponseSetStatus)
	t.Run("SetRequestDeliveryResponse", testSimpleRequestResponseSetRequestDeliveryResponse)
	t.Run("SetIncludeSpans", testSimpleRequestResponseSetIncludeSpans)

	var (
		expectedStatus                  int64 = 121
		expectedRequestDeliveryResponse int64 = 17
		expectedIncludeSpans            bool  = true

		messages = []SimpleRequestResponse{
			{},
			{
				Source:          "mac:121234345656",
				Destination:     "foobar.com/service",
				TransactionUUID: "a unique identifier",
			},
			{
				Source:                  "somewhere.comcast.net:9090/something",
				Destination:             "serial:1234/blergh",
				TransactionUUID:         "123-123-123",
				Status:                  &expectedStatus,
				RequestDeliveryResponse: &expectedRequestDeliveryResponse,
				IncludeSpans:            &expectedIncludeSpans,
			},
			{
				Source:          "external.com",
				Destination:     "mac:FFEEAADD44443333",
				TransactionUUID: "DEADBEEF",
				Headers:         []string{"Header1", "Header2"},
				Metadata:        map[string]string{"name": "value"},
				Spans:           [][]string{{"1", "2"}, {"3"}},
				Payload:         []byte{1, 2, 3, 4, 0xff, 0xce},
			},
		}
	)

	t.Run("Routable", func(t *testing.T) {
		for _, message := range messages {
			testSimpleRequestResponseRoutable(t, message)
		}
	})

	for _, format := range allFormats {
		t.Run(fmt.Sprintf("Encode%s", format), func(t *testing.T) {
			for _, message := range messages {
				testSimpleRequestResponseEncode(t, format, message)
			}
		})
	}
}

func testSimpleEventRoutable(t *testing.T, original SimpleEvent) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
	)

	assert.Equal(original.Type, original.MessageType())
	assert.Equal(original.Destination, original.To())
	assert.Equal(original.Source, original.From())
	assert.Empty(original.TransactionKey())
	assert.False(original.IsTransactionPart())

	routable := original.Response("testSimpleEventRoutable", 82)
	require.NotNil(routable)
	response, ok := routable.(*SimpleEvent)
	require.NotNil(response)
	require.True(ok)

	assert.Equal(original.Type, response.Type)
	assert.Equal(original.Source, response.Destination)
	assert.Equal("testSimpleEventRoutable", response.Source)
	assert.Nil(response.Payload)
}

func testSimpleEventEncode(t *testing.T, f Format, original SimpleEvent) {
	var (
		assert  = assert.New(t)
		decoded SimpleEvent

		buffer  bytes.Buffer
		encoder = NewEncoder(&buffer, f)
		decoder = NewDecoder(&buffer, f)
	)

	assert.NoError(encoder.Encode(&original))
	assert.True(buffer.Len() > 0)
	assert.Equal(SimpleEventMessageType, original.Type)
	assert.NoError(decoder.Decode(&decoded))
	assert.Equal(original, decoded)
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
			description: "Invalid0MessageType no ack",
			msg:         Message{Type: Invalid0MessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "SimpleRequestResponseMessageType no ack",
			msg:         Message{Type: SimpleRequestResponseMessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "CreateMessageType no ack",
			msg:         Message{Type: CreateMessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "RetrieveMessageType no ack",
			msg:         Message{Type: RetrieveMessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "UpdateMessageType no ack",
			msg:         Message{Type: UpdateMessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "DeleteMessageType no ack",
			msg:         Message{Type: DeleteMessageType, QualityOfService: QOSCriticalValue},
		},
		{
			description: "ServiceRegistrationMessageType no ack",
			msg:         Message{Type: ServiceRegistrationMessageType, QualityOfService: QOSCriticalValue},
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

func TestSimpleEvent(t *testing.T) {
	var messages = []SimpleEvent{
		{},
		{
			Source:      "simple.com/foo",
			Destination: "uuid:111111111111111",
			Payload:     []byte("this is a lovely payloed"),
		},
		{
			Source:      "mac:123123123123123123",
			Destination: "something.webpa.comcast.net:9090/here/is/a/path",
			ContentType: "text/plain",
			Headers:     []string{"header1"},
			Metadata:    map[string]string{"a": "b", "c": "d"},
			Payload:     []byte("check this out!"),
		},
	}

	t.Run("Routable", func(t *testing.T) {
		for _, message := range messages {
			testSimpleEventRoutable(t, message)
		}
	})

	for _, format := range allFormats {
		t.Run(fmt.Sprintf("Encode%s", format), func(t *testing.T) {
			for _, message := range messages {
				testSimpleEventEncode(t, format, message)
			}
		})
	}
}

func testCRUDSetStatus(t *testing.T) {
	var (
		assert  = assert.New(t)
		message CRUD
	)

	assert.Nil(message.Status)
	assert.True(&message == message.SetStatus(-72))
	assert.NotNil(message.Status)
	assert.Equal(int64(-72), *message.Status)
	assert.True(&message == message.SetStatus(172))
	assert.NotNil(message.Status)
	assert.Equal(int64(172), *message.Status)
}

func testCRUDSetRequestDeliveryResponse(t *testing.T) {
	var (
		assert  = assert.New(t)
		message CRUD
	)

	assert.Nil(message.RequestDeliveryResponse)
	assert.True(&message == message.SetRequestDeliveryResponse(123))
	assert.NotNil(message.RequestDeliveryResponse)
	assert.Equal(int64(123), *message.RequestDeliveryResponse)
	assert.True(&message == message.SetRequestDeliveryResponse(543543))
	assert.NotNil(message.RequestDeliveryResponse)
	assert.Equal(int64(543543), *message.RequestDeliveryResponse)
}

func testCRUDSetIncludeSpans(t *testing.T) {
	var (
		assert  = assert.New(t)
		message CRUD
	)

	assert.Nil(message.IncludeSpans)
	assert.True(&message == message.SetIncludeSpans(true))
	assert.NotNil(message.IncludeSpans)
	assert.Equal(true, *message.IncludeSpans)
	assert.True(&message == message.SetIncludeSpans(false))
	assert.NotNil(message.IncludeSpans)
	assert.Equal(false, *message.IncludeSpans)
}

func testCRUDRoutable(t *testing.T, original CRUD) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
	)

	assert.Equal(original.Type, original.MessageType())
	assert.Equal(original.Destination, original.To())
	assert.Equal(original.Source, original.From())
	assert.Equal(original.TransactionUUID, original.TransactionKey())
	assert.Equal(
		len(original.TransactionUUID) > 0,
		original.IsTransactionPart(),
	)

	routable := original.Response("testCRUDRoutable", 369)
	require.NotNil(routable)
	response, ok := routable.(*CRUD)
	require.NotNil(response)
	require.True(ok)

	assert.Equal(original.Type, response.Type)
	assert.Equal(original.Source, response.Destination)
	assert.Equal("testCRUDRoutable", response.Source)
	require.NotNil(response.RequestDeliveryResponse)
	assert.Equal(int64(369), *response.RequestDeliveryResponse)
}

func testCRUDEncode(t *testing.T, f Format, original CRUD) {
	var (
		assert  = assert.New(t)
		decoded CRUD

		buffer  bytes.Buffer
		encoder = NewEncoder(&buffer, f)
		decoder = NewDecoder(&buffer, f)
	)

	assert.NoError(encoder.Encode(&original))
	assert.True(buffer.Len() > 0)
	assert.NoError(decoder.Decode(&decoded))
	assert.Equal(original, decoded)
}

func TestCRUD(t *testing.T) {
	t.Run("SetStatus", testCRUDSetStatus)
	t.Run("SetRequestDeliveryResponse", testCRUDSetRequestDeliveryResponse)
	t.Run("SetIncludeSpans", testCRUDSetIncludeSpans)

	var (
		expectedStatus                  int64 = -273
		expectedRequestDeliveryResponse int64 = 7223
		expectedIncludeSpans            bool  = true

		messages = []CRUD{
			{},
			{
				Type:            DeleteMessageType,
				Source:          "mac:121234345656",
				Destination:     "foobar.com/service",
				TransactionUUID: "a unique identifier",
				Path:            "/a/b/c/d",
			},
			{
				Type:                    CreateMessageType,
				Source:                  "somewhere.comcast.net:9090/something",
				Destination:             "serial:1234/blergh",
				TransactionUUID:         "123-123-123",
				ContentType:             "text/plain",
				Status:                  &expectedStatus,
				RequestDeliveryResponse: &expectedRequestDeliveryResponse,
				IncludeSpans:            &expectedIncludeSpans,
				Path:                    "/somewhere/over/rainbow",
				Payload:                 []byte{1, 2, 3, 4, 0xff, 0xce},
			},
			{
				Type:            UpdateMessageType,
				Source:          "external.com",
				Destination:     "mac:FFEEAADD44443333",
				TransactionUUID: "DEADBEEF",
				Headers:         []string{"Header1", "Header2"},
				Metadata:        map[string]string{"name": "value"},
				Spans:           [][]string{{"1", "2"}, {"3"}},
			},
		}
	)

	t.Run("Routable", func(t *testing.T) {
		for _, message := range messages {
			testCRUDRoutable(t, message)
		}
	})

	for _, format := range allFormats {
		t.Run(fmt.Sprintf("Encode%s", format), func(t *testing.T) {
			for _, message := range messages {
				testCRUDEncode(t, format, message)
			}
		})
	}
}

func testServiceRegistrationEncode(t *testing.T, f Format, original ServiceRegistration) {
	var (
		assert  = assert.New(t)
		decoded ServiceRegistration

		buffer  bytes.Buffer
		encoder = NewEncoder(&buffer, f)
		decoder = NewDecoder(&buffer, f)
	)

	assert.NoError(encoder.Encode(&original))
	assert.True(buffer.Len() > 0)
	assert.Equal(ServiceRegistrationMessageType, original.Type)
	assert.NoError(decoder.Decode(&decoded))
	assert.Equal(original, decoded)
}

func TestServiceRegistration(t *testing.T) {
	var messages = []ServiceRegistration{
		{},
		{
			ServiceName: "systemd",
		},
		{
			ServiceName: "systemd",
			URL:         "local:/location/here",
		},
	}

	for _, format := range allFormats {
		t.Run(fmt.Sprintf("Encode%s", format), func(t *testing.T) {
			for _, message := range messages {
				testServiceRegistrationEncode(t, format, message)
			}
		})
	}
}

func testServiceAliveEncode(t *testing.T, f Format) {
	var (
		assert   = assert.New(t)
		original = ServiceAlive{}

		decoded ServiceAlive

		buffer  bytes.Buffer
		encoder = NewEncoder(&buffer, f)
		decoder = NewDecoder(&buffer, f)
	)

	assert.NoError(encoder.Encode(&original))
	assert.True(buffer.Len() > 0)
	assert.Equal(ServiceAliveMessageType, original.Type)
	assert.NoError(decoder.Decode(&decoded))
	assert.Equal(original, decoded)
}

func TestServiceAlive(t *testing.T) {
	for _, format := range allFormats {
		t.Run(fmt.Sprintf("Encode%s", format), func(t *testing.T) {
			testServiceAliveEncode(t, format)
		})
	}
}

func testUnknownEncode(t *testing.T, f Format) {
	var (
		assert   = assert.New(t)
		original = Unknown{}

		decoded Unknown

		buffer  bytes.Buffer
		encoder = NewEncoder(&buffer, f)
		decoder = NewDecoder(&buffer, f)
	)

	assert.NoError(encoder.Encode(&original))
	assert.True(buffer.Len() > 0)
	assert.Equal(UnknownMessageType, original.Type)
	assert.NoError(decoder.Decode(&decoded))
	assert.Equal(original, decoded)
}

func TestUnknown(t *testing.T) {
	for _, format := range allFormats {
		t.Run(fmt.Sprintf("Encode%s", format), func(t *testing.T) {
			testUnknownEncode(t, format)
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

			got, err := MessageFromEnviron(mapToEnviron(m))

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

func TestEnviron_SimpleRequestResponse(t *testing.T) {
	tests := []struct {
		description string
		want        SimpleRequestResponse
		err         error
	}{
		{
			description: "empty",
		}, {
			description: "simple",
			want: SimpleRequestResponse{
				Source:          "source",
				Destination:     "destination",
				TransactionUUID: "transaction_uuid",
				PartnerIDs:      []string{"foo", "bar", "baz"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			msg := tc.want
			m := msg.ToEnvironForm()
			assert.NotNil(t, m)

			got, err := SimpleRequestResponseFromEnviron(mapToEnviron(m))

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

func TestEnviron_SimpleEvent(t *testing.T) {
	tests := []struct {
		description string
		want        SimpleEvent
		err         error
	}{
		{
			description: "empty",
		}, {
			description: "simple",
			want: SimpleEvent{
				Source:      "source",
				Destination: "destination",
				PartnerIDs:  []string{"foo", "bar", "baz"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			msg := tc.want
			m := msg.ToEnvironForm()
			assert.NotNil(t, m)

			got, err := SimpleEventFromEnviron(mapToEnviron(m))

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

func TestEnviron_CRUD(t *testing.T) {
	tests := []struct {
		description string
		want        CRUD
		err         error
	}{
		{
			description: "empty",
		}, {
			description: "simple",
			want: CRUD{
				Source:      "source",
				Destination: "destination",
				PartnerIDs:  []string{"foo", "bar", "baz"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			msg := tc.want
			m := msg.ToEnvironForm()
			assert.NotNil(t, m)

			got, err := CRUDFromEnviron(mapToEnviron(m))

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

func TestEnviron_ServiceRegistration(t *testing.T) {
	tests := []struct {
		description string
		want        ServiceRegistration
		err         error
	}{
		{
			description: "empty",
		}, {
			description: "simple",
			want: ServiceRegistration{
				ServiceName: "service_name",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			msg := tc.want
			m := msg.ToEnvironForm()
			assert.NotNil(t, m)

			got, err := ServiceRegistrationFromEnviron(mapToEnviron(m))

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

func TestEnviron_ServiceAlive(t *testing.T) {
	tests := []struct {
		description string
		want        ServiceAlive
		err         error
	}{
		{
			description: "empty",
		}, {
			description: "simple",
			want: ServiceAlive{
				Type: ServiceAliveMessageType,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			msg := tc.want
			m := msg.ToEnvironForm()
			assert.NotNil(t, m)

			got, err := ServiceAliveFromEnviron(mapToEnviron(m))

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

func TestEnviron_Unknown(t *testing.T) {
	tests := []struct {
		description string
		want        Unknown
		err         error
	}{
		{
			description: "empty",
		}, {
			description: "simple",
			want: Unknown{
				Type: UnknownMessageType,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			msg := tc.want
			m := msg.ToEnvironForm()
			assert.NotNil(t, m)

			got, err := UnknownFromEnviron(mapToEnviron(m))

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
