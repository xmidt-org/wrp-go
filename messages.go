// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"errors"
	"unicode/utf8"
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

func (msg *Message) validate() error {
	switch msg.Type {
	case AuthorizationMessageType:
		var a Authorization
		return a.From(msg)
	case ServiceRegistrationMessageType:
		var sr ServiceRegistration
		return sr.From(msg)
	case ServiceAliveMessageType:
		var sa ServiceAlive
		return sa.From(msg)
	case SimpleRequestResponseMessageType:
		var srr SimpleRequestResponse
		return srr.From(msg)
	case SimpleEventMessageType:
		var se SimpleEvent
		return se.From(msg)
	case CreateMessageType, RetrieveMessageType, UpdateMessageType, DeleteMessageType:
		var crud CRUD
		return crud.From(msg)
	default:
		var u Unknown
		return u.From(msg)
	}
}

// -----------------------------------------------------------------------------

// converter is a compile-time helper interface that ensures all message types
// implement the From and To methods.  This interface is not intended for use in
// client code.  It is also used by test code.
type converter interface {
	From(*Message) error
	To() (*Message, error)
}

// -----------------------------------------------------------------------------

// SimpleRequestResponse represents a WRP message of type SimpleRequestResponseMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#simple-request-response-definition
type SimpleRequestResponse struct {
	Source                  string `required:""`
	Destination             string `required:""`
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
func (srr *SimpleRequestResponse) From(msg *Message) error {
	if msg.Type != SimpleRequestResponseMessageType {
		return ErrInvalidMessageType
	}
	if msg.Source == "" {
		return errors.Join(ErrMessageIsInvalid, ErrSourceRequired)
	}
	if msg.Destination == "" {
		return errors.Join(ErrMessageIsInvalid, ErrDestRequired)
	}
	if msg.TransactionUUID == "" {
		return errors.Join(ErrMessageIsInvalid, ErrTransactionRequired)
	}

	// Unsupported fields must be empty
	if msg.Path != "" ||
		msg.ServiceName != "" ||
		msg.URL != "" {
		return errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
	}

	if err := validateUTF8(msg); err != nil {
		return err
	}

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
	return nil
}

// To converts the SimpleRequestResponse struct to a Message struct.  The
// Message struct is validated before being returned.  If the Message struct
// is invalid, an error is returned.
func (srr *SimpleRequestResponse) To() (*Message, error) {
	msg := Message{
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

	if err := msg.validate(); err != nil {
		return nil, err
	}

	return &msg, nil
}

// SetStatus simplifies setting the optional Status field, which is a pointer type tagged with omitempty.
func (srr *SimpleRequestResponse) SetStatus(value int64) *SimpleRequestResponse {
	srr.Status = &value
	return srr
}

// SetRequestDeliveryResponse simplifies setting the optional RequestDeliveryResponse field, which is a pointer type tagged with omitempty.
func (srr *SimpleRequestResponse) SetRequestDeliveryResponse(value int64) *SimpleRequestResponse {
	srr.RequestDeliveryResponse = &value
	return srr
}

// -----------------------------------------------------------------------------

// SimpleEvent represents a WRP message of type SimpleEventMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#simple-event-definition
type SimpleEvent struct {
	Source                  string `required:""`
	Destination             string `required:""`
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
func (se *SimpleEvent) From(msg *Message) error {
	if msg.Type != SimpleEventMessageType {
		return ErrInvalidMessageType
	}
	if msg.Source == "" {
		return errors.Join(ErrMessageIsInvalid, ErrSourceRequired)
	}
	if msg.Destination == "" {
		return errors.Join(ErrMessageIsInvalid, ErrDestRequired)
	}

	// Unsupported fields must be empty
	if msg.Accept != "" ||
		msg.Status != nil ||
		msg.Path != "" ||
		msg.ServiceName != "" ||
		msg.URL != "" {
		return errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
	}

	if err := validateUTF8(msg); err != nil {
		return err
	}

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
	return nil
}

// To converts the SimpleEvent struct to a Message struct.  The Message struct is
// validated before being returned.  If the Message struct is invalid, an error
// is returned.
func (se *SimpleEvent) To() (*Message, error) {
	msg := Message{
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

	if err := msg.validate(); err != nil {
		return nil, err
	}

	return &msg, nil
}

// SetRequestDeliveryResponse simplifies setting the optional RequestDeliveryResponse field, which is a pointer type tagged with omitempty.
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
	Source                  string      `required:""`
	Destination             string      `required:""`
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
func (c *CRUD) From(msg *Message) error {
	switch msg.Type {
	case CreateMessageType, RetrieveMessageType, UpdateMessageType, DeleteMessageType:
	default:
		return ErrInvalidMessageType
	}
	if msg.Source == "" {
		return errors.Join(ErrMessageIsInvalid, ErrSourceRequired)
	}
	if msg.Destination == "" {
		return errors.Join(ErrMessageIsInvalid, ErrDestRequired)
	}
	if msg.TransactionUUID == "" {
		return errors.Join(ErrMessageIsInvalid, ErrTransactionRequired)
	}

	// Unsupported fields must be empty
	if msg.ServiceName != "" ||
		msg.URL != "" {
		return errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
	}

	if err := validateUTF8(msg); err != nil {
		return err
	}

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
	return nil
}

// To converts the CRUD struct to a Message struct.  The Message struct is
// validated before being returned.  If the Message struct is invalid, an error
// is returned.
func (c *CRUD) To() (*Message, error) {
	msg := Message{
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

	if err := msg.validate(); err != nil {
		return nil, err
	}

	return &msg, nil
}

// SetStatus simplifies setting the optional Status field, which is a pointer type tagged with omitempty.
func (c *CRUD) SetStatus(value int64) *CRUD {
	c.Status = &value
	return c
}

// SetRequestDeliveryResponse simplifies setting the optional RequestDeliveryResponse field, which is a pointer type tagged with omitempty.
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
func (sr *ServiceRegistration) From(msg *Message) error {
	if msg.Type != ServiceRegistrationMessageType {
		return ErrInvalidMessageType
	}
	if msg.ServiceName == "" {
		return errors.Join(ErrMessageIsInvalid, errors.New("service_name is required"))
	}
	if msg.URL == "" {
		return errors.Join(ErrMessageIsInvalid, errors.New("url is required"))
	}

	// Unsupported fields must be empty
	if msg.Source != "" ||
		msg.Destination != "" ||
		msg.TransactionUUID != "" ||
		msg.ContentType != "" ||
		msg.Accept != "" ||
		msg.Status != nil ||
		msg.RequestDeliveryResponse != nil ||
		len(msg.PartnerIDs) > 0 ||
		len(msg.Headers) > 0 ||
		len(msg.Metadata) > 0 ||
		msg.Path != "" ||
		len(msg.Payload) > 0 ||
		msg.SessionID != "" ||
		msg.QualityOfService != 0 ||
		msg.Payload != nil {
		return errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
	}

	if err := validateUTF8(msg); err != nil {
		return err
	}

	sr.ServiceName = msg.ServiceName
	sr.URL = msg.URL
	return nil
}

// To converts the ServiceRegistration struct to a Message struct.  The Message
// struct is validated before being returned.  If the Message struct is invalid,
// an error is returned.
func (sr *ServiceRegistration) To() (*Message, error) {
	msg := Message{
		Type:        ServiceRegistrationMessageType,
		ServiceName: sr.ServiceName,
		URL:         sr.URL,
	}

	if err := msg.validate(); err != nil {
		return nil, err
	}

	return &msg, nil
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
func (sa *ServiceAlive) From(msg *Message) error {
	if msg.Type != ServiceAliveMessageType {
		return ErrInvalidMessageType
	}

	// Unsupported fields must be empty
	if msg.Source != "" ||
		msg.Destination != "" ||
		msg.TransactionUUID != "" ||
		msg.ContentType != "" ||
		msg.Accept != "" ||
		msg.Status != nil ||
		msg.RequestDeliveryResponse != nil ||
		len(msg.PartnerIDs) > 0 ||
		len(msg.Headers) > 0 ||
		len(msg.Metadata) > 0 ||
		msg.Path != "" ||
		len(msg.Payload) > 0 ||
		msg.ServiceName != "" ||
		msg.URL != "" ||
		msg.SessionID != "" ||
		msg.QualityOfService != 0 ||
		msg.Payload != nil {
		return errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
	}
	return nil
}

// To converts the ServiceAlive struct to a Message struct.  The Message struct
// is validated before being returned.  If the Message struct is invalid, an
// error is returned.
func (sa *ServiceAlive) To() (*Message, error) {
	msg := Message{
		Type: ServiceAliveMessageType,
	}

	if err := msg.validate(); err != nil {
		return nil, err
	}

	return &msg, nil
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
func (u *Unknown) From(msg *Message) error {
	if msg.Type != UnknownMessageType {
		return ErrInvalidMessageType
	}

	// Unsupported fields must be empty
	if msg.Source != "" ||
		msg.Destination != "" ||
		msg.TransactionUUID != "" ||
		msg.ContentType != "" ||
		msg.Accept != "" ||
		msg.Status != nil ||
		msg.RequestDeliveryResponse != nil ||
		len(msg.PartnerIDs) > 0 ||
		len(msg.Headers) > 0 ||
		len(msg.Metadata) > 0 ||
		msg.Path != "" ||
		len(msg.Payload) > 0 ||
		msg.ServiceName != "" ||
		msg.URL != "" ||
		msg.SessionID != "" ||
		msg.QualityOfService != 0 ||
		msg.Payload != nil {
		return errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
	}

	return nil
}

// To converts the Unknown struct to a Message struct.  The Message struct is
// validated before being returned.  If the Message struct is invalid, an error
// is returned.
func (u *Unknown) To() (*Message, error) {
	msg := Message{
		Type: UnknownMessageType,
	}

	if err := msg.validate(); err != nil {
		return nil, err
	}

	return &msg, nil
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
func (a *Authorization) From(msg *Message) error {
	if msg.Type != AuthorizationMessageType {
		return ErrInvalidMessageType
	}
	if msg.Status == nil {
		return errors.Join(ErrMessageIsInvalid, errors.New("status is required"))
	}

	// Unsupported fields must be empty
	if msg.Source != "" ||
		msg.Destination != "" ||
		msg.TransactionUUID != "" ||
		msg.ContentType != "" ||
		msg.Accept != "" ||
		msg.RequestDeliveryResponse != nil ||
		len(msg.PartnerIDs) > 0 ||
		len(msg.Headers) > 0 ||
		len(msg.Metadata) > 0 ||
		msg.Path != "" ||
		len(msg.Payload) > 0 ||
		msg.ServiceName != "" ||
		msg.URL != "" ||
		msg.SessionID != "" ||
		msg.QualityOfService != 0 ||
		msg.Payload != nil {
		return errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
	}

	a.Status = *msg.Status
	return nil
}

// To converts the Authorization struct to a Message struct.  The Message struct
// is validated before being returned.  If the Message struct is invalid, an
// error is returned.
func (a *Authorization) To() (*Message, error) {
	msg := Message{
		Type:   AuthorizationMessageType,
		Status: &a.Status,
	}

	if err := msg.validate(); err != nil {
		return nil, err
	}

	return &msg, nil
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

// -----------------------------------------------------------------------------

func validateUTF8(msg *Message) error {
	var err error
	strUTF8Vador(msg.Source, "Source", &err)
	strUTF8Vador(msg.Destination, "Destination", &err)
	strUTF8Vador(msg.TransactionUUID, "TransactionUUID", &err)
	strUTF8Vador(msg.ContentType, "ContentType", &err)
	strUTF8Vador(msg.Accept, "Accept", &err)
	sArrayUTF8Vador(msg.Headers, "Headers", &err)
	mapUTF8Vador(msg.Metadata, "Metadata", &err)
	strUTF8Vador(msg.Path, "Path", &err)
	strUTF8Vador(msg.ServiceName, "ServiceName", &err)
	strUTF8Vador(msg.URL, "URL", &err)
	sArrayUTF8Vador(msg.PartnerIDs, "PartnerIDs", &err)
	strUTF8Vador(msg.SessionID, "SessionID", &err)

	return err
}

func strUTF8Vador(s, field string, err *error) {
	if *err != nil {
		return
	}

	if !utf8.ValidString(s) {
		*err = errors.Join(ErrMessageIsInvalid, ErrNotUTF8, errors.New("invalid "+field))
	}
}

func sArrayUTF8Vador(list []string, field string, err *error) {
	if *err != nil {
		return
	}

	for _, s := range list {
		if !utf8.ValidString(s) {
			*err = errors.Join(ErrMessageIsInvalid, ErrNotUTF8, errors.New("invalid "+field))
			return
		}
	}
}

func mapUTF8Vador(m map[string]string, field string, err *error) {
	if *err != nil {
		return
	}

	for k, v := range m {
		if !utf8.ValidString(k) || !utf8.ValidString(v) {
			*err = errors.Join(ErrMessageIsInvalid, ErrNotUTF8, errors.New("invalid "+field))
			return
		}
	}
}
