// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"net/http"
	"regexp"
)

//go:generate go install github.com/tinylib/msgp@latest
//go:generate msgp -io=false
//msgp:replace MessageType with:int64
//msgp:replace QOSValue with:int
//msgp:tag json
//msgp:newtime

var (
	// eventPattern is the precompiled regex that selects the top level event
	// classifier
	eventPattern = regexp.MustCompile(`^event:(?P<event>[^/]+)`)
)

// Message is the union of all WRP fields, made optional (except for Type).  This type is
// useful for transcoding streams, since deserializing from non-msgpack formats like JSON
// has some undesirable side effects.
//
// IMPORTANT: Anytime a new WRP field is added to any message, or a new message with new fields,
// those new fields must be added to this struct for transcoding to work properly.  And of course:
// update the tests!
//
// For server code that sends specific messages, use one of the other WRP structs in this package.
//
// For server code that needs to read one format and emit another, use this struct as it allows
// client code to transcode without knowledge of the exact type of message.
//
// swagger:response Message
type Message struct {
	// Type is the message type for the message.
	//
	// example: SimpleRequestResponseMessageType
	Type MessageType `json:"msg_type" env:"WRP_MSG_TYPE" http:"X-Xmidt-Message-Type,X-Midt-Msg-Type"`

	// Source is the device_id name of the device originating the request or response.
	//
	// example: dns:talaria.xmidt.example.com
	Source string `json:"source,omitempty" env:"WRP_SOURCE,omitempty" http:"X-Xmidt-Source,X-Midt-Source,omitempty"`

	// Destination is the device_id name of the target device of the request or response.
	//
	// example: event:device-status/mac:ffffffffdae4/online
	Destination string `json:"dest,omitempty" env:"WRP_DEST,omitempty" http:"X-Webpa-Device-Name,X-Xmidt-Dest,X-Midt-Dest,omitempty"`

	// TransactionUUID The transaction key for the message
	//
	// example: 546514d4-9cb6-41c9-88ca-ccd4c130c525
	TransactionUUID string `json:"transaction_uuid,omitempty" env:"WRP_TRANSACTION_UUID,omitempty" http:"X-Xmidt-Transaction-Uuid,X-Midt-Transaction-Uuid,omitempty"`

	// ContentType The media type of the payload.
	//
	// example: json
	ContentType string `json:"content_type,omitempty" env:"WRP_CONTENT_TYPE,omitempty" http:"Content-Type,omitempty"`

	// Accept is the media type accepted in the response.
	Accept string `json:"accept,omitempty" env:"WRP_ACCEPT,omitempty" http:"X-Xmidt-Accept,X-Midt-Accept,omitempty"`

	// Status is the response status from the originating service.
	Status *int64 `json:"status,omitempty" env:"WRP_STATUS,omitempty" http:"X-Xmidt-Status,X-Midt-Status,omitempty"`

	// RequestDeliveryResponse is the request delivery response is the delivery result
	// of the previous (implied request) message with a matching transaction_uuid
	RequestDeliveryResponse *int64 `json:"rdr,omitempty" env:"WRP_RDR,omitempty" http:"X-Xmidt-Request-Delivery-Response,X-Midt-Request-Delivery-Response,omitempty"`

	// Headers is the headers associated with the payload.
	Headers []string `json:"headers,omitempty" env:"WRP_HEADERS,omitempty,multiline" http:"X-Xmidt-Headers,X-Midt-Headers,omitempty,multiline"`

	// Metadata is the map of name/value pairs used by consumers of WRP messages for filtering & other purposes.
	//
	// example: {"/boot-time":"1542834188","/last-reconnect-reason":"spanish inquisition"}
	Metadata map[string]string `json:"metadata,omitempty" env:"WRP_METADATA,omitempty" http:"X-Xmidt-Metadata,X-Midt-Metadata,omitempty"`

	// Path is the path to which to apply the payload.
	Path string `json:"path,omitempty" env:"WRP_PATH,omitempty" http:"X-Xmidt-Path,X-Midt-Path,omitempty"`

	// Payload is the payload for this message.  It's format is expected to match ContentType.
	//
	// For JSON, this field must be a UTF-8 string.  Binary payloads may be base64-encoded.
	//
	// For msgpack, this field may be raw binary or a UTF-8 string.
	//
	// example: eyJpZCI6IjUiLCJ0cyI6IjIwMTktMDItMTJUMTE6MTA6MDIuNjE0MTkxNzM1WiIsImJ5dGVzLXNlbnQiOjAsIm1lc3NhZ2VzLXNlbnQiOjEsImJ5dGVzLXJlY2VpdmVkIjowLCJtZXNzYWdlcy1yZWNlaXZlZCI6MH0=
	Payload []byte `json:"payload,omitempty" env:"WRP_PAYLOAD,omitempty"`

	// ServiceName is the originating point of the request or response.
	ServiceName string `json:"service_name,omitempty" env:"WRP_SERVICE_NAME,omitempty" http:"X-Xmidt-Service-Name,X-Midt-Service-Name,omitempty"`

	// URL is the url to use when connecting to the nanomsg pipeline.
	URL string `json:"url,omitempty" env:"WRP_URL,omitempty" http:"X-Xmidt-Url,X-Midt-Url,omitempty"`

	// PartnerIDs is the list of partner ids the message is meant to target.
	//
	// example: ["hello","world"]
	PartnerIDs []string `json:"partner_ids,omitempty" env:"WRP_PARTNER_IDS,omitempty" http:"X-Xmidt-Partner-Id,X-Midt-Partner-Id,omitempty"`

	// SessionID is the ID for the current session.
	SessionID string `json:"session_id,omitempty" env:"WRP_SESSION_ID,omitempty" http:"X-Xmidt-Session-Id,X-Midt-Session-Id,omitempty"`

	// QualityOfService is the qos value associated with this message.  Values between 0 and 99, inclusive,
	// are defined by the wrp spec.  Negative values are assumed to be zero, and values larger than 99
	// are assumed to be 99.
	QualityOfService QOSValue `json:"qos" env:"WRP_QOS,omitempty" http:"X-Xmidt-Qos,X-Midt-Qos,omitempty"`
}

