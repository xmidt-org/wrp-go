/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package wrp

import (
	"regexp"
)

//go:generate go install github.com/ugorji/go/codec/codecgen@latest
//go:generate codecgen -st "json" -o messages_codec.go messages.go

var (
	// eventPattern is the precompiled regex that selects the top level event
	// classifier
	eventPattern = regexp.MustCompile(`^event:(?P<event>[^/]+)`)
)

// Typed is implemented by any WRP type which is associated with a MessageType.  All
// message types implement this interface.
type Typed interface {
	// MessageType is the type of message represented by this Typed.
	MessageType() MessageType
}

// Routable describes an object which can be routed.  Implementations will most
// often also be WRP Message instances.  All Routable objects may be passed to
// Encoders and Decoders.
//
// Not all WRP messages are Routable.  Only messages that can be sent through
// routing software (e.g. talaria) implement this interface.
type Routable interface {
	Typed

	// To is the destination of this Routable instance.  It corresponds to the Destination field
	// in WRP messages defined in this package.
	To() string

	// From is the originator of this Routable instance.  It corresponds to the Source field
	// in WRP messages defined in this package.
	From() string

	// IsTransactionPart tests if this message represents part of a transaction.  For this to be true,
	// both (1) the msg_type field must be of a type that participates in transactions and (2) a transaction_uuid
	// must exist in the message (see TransactionKey).
	//
	// If this method returns true, TransactionKey will always return a non-empty string.
	IsTransactionPart() bool

	// TransactionKey corresponds to the transaction_uuid field.  If present, this field is used
	// to match up responses from devices.
	//
	// Not all Routables support transactions, e.g. SimpleEvent.  For those Routable messages that do
	// not possess a transaction_uuid field, this method returns an empty string.
	TransactionKey() string

	// Response produces a new Routable instance which is a response to this one.  The new Routable's
	// destination (From) is set to the original source (To), with the supplied newSource used as the response's source.
	// The requestDeliveryResponse parameter indicates the success or failure of this response.  The underlying
	// type of the returned Routable will be the same as this type, i.e. if this instance is a Message,
	// the returned Routable will also be a Message.
	//
	// If applicable, the response's payload is set to nil.  All other fields are copied as is into the response.
	Response(newSource string, requestDeliveryResponse int64) Routable
}

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
	Type MessageType `json:"msg_type"`

	// Source is the device_id name of the device originating the request or response.
	//
	// example: dns:talaria.xmidt.example.com
	Source string `json:"source,omitempty"`

	// Destination is the device_id name of the target device of the request or response.
	//
	// example: event:device-status/mac:ffffffffdae4/online
	Destination string `json:"dest,omitempty"`

	// TransactionUUID The transaction key for the message
	//
	// example: 546514d4-9cb6-41c9-88ca-ccd4c130c525
	TransactionUUID string `json:"transaction_uuid,omitempty"`

	// ContentType The media type of the payload.
	//
	// example: json
	ContentType string `json:"content_type,omitempty"`

	// Accept is the media type accepted in the response.
	Accept string `json:"accept,omitempty"`

	// Status is the response status from the originating service.
	Status *int64 `json:"status,omitempty"`

	// RequestDeliveryResponse is the request delivery response is the delivery result
	// of the previous (implied request) message with a matching transaction_uuid
	RequestDeliveryResponse *int64 `json:"rdr,omitempty"`

	// Headers is the headers associated with the payload.
	Headers []string `json:"headers,omitempty"`

	// Metadata is the map of name/value pairs used by consumers of WRP messages for filtering & other purposes.
	//
	// example: {"/boot-time":"1542834188","/last-reconnect-reason":"spanish inquisition"}
	Metadata map[string]string `json:"metadata,omitempty"`

	// Spans is an array of arrays of timing values as a list in the format: "parent" (string), "name" (string),
	// "start time" (int), "duration" (int), "status" (int)
	Spans [][]string `json:"spans,omitempty"`

	// IncludeSpans indicates whether timing values should be included in the response.
	//
	// Deprecated: A future version of wrp will remove this field.
	IncludeSpans *bool `json:"include_spans,omitempty"`

	// Path is the path to which to apply the payload.
	Path string `json:"path,omitempty"`

	// Payload is the payload for this message.  It's format is expected to match ContentType.
	//
	// For JSON, this field must be a UTF-8 string.  Binary payloads may be base64-encoded.
	//
	// For msgpack, this field may be raw binary or a UTF-8 string.
	//
	// example: eyJpZCI6IjUiLCJ0cyI6IjIwMTktMDItMTJUMTE6MTA6MDIuNjE0MTkxNzM1WiIsImJ5dGVzLXNlbnQiOjAsIm1lc3NhZ2VzLXNlbnQiOjEsImJ5dGVzLXJlY2VpdmVkIjowLCJtZXNzYWdlcy1yZWNlaXZlZCI6MH0=
	Payload []byte `json:"payload,omitempty"`

	// ServiceName is the originating point of the request or response.
	ServiceName string `json:"service_name,omitempty"`

	// URL is the url to use when connecting to the nanomsg pipeline.
	URL string `json:"url,omitempty"`

	// PartnerIDs is the list of partner ids the message is meant to target.
	//
	// example: ["hello","world"]
	PartnerIDs []string `json:"partner_ids,omitempty"`

	// SessionID is the ID for the current session.
	SessionID string `json:"session_id,omitempty"`

	// QualityOfService is the qos value associated with this message.  Values between 0 and 99, inclusive,
	// are defined by the wrp spec.  Negative values are assumed to be zero, and values larger than 99
	// are assumed to be 99.
	QualityOfService QOSValue `json:"qos"`
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
	return msg.Type.SupportsTransaction() && len(msg.TransactionUUID) > 0
}

