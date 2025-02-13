// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"errors"
)

//go:generate go install github.com/tinylib/msgp@latest
//go:generate msgp -io=false -tests=false
//msgp:replace MessageType with:int64
//msgp:replace QOSValue with:int
//msgp:tag json
//msgp:ignore Authorization
//msgp:ignore ServiceRegistration
//msgp:ignore ServiceAlive
//msgp:ignore Unknown
//msgp:ignore SimpleRequestResponse
//msgp:ignore SimpleEvent
//msgp:ignore CRUD
//msgp:newtime
//go:generate sed -i "s/MarshalMsg/marshalMsg/g"     messages_gen.go
//go:generate sed -i "s/UnmarshalMsg/unmarshalMsg/g" messages_gen.go
//go:generate sed -i "s/Msgsize/msgsize/g"           messages_gen.go

var (
	ErrInvalidMessageType   = errors.New("invalid message type")
	ErrMessageIsInvalid     = errors.New("message is invalid")
	ErrSourceRequired       = errors.New("source is required")
	ErrDestRequired         = errors.New("dest is required")
	ErrTransactionRequired  = errors.New("transaction_uuid is required")
	ErrUnsupportedFieldsSet = errors.New("unsupported fields set")
	ErrNotUTF8              = errors.New("field contains non-utf-8 characters")
	ErrInvalidQOSValue      = errors.New("qos value is invalid")
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

	// Path is the path to which to apply the payload.
	Path string `json:"path,omitempty"`

	// Payload is the payload for this message.  It's format is expected to match ContentType.
	//
	// For JSON, this field must be a UTF-8 string.  Binary payloads may be base64-encoded.
	//
	// For msgpack, this field is encoded as binary.
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

func (msg *Message) from(m *Message) {
	msg.Type = m.Type
	msg.Source = m.Source
	msg.Destination = m.Destination
	msg.TransactionUUID = m.TransactionUUID
	msg.ContentType = m.ContentType
	msg.Accept = m.Accept
	msg.Status = m.Status
	msg.RequestDeliveryResponse = m.RequestDeliveryResponse
	msg.Headers = m.Headers
	msg.Metadata = m.Metadata
	msg.Path = m.Path
	msg.Payload = m.Payload
	msg.ServiceName = m.ServiceName
	msg.URL = m.URL
	msg.PartnerIDs = m.PartnerIDs
	msg.SessionID = m.SessionID
	msg.QualityOfService = m.QualityOfService
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

// Validate checks the message for correctness.  If the message is invalid, an
// error is returned.
func (msg *Message) Validate(validators ...Processor) error {
	return Validate(msg, validators...)
}

// MessageStructs is a union of all WRP message types.  This type is useful for
// generics based code that needs to handle multiple message types.
type MessageStructs interface {
	Message |
		Authorization |
		SimpleRequestResponse |
		SimpleEvent |
		CRUD |
		ServiceRegistration |
		ServiceAlive |
		Unknown
}

// -----------------------------------------------------------------------------

// converter is a compile-time helper interface that ensures all message types
// implement the From, To and Validate methods.  This interface is not intended
// for use in client code.  It is also used by test code.
type converter interface {
	From(*Message, ...Processor) error
	To(validators ...Processor) (*Message, error)
	Validate(...Processor) error

	// to converts the specific struct to a Message struct, with no
	// error checking.  This allows validateTo to call this function and avoid any
	// circular loops.
	to() *Message
	from(*Message)
}

// -----------------------------------------------------------------------------

// SimpleRequestResponse represents a WRP message of type SimpleRequestResponseMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#simple-request-response-definition
type SimpleRequestResponse struct {
	Source                  string `required:"" locator:""`
	Destination             string `required:"" locator:""`
	TransactionUUID         string `required:""`
	ContentType             string
	Accept                  string
	Status                  *int64
	RequestDeliveryResponse *int64
	PartnerIDs              []string
	Headers                 []string
	Metadata                map[string]string
	QualityOfService        QOSValue
	SessionID               string
	Payload                 []byte
}

var _ converter = (*SimpleRequestResponse)(nil)

// From converts a Message struct to a SimpleRequestResponse struct.  The
// Message struct is validated before being converted.  If the Message struct is
// invalid, an error is returned.
func (srr *SimpleRequestResponse) From(msg *Message, validators ...Processor) error {
	if err := Validate(msg, validators...); err != nil {
		return err
	}

	srr.from(msg)
	return nil
}

func (srr *SimpleRequestResponse) from(msg *Message) {
	srr.Source = msg.Source
	srr.Destination = msg.Destination
	srr.TransactionUUID = msg.TransactionUUID
	srr.ContentType = msg.ContentType
	srr.Accept = msg.Accept
	srr.Status = msg.Status
	srr.RequestDeliveryResponse = msg.RequestDeliveryResponse
	srr.PartnerIDs = trimPartnerIDs(msg.PartnerIDs)
	srr.Headers = msg.Headers
	srr.Metadata = msg.Metadata
	srr.QualityOfService = msg.QualityOfService
	srr.SessionID = msg.SessionID
	srr.Payload = msg.Payload
}

// To converts the SimpleRequestResponse struct to a Message struct.  The
// Message struct is validated before being returned.  If the Message struct
// is invalid, an error is returned.
func (srr *SimpleRequestResponse) To(validators ...Processor) (*Message, error) {
	return validateTo(srr, validators...)
}

func (srr *SimpleRequestResponse) to() *Message {
	return &Message{
		Type:                    SimpleRequestResponseMessageType,
		Source:                  srr.Source,
		Destination:             srr.Destination,
		TransactionUUID:         srr.TransactionUUID,
		ContentType:             srr.ContentType,
		Accept:                  srr.Accept,
		Status:                  srr.Status,
		RequestDeliveryResponse: srr.RequestDeliveryResponse,
		PartnerIDs:              trimPartnerIDs(srr.PartnerIDs),
		Headers:                 srr.Headers,
		Metadata:                srr.Metadata,
		QualityOfService:        srr.QualityOfService,
		SessionID:               srr.SessionID,
		Payload:                 srr.Payload,
	}
}

// Validate checks the SimpleRequestResponse struct for correctness.  If the
// SimpleRequestResponse struct is invalid, an error is returned.
func (srr *SimpleRequestResponse) Validate(validators ...Processor) error {
	return Validate(srr, validators...)
}

// SetStatus simplifies setting the optional Status field.
func (srr *SimpleRequestResponse) SetStatus(value int64) *SimpleRequestResponse {
	srr.Status = &value
	return srr
}

// SetRequestDeliveryResponse simplifies setting the optional
// RequestDeliveryResponse field.
func (srr *SimpleRequestResponse) SetRequestDeliveryResponse(value int64) *SimpleRequestResponse {
	srr.RequestDeliveryResponse = &value
	return srr
}

// -----------------------------------------------------------------------------

// SimpleEvent represents a WRP message of type SimpleEventMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#simple-event-definition
type SimpleEvent struct {
	Source                  string `required:"" locator:""`
	Destination             string `required:"" locator:""`
	TransactionUUID         string `suggested:""`
	ContentType             string
	RequestDeliveryResponse *int64
	PartnerIDs              []string
	Headers                 []string
	Metadata                map[string]string
	SessionID               string
	QualityOfService        QOSValue
	Payload                 []byte
}

var _ converter = (*SimpleEvent)(nil)

// From converts a Message struct to a SimpleEvent struct.  The Message struct is
// validated before being converted.  If the Message struct is invalid, an error
// is returned.
func (se *SimpleEvent) From(msg *Message, validators ...Processor) error {
	if err := Validate(msg, validators...); err != nil {
		return err
	}

	se.from(msg)
	return nil
}

func (se *SimpleEvent) from(msg *Message) {
	se.Source = msg.Source
	se.Destination = msg.Destination
	se.TransactionUUID = msg.TransactionUUID
	se.ContentType = msg.ContentType
	se.RequestDeliveryResponse = msg.RequestDeliveryResponse
	se.PartnerIDs = trimPartnerIDs(msg.PartnerIDs)
	se.Headers = msg.Headers
	se.Metadata = msg.Metadata
	se.SessionID = msg.SessionID
	se.QualityOfService = msg.QualityOfService
	se.Payload = msg.Payload
}

// To converts the SimpleEvent struct to a Message struct.  The Message struct is
// validated before being returned.  If the Message struct is invalid, an error
// is returned.
func (se *SimpleEvent) To(validators ...Processor) (*Message, error) {
	return validateTo(se, validators...)
}

func (se *SimpleEvent) to() *Message {
	return &Message{
		Type:                    SimpleEventMessageType,
		Source:                  se.Source,
		Destination:             se.Destination,
		TransactionUUID:         se.TransactionUUID,
		ContentType:             se.ContentType,
		RequestDeliveryResponse: se.RequestDeliveryResponse,
		PartnerIDs:              trimPartnerIDs(se.PartnerIDs),
		Headers:                 se.Headers,
		Metadata:                se.Metadata,
		SessionID:               se.SessionID,
		QualityOfService:        se.QualityOfService,
		Payload:                 se.Payload,
	}
}

// Validate checks the SimpleEvent struct for correctness.  If the SimpleEvent
// struct is invalid, an error is returned.
func (se *SimpleEvent) Validate(validators ...Processor) error {
	return Validate(se, validators...)
}

// SetRequestDeliveryResponse simplifies setting the optional
// RequestDeliveryResponse field.
func (se *SimpleEvent) SetRequestDeliveryResponse(value int64) *SimpleEvent {
	se.RequestDeliveryResponse = &value
	return se
}

// -----------------------------------------------------------------------------

// CRUD represents a WRP message of one of the CRUD message types.  This type does not implement BeforeEncode,
// and so does not automatically set the Type field.  Client code must set the Type code appropriately.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#crud-message-definition
type CRUD struct {
	Type                    MessageType `required:""`
	Source                  string      `required:"" locator:""`
	Destination             string      `required:"" locator:""`
	TransactionUUID         string      `required:""`
	ContentType             string
	Accept                  string
	Status                  *int64
	Path                    string
	RequestDeliveryResponse *int64
	PartnerIDs              []string
	Headers                 []string
	Metadata                map[string]string
	QualityOfService        QOSValue
	SessionID               string
	Payload                 []byte
}

var _ converter = (*CRUD)(nil)

// From converts a Message struct to a CRUD struct.  The Message struct is
// validated before being converted.  If the Message struct is invalid, an error
// is returned.
func (c *CRUD) From(msg *Message, validators ...Processor) error {
	if err := Validate(msg, validators...); err != nil {
		return err
	}

	c.from(msg)
	return nil
}

func (c *CRUD) from(msg *Message) {
	c.Type = msg.Type
	c.Source = msg.Source
	c.Destination = msg.Destination
	c.TransactionUUID = msg.TransactionUUID
	c.ContentType = msg.ContentType
	c.Accept = msg.Accept
	c.Status = msg.Status
	c.Path = msg.Path
	c.RequestDeliveryResponse = msg.RequestDeliveryResponse
	c.PartnerIDs = trimPartnerIDs(msg.PartnerIDs)
	c.Headers = msg.Headers
	c.Metadata = msg.Metadata
	c.QualityOfService = msg.QualityOfService
	c.SessionID = msg.SessionID
	c.Payload = msg.Payload
}

// To converts the CRUD struct to a Message struct.  The Message struct is
// validated before being returned.  If the Message struct is invalid, an error
// is returned.
func (c *CRUD) To(validators ...Processor) (*Message, error) {
	return validateTo(c, validators...)
}

func (c *CRUD) to() *Message {
	return &Message{
		Type:                    c.Type,
		Source:                  c.Source,
		Destination:             c.Destination,
		TransactionUUID:         c.TransactionUUID,
		ContentType:             c.ContentType,
		Accept:                  c.Accept,
		Status:                  c.Status,
		Path:                    c.Path,
		RequestDeliveryResponse: c.RequestDeliveryResponse,
		PartnerIDs:              trimPartnerIDs(c.PartnerIDs),
		Headers:                 c.Headers,
		Metadata:                c.Metadata,
		QualityOfService:        c.QualityOfService,
		SessionID:               c.SessionID,
		Payload:                 c.Payload,
	}
}

// Validate checks the CRUD struct for correctness.  If the CRUD struct is
// invalid, an error is returned.
func (c *CRUD) Validate(validators ...Processor) error {
	return Validate(c, validators...)
}

// SetStatus simplifies setting the optional Status field.
func (c *CRUD) SetStatus(value int64) *CRUD {
	c.Status = &value
	return c
}

// SetRequestDeliveryResponse simplifies setting the optional
// RequestDeliveryResponse.
func (c *CRUD) SetRequestDeliveryResponse(value int64) *CRUD {
	c.RequestDeliveryResponse = &value
	return c
}

// -----------------------------------------------------------------------------

// ServiceRegistration represents a WRP message of type ServiceRegistrationMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#on-device-service-registration-message-definition
type ServiceRegistration struct {
	ServiceName string `required:""`
	URL         string `required:""`
}

var _ converter = (*ServiceRegistration)(nil)

// From converts a Message struct to a ServiceRegistration struct.  The Message
// struct is validated before being converted.  If the Message struct is invalid,
// an error is returned.
func (sr *ServiceRegistration) From(msg *Message, validators ...Processor) error {
	if err := Validate(msg, validators...); err != nil {
		return err
	}

	sr.from(msg)
	return nil
}

func (sr *ServiceRegistration) from(msg *Message) {
	sr.ServiceName = msg.ServiceName
	sr.URL = msg.URL
}

// To converts the ServiceRegistration struct to a Message struct.  The Message
// struct is validated before being returned.  If the Message struct is invalid,
// an error is returned.
func (sr *ServiceRegistration) To(validators ...Processor) (*Message, error) {
	return validateTo(sr, validators...)
}

func (sr *ServiceRegistration) to() *Message {
	return &Message{
		Type:        ServiceRegistrationMessageType,
		ServiceName: sr.ServiceName,
		URL:         sr.URL,
	}
}

// Validate checks the ServiceRegistration struct for correctness.  If the
// ServiceRegistration struct is invalid, an error is returned.
func (sr *ServiceRegistration) Validate(validators ...Processor) error {
	return Validate(sr, validators...)
}

// -----------------------------------------------------------------------------

// ServiceAlive represents a WRP message of type ServiceAliveMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#on-device-service-alive-message-definition
type ServiceAlive struct{}

var _ converter = (*ServiceAlive)(nil)

// From converts a Message struct to a ServiceAlive struct.  The Message struct
// is validated before being converted.  If the Message struct is invalid, an
// error is returned.
func (sa *ServiceAlive) From(msg *Message, validators ...Processor) error {
	if err := Validate(msg, validators...); err != nil {
		return err
	}
	return nil
}

func (sa *ServiceAlive) from(msg *Message) {}

// To converts the ServiceAlive struct to a Message struct.  The Message struct
// is validated before being returned.  If the Message struct is invalid, an
// error is returned.
func (sa *ServiceAlive) To(validators ...Processor) (*Message, error) {
	return validateTo(sa, validators...)
}

func (sa *ServiceAlive) to() *Message {
	return &Message{
		Type: ServiceAliveMessageType,
	}
}

// Validate checks the ServiceAlive struct for correctness.  If the ServiceAlive
// struct is invalid, an error is returned.
func (sa *ServiceAlive) Validate(validators ...Processor) error {
	return Validate(sa, validators...)
}

// -----------------------------------------------------------------------------

// Unknown represents a WRP message of type UnknownMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#unknown-message-definition
type Unknown struct{}

var _ converter = (*Unknown)(nil)

// From converts a Message struct to an Unknown struct.  The Message struct is
// validated before being converted.  If the Message struct is invalid, an error
// is returned.
func (u *Unknown) From(msg *Message, validators ...Processor) error {
	if err := Validate(msg, validators...); err != nil {
		return err
	}

	return nil
}

func (u *Unknown) from(msg *Message) {}

// To converts the Unknown struct to a Message struct.  The Message struct is
// validated before being returned.  If the Message struct is invalid, an error
// is returned.
func (u *Unknown) To(validators ...Processor) (*Message, error) {
	return validateTo(u, validators...)
}

func (u *Unknown) to() *Message {
	return &Message{
		Type: UnknownMessageType,
	}
}

// Validate checks the Unknown struct for correctness.  If the Unknown struct is
// invalid, an error is returned.
func (u *Unknown) Validate(validators ...Processor) error {
	return Validate(u, validators...)
}

// -----------------------------------------------------------------------------

// Authorization is a message type that represents an authorization message.
type Authorization struct {
	Status int64 `required:""`
}

var _ converter = (*Authorization)(nil)

// From converts a Message struct to an Authorization struct.  The Message struct
// is validated before being converted.  If the Message struct is invalid, an
// error is returned.
func (a *Authorization) From(msg *Message, validators ...Processor) error {
	if err := Validate(msg, validators...); err != nil {
		return err
	}

	a.from(msg)
	return nil
}

func (a *Authorization) from(msg *Message) {
	a.Status = *msg.Status
}

// To converts the Authorization struct to a Message struct.  The Message struct
// is validated before being returned.  If the Message struct is invalid, an
// error is returned.
func (a *Authorization) To(validators ...Processor) (*Message, error) {
	return validateTo(a, validators...)
}

func (a *Authorization) to() *Message {
	return &Message{
		Type:   AuthorizationMessageType,
		Status: &a.Status,
	}
}

// Validate checks the Authorization struct for correctness.  If the
// Authorization struct is invalid, an error is returned.
func (a *Authorization) Validate(validators ...Processor) error {
	return Validate(a, validators...)
}

// -----------------------------------------------------------------------------

func trimPartnerIDs(partners []string) []string {
	trimmed := make([]string, 0, len(partners))
	for _, id := range partners {
		if id != "" {
			trimmed = append(trimmed, id)
		}
	}

	if len(trimmed) == 0 {
		return nil
	}
	return trimmed
}
