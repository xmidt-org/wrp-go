package wrp

import "regexp"

//go:generate codecgen -st "wrp" -o messages_codec.go messages.go

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
	// Type The message type for the message
	//
	// required: true
	// example: 4
	Type MessageType `wrp:"msg_type" json:"msg_type"`

	// Source The device_id name of the device originating the request or response.
	//
	// required: false
	// example: dns:talaria.xmidt.example.com
	Source string `wrp:"source,omitempty" json:"source,omitempty"`

	// Destination The device_id name of the target device of the request or response.
	//
	// required: false
	// example: event:device-status/mac:ffffffffdae4/online
	Destination string `wrp:"dest,omitempty" json:"dest,omitempty"`

	// TransactionUUID The transaction key for the message
	//
	// required: false
	// example: 546514d4-9cb6-41c9-88ca-ccd4c130c525
	TransactionUUID string `wrp:"transaction_uuid,omitempty" json:"transaction_uuid,omitempty"`

	// ContentType The media type of the payload.
	//
	// required: false
	// example: json
	ContentType string `wrp:"content_type,omitempty" json:"content_type,omitempty"`

	// Accept  The media type accepted in the response.
	//
	// required: false
	Accept string `wrp:"accept,omitempty" json:"accept,omitempty"`

	// Status The response status from the originating service.
	//
	// required: false
	Status *int64 `wrp:"status,omitempty" json:"status,omitempty"`

	// RequestDeliveryResponse The request delivery response is the delivery result of the previous (implied request)
	// message with a matching transaction_uuid
	//
	// required: false
	RequestDeliveryResponse *int64 `wrp:"rdr,omitempty" json:"rdr,omitempty"`

	// Headers The headers associated with the payload.
	//
	// required: false
	Headers []string `wrp:"headers,omitempty" json:"headers,omitempty"`

	// Metadata The map of name/value pairs used by consumers of WRP messages for filtering & other purposes.
	//
	// required: false
	// example: {"/boot-time":1542834188,"/last-reconnect-reason":"spanish inquisition"}
	Metadata map[string]string `wrp:"metadata,omitempty" json:"metadata,omitempty"`

	// Spans An array of arrays of timing values as a list in the format: "parent" (string), "name" (string),
	// "start time" (int), "duration" (int), "status" (int)
	//
	// required: false
	Spans [][]string `wrp:"spans,omitempty" json:"spans,omitempty"`

	// IncludeSpans (Deprecated) If the timing values should be included in the response.
	//
	// required: false
	IncludeSpans *bool `wrp:"include_spans,omitempty" json:"include_spans,omitempty"`

	// Path The path to which to apply the payload.
	//
	// required: false
	Path string `wrp:"path,omitempty" json:"path,omitempty"`

	// Payload The string encoded of the ContentType
	//
	// required: false
	// example: eyJpZCI6IjUiLCJ0cyI6IjIwMTktMDItMTJUMTE6MTA6MDIuNjE0MTkxNzM1WiIsImJ5dGVzLXNlbnQiOjAsIm1lc3NhZ2VzLXNlbnQiOjEsImJ5dGVzLXJlY2VpdmVkIjowLCJtZXNzYWdlcy1yZWNlaXZlZCI6MH0=
	Payload []byte `wrp:"payload,omitempty" json:"payload,omitempty"`

	// ServiceName The originating point of the request or response
	//
	// required: false
	ServiceName string `wrp:"service_name,omitempty" json:"service_name,omitempty"`

	// URL The url to use when connecting to the nanomsg pipeline
	//
	// required: false
	URL string `wrp:"url,omitempty" json:"url,omitempty"`

	// PartnerIDs The list of partner ids the message is meant to target.
	//
	// required: false
	// example: ["hello","world"]
	PartnerIDs []string `wrp:"partner_ids,omitempty" json:"partner_ids,omitempty"`
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
// https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol#simple-request-response-definition
type SimpleRequestResponse struct {
	// Type is exposed principally for encoding.  This field *must* be set to SimpleRequestResponseMessageType,
	// and is automatically set by the BeforeEncode method.
	Type                    MessageType       `wrp:"msg_type" json:"msg_type"`
	Source                  string            `wrp:"source" json:"source"`
	Destination             string            `wrp:"dest" json:"dest"`
	ContentType             string            `wrp:"content_type,omitempty" json:"content_type,omitempty"`
	Accept                  string            `wrp:"accept,omitempty" json:"accept,omitempty"`
	TransactionUUID         string            `wrp:"transaction_uuid,omitempty" json:"transaction_uuid,omitempty"`
	Status                  *int64            `wrp:"status,omitempty" json:"status,omitempty"`
	RequestDeliveryResponse *int64            `wrp:"rdr,omitempty" json:"rdr,omitempty"`
	Headers                 []string          `wrp:"headers,omitempty" json:"headers,omitempty"`
	Metadata                map[string]string `wrp:"metadata,omitempty" json:"metadata,omitempty"`
	Spans                   [][]string        `wrp:"spans,omitempty" json:"spans,omitempty"`
	IncludeSpans            *bool             `wrp:"include_spans,omitempty" json:"include_spans,omitempty"`
	Payload                 []byte            `wrp:"payload,omitempty" json:"payload,omitempty"`
	PartnerIDs              []string          `wrp:"partner_ids,omitempty" json:"partner_ids,omitempty"`
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
// https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol#simple-event-definition
type SimpleEvent struct {
	// Type is exposed principally for encoding.  This field *must* be set to SimpleEventMessageType,
	// and is automatically set by the BeforeEncode method.
	Type        MessageType       `wrp:"msg_type" json:"msg_type"`
	Source      string            `wrp:"source" json:"source"`
	Destination string            `wrp:"dest" json:"dest"`
	ContentType string            `wrp:"content_type,omitempty" json:"content_type,omitempty"`
	Headers     []string          `wrp:"headers,omitempty" json:"headers,omitempty"`
	Metadata    map[string]string `wrp:"metadata,omitempty" json:"metadata,omitempty"`
	Payload     []byte            `wrp:"payload,omitempty" json:"payload,omitempty"`
	PartnerIDs  []string          `wrp:"partner_ids,omitempty" json:"partner_ids,omitempty"`
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
// https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol#crud-message-definition
type CRUD struct {
	Type                    MessageType       `wrp:"msg_type" json:"msg_type"`
	Source                  string            `wrp:"source" json:"source"`
	Destination             string            `wrp:"dest" json:"dest"`
	TransactionUUID         string            `wrp:"transaction_uuid,omitempty" json:"transaction_uuid,omitempty"`
	ContentType             string            `wrp:"content_type,omitempty" json:"content_type,omitempty"`
	Headers                 []string          `wrp:"headers,omitempty" json:"headers,omitempty"`
	Metadata                map[string]string `wrp:"metadata,omitempty" json:"metadata,omitempty"`
	Spans                   [][]string        `wrp:"spans,omitempty" json:"spans,omitempty"`
	IncludeSpans            *bool             `wrp:"include_spans,omitempty" json:"include_spans,omitempty"`
	Status                  *int64            `wrp:"status,omitempty" json:"status,omitempty"`
	RequestDeliveryResponse *int64            `wrp:"rdr,omitempty" json:"rdr,omitempty"`
	Path                    string            `wrp:"path" json:"path"`
	Payload                 []byte            `wrp:"payload,omitempty" json:"payload,omitempty"`
	PartnerIDs              []string          `wrp:"partner_ids,omitempty" json:"partner_ids,omitempty"`
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
// https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol#on-device-service-registration-message-definition
type ServiceRegistration struct {
	// Type is exposed principally for encoding.  This field *must* be set to ServiceRegistrationMessageType,
	// and is automatically set by the BeforeEncode method.
	Type        MessageType `wrp:"msg_type" json:"msg_type"`
	ServiceName string      `wrp:"service_name" json:"service_name"`
	URL         string      `wrp:"url" json:"url"`
}

func (msg *ServiceRegistration) BeforeEncode() error {
	msg.Type = ServiceRegistrationMessageType
	return nil
}

// ServiceAlive represents a WRP message of type ServiceAliveMessageType.
//
// https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol#on-device-service-alive-message-definition
type ServiceAlive struct {
	// Type is exposed principally for encoding.  This field *must* be set to ServiceAliveMessageType,
	// and is automatically set by the BeforeEncode method.
	Type MessageType `wrp:"msg_type" json:"msg_type"`
}

func (msg *ServiceAlive) BeforeEncode() error {
	msg.Type = ServiceAliveMessageType
	return nil
}

// Unknown represents a WRP message of type UnknownMessageType.
//
// https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol#unknown-message-definition
type Unknown struct {
	// Type is exposed principally for encoding.  This field *must* be set to UnknownMessageType,
	// and is automatically set by the BeforeEncode method.
	Type MessageType `wrp:"msg_type" json:"msg_type"`
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