func (msg *Message) TransactionKey() string {
	return msg.TransactionUUID
}

func (msg *Message) Response(newSource string, requestDeliveryResponse int64) Routable {
	response := *msg
	response.Destination = msg.Source
	response.Source = newSource
	response.RequestDeliveryResponse = &requestDeliveryResponse
	response.Payload = nil

	return &response
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

// SetIncludeSpans simplifies setting the optional IncludeSpans field, which is a pointer type tagged with omitempty.
func (msg *Message) SetIncludeSpans(value bool) *Message {
	msg.IncludeSpans = &value
	return msg
}

// SimpleRequestResponse represents a WRP message of type SimpleRequestResponseMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#simple-request-response-definition
//
// Deprecated: A future version of wrp will remove this type.
type SimpleRequestResponse struct {
	// Type is exposed principally for encoding.  This field *must* be set to SimpleRequestResponseMessageType,
	// and is automatically set by the BeforeEncode method.
	Type                    MessageType       `json:"msg_type"`
	Source                  string            `json:"source"`
	Destination             string            `json:"dest"`
	ContentType             string            `json:"content_type,omitempty"`
	Accept                  string            `json:"accept,omitempty"`
	TransactionUUID         string            `json:"transaction_uuid,omitempty"`
	Status                  *int64            `json:"status,omitempty"`
	RequestDeliveryResponse *int64            `json:"rdr,omitempty"`
	Headers                 []string          `json:"headers,omitempty"`
	Metadata                map[string]string `json:"metadata,omitempty"`
	Spans                   [][]string        `json:"spans,omitempty"`
	IncludeSpans            *bool             `json:"include_spans,omitempty"`
	Payload                 []byte            `json:"payload,omitempty"`
	PartnerIDs              []string          `json:"partner_ids,omitempty"`
}

func (msg *SimpleRequestResponse) FindEventStringSubMatch() string {
	return findEventStringSubMatch(msg.Destination)
}

// SetStatus simplifies setting the optional Status field, which is a pointer type tagged with omitempty.
func (msg *SimpleRequestResponse) SetStatus(value int64) *SimpleRequestResponse {
	msg.Status = &value
	return msg
}

// SetRequestDeliveryResponse simplifies setting the optional RequestDeliveryResponse field, which is a pointer type tagged with omitempty.
func (msg *SimpleRequestResponse) SetRequestDeliveryResponse(value int64) *SimpleRequestResponse {
	msg.RequestDeliveryResponse = &value
	return msg
}

// SetIncludeSpans simplifies setting the optional IncludeSpans field, which is a pointer type tagged with omitempty.
func (msg *SimpleRequestResponse) SetIncludeSpans(value bool) *SimpleRequestResponse {
	msg.IncludeSpans = &value
	return msg
}

func (msg *SimpleRequestResponse) BeforeEncode() error {
	msg.Type = SimpleRequestResponseMessageType
	return nil
}

func (msg *SimpleRequestResponse) MessageType() MessageType {
	return msg.Type
}

func (msg *SimpleRequestResponse) To() string {
	return msg.Destination
}

func (msg *SimpleRequestResponse) From() string {
	return msg.Source
}

func (msg *SimpleRequestResponse) IsTransactionPart() bool {
	return len(msg.TransactionUUID) > 0
}

func (msg *SimpleRequestResponse) TransactionKey() string {
	return msg.TransactionUUID
}

func (msg *SimpleRequestResponse) Response(newSource string, requestDeliveryResponse int64) Routable {
	response := *msg
	response.Destination = msg.Source
	response.Source = newSource
	response.RequestDeliveryResponse = &requestDeliveryResponse
	response.Payload = nil

	return &response
}

// SimpleEvent represents a WRP message of type SimpleEventMessageType.
//
// This type implements Routable, and as such has a Response method.  However, in actual practice
// failure responses are not sent for messages of this type.  Response is merely supplied in order to satisfy
// the Routable interface.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#simple-event-definition
//
// Deprecated: A future version of wrp will remove this type.
type SimpleEvent struct {
	// Type is exposed principally for encoding.  This field *must* be set to SimpleEventMessageType,
	// and is automatically set by the BeforeEncode method.
	Type        MessageType       `json:"msg_type"`
	Source      string            `json:"source"`
	Destination string            `json:"dest"`
	ContentType string            `json:"content_type,omitempty"`
	Headers     []string          `json:"headers,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Payload     []byte            `json:"payload,omitempty"`
	PartnerIDs  []string          `json:"partner_ids,omitempty"`
	SessionID   string            `json:"session_id,omitempty"`
}

func (msg *SimpleEvent) BeforeEncode() error {
	msg.Type = SimpleEventMessageType
	return nil
}

func (msg *SimpleEvent) MessageType() MessageType {
	return msg.Type
}

func (msg *SimpleEvent) To() string {
	return msg.Destination
}

func (msg *SimpleEvent) From() string {
	return msg.Source
}

// IsTransactionPart for SimpleEvent types always returns false
func (msg *SimpleEvent) IsTransactionPart() bool {
	return false
}

func (msg *SimpleEvent) TransactionKey() string {
	return ""
}

func (msg *SimpleEvent) Response(newSource string, requestDeliveryResponse int64) Routable {
	response := *msg
	response.Destination = msg.Source
	response.Source = newSource
	response.Payload = nil

	return &response
}

// CRUD represents a WRP message of one of the CRUD message types.  This type does not implement BeforeEncode,
// and so does not automatically set the Type field.  Client code must set the Type code appropriately.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#crud-message-definition
//
// Deprecated: A future version of wrp will remove this type.
type CRUD struct {
	Type                    MessageType       `json:"msg_type"`
	Source                  string            `json:"source"`
	Destination             string            `json:"dest"`
	TransactionUUID         string            `json:"transaction_uuid,omitempty"`
	ContentType             string            `json:"content_type,omitempty"`
	Headers                 []string          `json:"headers,omitempty"`
	Metadata                map[string]string `json:"metadata,omitempty"`
	Spans                   [][]string        `json:"spans,omitempty"`
	IncludeSpans            *bool             `json:"include_spans,omitempty"`
	Status                  *int64            `json:"status,omitempty"`
	RequestDeliveryResponse *int64            `json:"rdr,omitempty"`
	Path                    string            `json:"path"`
	Payload                 []byte            `json:"payload,omitempty"`
	PartnerIDs              []string          `json:"partner_ids,omitempty"`
}

// SetStatus simplifies setting the optional Status field, which is a pointer type tagged with omitempty.
func (msg *CRUD) SetStatus(value int64) *CRUD {
	msg.Status = &value
	return msg
}

// SetRequestDeliveryResponse simplifies setting the optional RequestDeliveryResponse field, which is a pointer type tagged with omitempty.
func (msg *CRUD) SetRequestDeliveryResponse(value int64) *CRUD {
	msg.RequestDeliveryResponse = &value
	return msg
}

// SetIncludeSpans simplifies setting the optional IncludeSpans field, which is a pointer type tagged with omitempty.
func (msg *CRUD) SetIncludeSpans(value bool) *CRUD {
	msg.IncludeSpans = &value
	return msg
}

func (msg *CRUD) MessageType() MessageType {
	return msg.Type
}

func (msg *CRUD) To() string {
	return msg.Destination
}

func (msg *CRUD) From() string {
	return msg.Source
}

func (msg *CRUD) IsTransactionPart() bool {
	return len(msg.TransactionUUID) > 0
}

func (msg *CRUD) TransactionKey() string {
	return msg.TransactionUUID
}

func (msg *CRUD) Response(newSource string, requestDeliveryResponse int64) Routable {
	response := *msg
	response.Destination = msg.Source
	response.Source = newSource
	response.RequestDeliveryResponse = &requestDeliveryResponse

	return &response
}

// ServiceRegistration represents a WRP message of type ServiceRegistrationMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#on-device-service-registration-message-definition
//
// Deprecated: A future version of wrp will remove this type.
type ServiceRegistration struct {
	// Type is exposed principally for encoding.  This field *must* be set to ServiceRegistrationMessageType,
	// and is automatically set by the BeforeEncode method.
	Type        MessageType `json:"msg_type"`
	ServiceName string      `json:"service_name"`
	URL         string      `json:"url"`
}

func (msg *ServiceRegistration) BeforeEncode() error {
	msg.Type = ServiceRegistrationMessageType
	return nil
}

// ServiceAlive represents a WRP message of type ServiceAliveMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#on-device-service-alive-message-definition
//
// Deprecated: A future version of wrp will remove this type.
type ServiceAlive struct {
	// Type is exposed principally for encoding.  This field *must* be set to ServiceAliveMessageType,
	// and is automatically set by the BeforeEncode method.
	Type MessageType `json:"msg_type"`
}

func (msg *ServiceAlive) BeforeEncode() error {
	msg.Type = ServiceAliveMessageType
	return nil
}

// Unknown represents a WRP message of type UnknownMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#unknown-message-definition
//
// Deprecated: A future version of wrp will remove this type.
type Unknown struct {
	// Type is exposed principally for encoding.  This field *must* be set to UnknownMessageType,
	// and is automatically set by the BeforeEncode method.
	Type MessageType `json:"msg_type"`
}

func (msg *Unknown) BeforeEncode() error {
	msg.Type = UnknownMessageType
	return nil
}

func findEventStringSubMatch(s string) string {
	var match = eventPattern.FindStringSubmatch(s)

	event := "unknown"
	if match != nil {
		event = match[1]
	}

	return event
}