func (msg *Message) FindEventStringSubMatch() string {
	return findEventStringSubMatch(msg.Destination)
}

func (msg *Message) MessageType() MessageType {
	return msg.Type
}

func (msg *Message) To() string {
	return msg.Destination
}

func (msg *Message) From() string {
	return msg.Source
}

func (msg *Message) IsTransactionPart() bool {
	return msg.Type.RequiresTransaction() && len(msg.TransactionUUID) > 0
}

func (msg *Message) TransactionKey() string {
	return msg.TransactionUUID
}

// IsQOSAckPart determines whether or not a message can QOS ack.
func (msg *Message) IsQOSAckPart() bool {
	if !msg.Type.SupportsQOSAck() {
		return false
	}

	// https://xmidt.io/docs/wrp/basics/#qos-description-qos
	switch msg.QualityOfService.Level() {
	case QOSMedium, QOSHigh, QOSCritical:
		return true
	default:
		return false
	}
}

// Response creates a new message that is a response to the current message.
// The following fields are copied from the current message:
//   - Type
//   - Source (becomes the new Destination)
//   - Destination (becomes the new Source)
//   - TransactionUUID
//   - RequestDeliveryResponse
//   - QualityOfService
//   - SessionID
//   - PartnerIDs
func (msg *Message) Response() *Message {
	return &Message{
		Type:                    msg.Type,
		Destination:             msg.Source,
		Source:                  msg.Destination,
		TransactionUUID:         msg.TransactionUUID,
		RequestDeliveryResponse: msg.RequestDeliveryResponse,
		PartnerIDs:              msg.PartnerIDs,
		SessionID:               msg.SessionID,
		QualityOfService:        msg.QualityOfService,
	}
}

// SetStatus simplifies setting the optional Status field, which is a pointer type tagged with omitempty.
func (msg *Message) SetStatus(value int64) *Message {
	msg.Status = &value
	return msg
}

// SetRequestDeliveryResponse simplifies setting the optional RequestDeliveryResponse field, which is a pointer type tagged with omitempty.
func (msg *Message) SetRequestDeliveryResponse(value int64) *Message {
	msg.RequestDeliveryResponse = &value
	return msg
}

// TrimmedPartnerIDs returns a copy of the PartnerIDs field with all empty strings removed.
func (msg *Message) TrimmedPartnerIDs() []string {
	trimmed := make([]string, 0, len(msg.PartnerIDs))
	for _, id := range msg.PartnerIDs {
		if id != "" {
			trimmed = append(trimmed, id)
		}
	}
	return trimmed
}

// ToEnvironForm converts the message to a map of strings suitable for
// use with os.Setenv().
func (msg *Message) ToEnvironForm() map[string]string {
	return toEnvMap(msg)
}

// NewMessageFromEnviron creates a new Message from an array of strings, such as
// that returned by os.Environ().
func NewMessageFromEnviron(env []string) (*Message, error) {
	var msg Message
	err := fromEnvMap(env, &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

// ToHeaderForm converts the message to a map of strings suitable for use with
// http.Header.Set().  If existing is provided, the new headers are applied to
// the existing headers, overwriting any that are in conflict.  Only the first
// existing set of headers is modified, if present.
func (msg *Message) ToHeaderForm(existing ...http.Header) (headers http.Header, payload []byte) {
	existing = append(existing, http.Header{})

	if len(msg.Payload) > 0 {
		defaultAHeader(&existing[0], "Content-Type", []string{MimeTypeOctetStream})
	}
	toHeaders(msg, existing[0])

	return existing[0], msg.Payload
}

// MessageFromHeader creates a new Message from an http.Header and a payload.
func NewMessageFromHeaders(header http.Header, payload []byte) (*Message, error) {
	var msg Message

	if len(payload) > 0 {
		defaultAHeader(&header, "Content-Type", []string{MimeTypeOctetStream})
	}

	err := fromHeaders(header, &msg)
	if err != nil {
		return nil, err
	}
	msg.Payload = payload

	return &msg, nil
}

func findEventStringSubMatch(s string) string {
	var match = eventPattern.FindStringSubmatch(s)

	event := "unknown"
	if match != nil {
		event = match[1]
	}

	return event
}
