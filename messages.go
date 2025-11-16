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
	ErrUnsupportedFieldsSet = errors.New("unsupported fields set")
	ErrNotUTF8              = errors.New("field contains non-utf-8 characters")
)

// Message is the union of all WRP fields, made optional (except for Type).  This
// type is useful for general purpose code that needs to transcode messages
// without knowing the exact type of message.  It is also the type used by the
// Observer, Processor and Modifier interfaces.
//
// IMPORTANT: Anytime a new WRP field is added to any message, or a new message
// with new fields, those new fields must be added to this struct for transcoding
// to work properly.  There are tests in the wrp package that ensure this struct
// is kept up-to-date.
//
// For code that sends specific messages, use one of the other WRP structs in
// this package because they will help you produce correct messages.
type Message struct {
	// Type is the message type for the message.
	//
	// example: SimpleRequestResponseMessageType
	Type MessageType `json:"msg_type"`

	// Source is the device_id name of the device originating the request or
	// response.
	//
	// example: dns:talaria.xmidt.example.com
	Source string `json:"source,omitempty"`

	// Destination is the device_id name of the target device of the request or
	// response.
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

	// RequestDeliveryResponse is the request delivery response is the delivery
	// result of the previous (implied request) message with a matching
	// transaction_uuid.
	RequestDeliveryResponse *int64 `json:"rdr,omitempty"`

	// Headers is the headers associated with the payload.
	Headers []string `json:"headers,omitempty"`

	// Metadata is the map of name/value pairs used by consumers of WRP messages
	// for filtering & other purposes.
	//
	// example: {
	//   "/boot-time":"1542834188",
	//   "/last-reconnect-reason":"spanish inquisition",
	// }
	Metadata map[string]string `json:"metadata,omitempty"`

	// Path is the path to which to apply the payload.
	Path string `json:"path,omitempty"`

	// Payload is the payload for this message.  It's format is expected to match
	// ContentType.
	//
	// For JSON, this field must be a UTF-8 string.  Binary payloads may be
	// base64-encoded.
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

	// QualityOfService is the qos value associated with this message.  Values
	// between 0 and 99, inclusive, are defined by the wrp spec.  Negative values
	// are assumed to be zero, and values larger than 99 are assumed to be 99.
	QualityOfService QOSValue `json:"qos,omitempty"`
}

// SetStatus simplifies setting the optional Status field.
func (msg *Message) SetStatus(value int64) *Message {
	msg.Status = &value
	return msg
}

// SetRequestDeliveryResponse simplifies setting the optional
// RequestDeliveryResponse field.
func (msg *Message) SetRequestDeliveryResponse(value int64) *Message {
	msg.RequestDeliveryResponse = &value
	return msg
}

// MsgType returns the message type for the struct.
func (msg *Message) MsgType() MessageType {
	return msg.Type
}

// From converts a Message struct to the struct.  The Message struct is
// validated before being converted.  If the Message struct is invalid, an error
// is returned.  If all of the Processors return ErrNotHandled, the resulting
// message is considered valid, and no error is returned.  Otherwise the first
// error encountered is returned, or nil is returned.  If NoStandardValidation()
// is provided, the message is not validated against the standard validators,
// but any user-provided validators are still run.
func (msg *Message) From(m *Message, validators ...Processor) error {
	err := validate(m, validators...)
	if err == nil {
		msg.from(m)
	}

	return err
}

// To converts the struct into a Message struct.  The Message struct is validated
// before being returned.  If the Message struct is invalid, an error is returned.
// If all of the Processors return ErrNotHandled, the resulting message is
// considered valid, and no error is returned.  Otherwise the first error
// encountered is returned, or nil is returned.  If NoStandardValidation() is
// provided, the message is not validated against the standard validators, but
// any user-provided validators are still run.
func (msg *Message) To(m *Message, validators ...Processor) error {
	err := validate(msg, validators...)
	if err == nil {
		m.from(msg)
	}

	return err
}

// Validate checks the struct for correctness.  The check is performed on the
// *Message form of the struct using the provided validators.  If the struct is
// invalid, an error is returned.  If all of the Processors return ErrNotHandled,
// the resulting message is considered valid, and no error is returned.  Otherwise
// the first error encountered is returned, or nil is returned.  If
// NoStandardValidation() is provided, the message is not validated against the
// standard validators, but any user-provided validators are still run.
func (msg *Message) Validate(validators ...Processor) error {
	return validate(msg, validators...)
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

// EncodeMsgpack encodes the message to msgpack format with zero allocations
// when the provided buffer has sufficient capacity. This is the fastest encoding
// path, approximately 12x faster than using NewEncoder() with zero allocations.
//
// IMPORTANT: This method does NOT perform validation. Use this only when:
//   - Encoding trusted internal messages
//   - Performance is critical (hot path)
//   - Messages have already been validated upstream
//
// The buffer is reused and may be grown if needed. Returns the encoded data,
// which may be a different slice if the buffer was grown (1 allocation).
//
// For validated encoding, use NewEncoder() instead.
//
// Performance characteristics:
//   - 0 allocations if buffer has sufficient capacity (typical case)
//   - 1 allocation if buffer needs to grow (rare with properly sized buffer)
//   - ~285 ns/op (vs ~3,500 ns/op for validated encoding)
//
// Example usage:
//
//	buf := make([]byte, 0, 1024)  // Pre-allocate buffer
//	for _, msg := range messages {
//	    buf = buf[:0]  // Reuse buffer without deallocating
//	    encoded, err := msg.EncodeMsgpack(buf)
//	    if err != nil {
//	        return err
//	    }
//	    // Use encoded...
//	}
func (msg *Message) EncodeMsgpack(buf []byte) ([]byte, error) {
	return msg.marshalMsg(buf)
}

// DecodeMsgpack decodes msgpack data into the message with minimal allocations.
// This is significantly faster than using NewDecoder(), approximately 6x faster
// with fewer allocations.
//
// IMPORTANT: This method does NOT perform validation. Use this only when:
//   - Decoding trusted internal messages
//   - Performance is critical (hot path)
//   - Messages will be validated downstream
//
// Returns the remaining bytes after decoding, which is useful for stream
// processing or handling multiple concatenated messages.
//
// For validated decoding, use NewDecoder() instead.
//
// Performance characteristics:
//   - 13 allocations (for message fields, unavoidable during unmarshaling)
//   - ~572 ns/op (vs ~4,351 ns/op for validated decoding)
//
// Example usage:
//
//	var msg wrp.Message
//	remaining, err := msg.DecodeMsgpack(data)
//	if err != nil {
//	    return err
//	}
//	// Process msg...
//	// Process remaining if needed...
func (msg *Message) DecodeMsgpack(data []byte) ([]byte, error) {
	return msg.unmarshalMsg(data)
}

// -----------------------------------------------------------------------------

// Union is an interface that all WRP message types implement.  This interface
// is used by the Is() and As() functions to determine the message type, to
// convert between message types, and for encoding and decoding messages.
//
// This interface is designed so consumers of the WRP library can compose their
// own structs that implement this interface, and then use the Is() and As() to
// determine the message type and convert between message types.
type Union interface {
	// MsgType returns the message type for the struct that implements
	// this interface.
	MsgType() MessageType

	// From converts a Message struct to the struct that implements this
	// interface.  The Message struct is validated before being converted.  If
	// the Message struct is invalid, an error is returned.  If all of the
	// Processors return ErrNotHandled, the resulting message is considered
	// valid, and no error is returned.  Otherwise the first error encountered
	// is returned, or nil is returned.  If NoStandardValidation() is provided,
	// the message is not validated against the standard validators, but any
	// user-provided validators are still run.
	From(*Message, ...Processor) error

	// To converts the struct that implements this interface to a Message
	// struct.  The Message struct is validated before being returned.  If the
	// Message struct is invalid, an error is returned.  If all of the
	// Processors return ErrNotHandled, the resulting message is considered
	// valid, and no error is returned.  Otherwise the first error encountered
	// is returned, or nil is returned.  If NoStandardValidation() is provided,
	// the message is not validated against the standard validators, but any
	// user-provided validators are still run.
	To(*Message, ...Processor) error

	// Validate checks the struct that implements this interface for correctness.
	// The check is performed on the *Message form of the struct using the
	// provided validators.  If the struct is invalid, an error is returned.  If
	// all of the Processors return ErrNotHandled, the resulting message is
	// considered valid, and no error is returned.  Otherwise the first error
	// encountered is returned, or nil is returned.  If NoStandardValidation() is
	// provided, the message is not validated against the standard validators, but
	// any user-provided validators are still run.
	Validate(...Processor) error
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

var _ Union = (*SimpleRequestResponse)(nil)

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

// MsgType returns the message type for the struct.
func (srr *SimpleRequestResponse) MsgType() MessageType {
	return SimpleRequestResponseMessageType
}

// From converts a Message struct to the struct.  The Message struct is
// validated before being converted.  If the Message struct is invalid, an error
// is returned.  If all of the Processors return ErrNotHandled, the resulting
// message is considered valid, and no error is returned.  Otherwise the first
// error encountered is returned, or nil is returned.  If NoStandardValidation()
// is provided, the message is not validated against the standard validators,
// but any user-provided validators are still run.
func (srr *SimpleRequestResponse) From(msg *Message, validators ...Processor) error {
	if msg.Type != SimpleRequestResponseMessageType {
		return ErrInvalidMessageType
	}

	if err := validate(msg, validators...); err != nil {
		return err
	}

	srr.from(msg)
	return nil
}

// To converts the struct into a Message struct.  The Message struct is validated
// before being returned.  If the Message struct is invalid, an error is returned.
// If all of the Processors return ErrNotHandled, the resulting message is
// considered valid, and no error is returned.  Otherwise the first error
// encountered is returned, or nil is returned.  If NoStandardValidation() is
// provided, the message is not validated against the standard validators, but
// any user-provided validators are still run.
func (srr *SimpleRequestResponse) To(msg *Message, validators ...Processor) error {
	var tmp Message
	srr.to(&tmp)
	if err := validate(&tmp, validators...); err != nil {
		return err
	}

	srr.to(msg)
	return nil
}

// Validate checks the struct for correctness.  The check is performed on the
// *Message form of the struct using the provided validators.  If the struct is
// invalid, an error is returned.  If all of the Processors return ErrNotHandled,
// the resulting message is considered valid, and no error is returned.  Otherwise
// the first error encountered is returned, or nil is returned.  If
// NoStandardValidation() is provided, the message is not validated against the
// standard validators, but any user-provided validators are still run.
func (srr *SimpleRequestResponse) Validate(validators ...Processor) error {
	var msg Message
	srr.to(&msg)
	return validate(&msg, validators...)
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

func (srr *SimpleRequestResponse) to(msg *Message) {
	msg.Type = SimpleRequestResponseMessageType
	msg.Source = srr.Source
	msg.Destination = srr.Destination
	msg.TransactionUUID = srr.TransactionUUID
	msg.ContentType = srr.ContentType
	msg.Accept = srr.Accept
	msg.Status = srr.Status
	msg.RequestDeliveryResponse = srr.RequestDeliveryResponse
	msg.PartnerIDs = trimPartnerIDs(srr.PartnerIDs)
	msg.Headers = srr.Headers
	msg.Metadata = srr.Metadata
	msg.QualityOfService = srr.QualityOfService
	msg.SessionID = srr.SessionID
	msg.Payload = srr.Payload
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

var _ Union = (*SimpleEvent)(nil)

// SetRequestDeliveryResponse simplifies setting the optional
// RequestDeliveryResponse field.
func (se *SimpleEvent) SetRequestDeliveryResponse(value int64) *SimpleEvent {
	se.RequestDeliveryResponse = &value
	return se
}

// MsgType returns the message type for the struct.
func (se *SimpleEvent) MsgType() MessageType {
	return SimpleEventMessageType
}

// From converts a Message struct to the struct.  The Message struct is
// validated before being converted.  If the Message struct is invalid, an error
// is returned.  If all of the Processors return ErrNotHandled, the resulting
// message is considered valid, and no error is returned.  Otherwise the first
// error encountered is returned, or nil is returned.  If NoStandardValidation()
// is provided, the message is not validated against the standard validators,
// but any user-provided validators are still run.
func (se *SimpleEvent) From(msg *Message, validators ...Processor) error {
	if msg.Type != SimpleEventMessageType {
		return ErrInvalidMessageType
	}

	if err := validate(msg, validators...); err != nil {
		return err
	}

	se.from(msg)
	return nil
}

// To converts the struct into a Message struct.  The Message struct is validated
// before being returned.  If the Message struct is invalid, an error is returned.
// If all of the Processors return ErrNotHandled, the resulting message is
// considered valid, and no error is returned.  Otherwise the first error
// encountered is returned, or nil is returned.  If NoStandardValidation() is
// provided, the message is not validated against the standard validators, but
// any user-provided validators are still run.
func (se *SimpleEvent) To(msg *Message, validators ...Processor) error {
	var tmp Message
	se.to(&tmp)
	if err := validate(&tmp, validators...); err != nil {
		return err
	}

	se.to(msg)
	return nil
}

// Validate checks the struct for correctness.  The check is performed on the
// *Message form of the struct using the provided validators.  If the struct is
// invalid, an error is returned.  If all of the Processors return ErrNotHandled,
// the resulting message is considered valid, and no error is returned.  Otherwise
// the first error encountered is returned, or nil is returned.  If
// NoStandardValidation() is provided, the message is not validated against the
// standard validators, but any user-provided validators are still run.
func (se *SimpleEvent) Validate(validators ...Processor) error {
	var msg Message
	se.to(&msg)
	return validate(&msg, validators...)
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

func (se *SimpleEvent) to(msg *Message) {
	msg.Type = SimpleEventMessageType
	msg.Source = se.Source
	msg.Destination = se.Destination
	msg.TransactionUUID = se.TransactionUUID
	msg.ContentType = se.ContentType
	msg.RequestDeliveryResponse = se.RequestDeliveryResponse
	msg.PartnerIDs = trimPartnerIDs(se.PartnerIDs)
	msg.Headers = se.Headers
	msg.Metadata = se.Metadata
	msg.SessionID = se.SessionID
	msg.QualityOfService = se.QualityOfService
	msg.Payload = se.Payload
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

var _ Union = (*CRUD)(nil)

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

// MsgType returns the message type for the struct.
func (c *CRUD) MsgType() MessageType {
	return c.Type
}

// From converts a Message struct to the struct.  The Message struct is
// validated before being converted.  If the Message struct is invalid, an error
// is returned.  If all of the Processors return ErrNotHandled, the resulting
// message is considered valid, and no error is returned.  Otherwise the first
// error encountered is returned, or nil is returned.  If NoStandardValidation()
// is provided, the message is not validated against the standard validators,
// but any user-provided validators are still run.
func (c *CRUD) From(msg *Message, validators ...Processor) error {
	if msg.Type != c.Type {
		return ErrInvalidMessageType
	}

	if err := validate(msg, validators...); err != nil {
		return err
	}

	c.from(msg)
	return nil
}

// To converts the struct into a Message struct.  The Message struct is validated
// before being returned.  If the Message struct is invalid, an error is returned.
// If all of the Processors return ErrNotHandled, the resulting message is
// considered valid, and no error is returned.  Otherwise the first error
// encountered is returned, or nil is returned.  If NoStandardValidation() is
// provided, the message is not validated against the standard validators, but
// any user-provided validators are still run.
func (c *CRUD) To(msg *Message, validators ...Processor) error {
	var tmp Message
	c.to(&tmp)
	if err := validate(&tmp, validators...); err != nil {
		return err
	}

	c.to(msg)
	return nil
}

// Validate checks the struct for correctness.  The check is performed on the
// *Message form of the struct using the provided validators.  If the struct is
// invalid, an error is returned.  If all of the Processors return ErrNotHandled,
// the resulting message is considered valid, and no error is returned.  Otherwise
// the first error encountered is returned, or nil is returned.  If
// NoStandardValidation() is provided, the message is not validated against the
// standard validators, but any user-provided validators are still run.
func (c *CRUD) Validate(validators ...Processor) error {
	var msg Message
	c.to(&msg)
	return validate(&msg, validators...)
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

func (c *CRUD) to(msg *Message) {
	msg.Type = c.Type
	msg.Source = c.Source
	msg.Destination = c.Destination
	msg.TransactionUUID = c.TransactionUUID
	msg.ContentType = c.ContentType
	msg.Accept = c.Accept
	msg.Status = c.Status
	msg.Path = c.Path
	msg.RequestDeliveryResponse = c.RequestDeliveryResponse
	msg.PartnerIDs = trimPartnerIDs(c.PartnerIDs)
	msg.Headers = c.Headers
	msg.Metadata = c.Metadata
	msg.QualityOfService = c.QualityOfService
	msg.SessionID = c.SessionID
	msg.Payload = c.Payload
}

// -----------------------------------------------------------------------------

// ServiceRegistration represents a WRP message of type ServiceRegistrationMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#on-device-service-registration-message-definition
type ServiceRegistration struct {
	ServiceName string `required:""`
	URL         string `required:""`
}

var _ Union = (*ServiceRegistration)(nil)

// MsgType returns the message type for the struct.
func (sr *ServiceRegistration) MsgType() MessageType {
	return ServiceRegistrationMessageType
}

// From converts a Message struct to the struct.  The Message struct is
// validated before being converted.  If the Message struct is invalid, an error
// is returned.  If all of the Processors return ErrNotHandled, the resulting
// message is considered valid, and no error is returned.  Otherwise the first
// error encountered is returned, or nil is returned.  If NoStandardValidation()
// is provided, the message is not validated against the standard validators,
// but any user-provided validators are still run.
func (sr *ServiceRegistration) From(msg *Message, validators ...Processor) error {
	if msg.Type != ServiceRegistrationMessageType {
		return ErrInvalidMessageType
	}

	if err := validate(msg, validators...); err != nil {
		return err
	}

	sr.from(msg)
	return nil
}

// To converts the struct into a Message struct.  The Message struct is validated
// before being returned.  If the Message struct is invalid, an error is returned.
// If all of the Processors return ErrNotHandled, the resulting message is
// considered valid, and no error is returned.  Otherwise the first error
// encountered is returned, or nil is returned.  If NoStandardValidation() is
// provided, the message is not validated against the standard validators, but
// any user-provided validators are still run.
func (sr *ServiceRegistration) To(msg *Message, validators ...Processor) error {
	var tmp Message
	sr.to(&tmp)
	if err := validate(&tmp, validators...); err != nil {
		return err
	}
	sr.to(msg)
	return nil
}

// Validate checks the struct for correctness.  The check is performed on the
// *Message form of the struct using the provided validators.  If the struct is
// invalid, an error is returned.  If all of the Processors return ErrNotHandled,
// the resulting message is considered valid, and no error is returned.  Otherwise
// the first error encountered is returned, or nil is returned.  If
// NoStandardValidation() is provided, the message is not validated against the
// standard validators, but any user-provided validators are still run.
func (sr *ServiceRegistration) Validate(validators ...Processor) error {
	var msg Message
	sr.to(&msg)
	return validate(&msg, validators...)
}

func (sr *ServiceRegistration) from(msg *Message) {
	sr.ServiceName = msg.ServiceName
	sr.URL = msg.URL
}

func (sr *ServiceRegistration) to(msg *Message) {
	msg.Type = ServiceRegistrationMessageType
	msg.ServiceName = sr.ServiceName
	msg.URL = sr.URL
}

// -----------------------------------------------------------------------------

// ServiceAlive represents a WRP message of type ServiceAliveMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#on-device-service-alive-message-definition
type ServiceAlive struct{}

var _ Union = (*ServiceAlive)(nil)

// MsgType returns the message type for the struct.
func (sa *ServiceAlive) MsgType() MessageType {
	return ServiceAliveMessageType
}

// From converts a Message struct to the struct.  The Message struct is
// validated before being converted.  If the Message struct is invalid, an error
// is returned.  If all of the Processors return ErrNotHandled, the resulting
// message is considered valid, and no error is returned.  Otherwise the first
// error encountered is returned, or nil is returned.  If NoStandardValidation()
// is provided, the message is not validated against the standard validators,
// but any user-provided validators are still run.
func (sa *ServiceAlive) From(msg *Message, validators ...Processor) error {
	if msg.Type != ServiceAliveMessageType {
		return ErrInvalidMessageType
	}

	if err := validate(msg, validators...); err != nil {
		return err
	}
	return nil
}

// To converts the struct into a Message struct.  The Message struct is validated
// before being returned.  If the Message struct is invalid, an error is returned.
// If all of the Processors return ErrNotHandled, the resulting message is
// considered valid, and no error is returned.  Otherwise the first error
// encountered is returned, or nil is returned.  If NoStandardValidation() is
// provided, the message is not validated against the standard validators, but
// any user-provided validators are still run.
func (sa *ServiceAlive) To(msg *Message, validators ...Processor) error {
	var tmp Message
	sa.to(&tmp)
	if err := validate(&tmp, validators...); err != nil {
		return err
	}

	sa.to(msg)
	return nil
}

// Validate checks the struct for correctness.  The check is performed on the
// *Message form of the struct using the provided validators.  If the struct is
// invalid, an error is returned.  If all of the Processors return ErrNotHandled,
// the resulting message is considered valid, and no error is returned.  Otherwise
// the first error encountered is returned, or nil is returned.  If
// NoStandardValidation() is provided, the message is not validated against the
// standard validators, but any user-provided validators are still run.
func (sa *ServiceAlive) Validate(validators ...Processor) error {
	var msg Message
	sa.to(&msg)
	return validate(&msg, validators...)
}

func (sa *ServiceAlive) to(msg *Message) {
	msg.Type = ServiceAliveMessageType
}

// -----------------------------------------------------------------------------

// Unknown represents a WRP message of type UnknownMessageType.
//
// https://github.com/xmidt-org/wrp-c/wiki/Web-Routing-Protocol#unknown-message-definition
type Unknown struct{}

var _ Union = (*Unknown)(nil)

// MsgType returns the message type for the struct.
func (u *Unknown) MsgType() MessageType {
	return UnknownMessageType
}

// From converts a Message struct to the struct.  The Message struct is
// validated before being converted.  If the Message struct is invalid, an error
// is returned.  If all of the Processors return ErrNotHandled, the resulting
// message is considered valid, and no error is returned.  Otherwise the first
// error encountered is returned, or nil is returned.  If NoStandardValidation()
// is provided, the message is not validated against the standard validators,
// but any user-provided validators are still run.
func (u *Unknown) From(msg *Message, validators ...Processor) error {
	if msg.Type != UnknownMessageType {
		return ErrInvalidMessageType
	}

	if err := validate(msg, validators...); err != nil {
		return err
	}

	return nil
}

// To converts the struct into a Message struct.  The Message struct is validated
// before being returned.  If the Message struct is invalid, an error is returned.
// If all of the Processors return ErrNotHandled, the resulting message is
// considered valid, and no error is returned.  Otherwise the first error
// encountered is returned, or nil is returned.  If NoStandardValidation() is
// provided, the message is not validated against the standard validators, but
// any user-provided validators are still run.
func (u *Unknown) To(msg *Message, validators ...Processor) error {
	var tmp Message
	u.to(&tmp)
	if err := validate(&tmp, validators...); err != nil {
		return err
	}

	u.to(msg)
	return nil
}

// Validate checks the struct for correctness.  The check is performed on the
// *Message form of the struct using the provided validators.  If the struct is
// invalid, an error is returned.  If all of the Processors return ErrNotHandled,
// the resulting message is considered valid, and no error is returned.  Otherwise
// the first error encountered is returned, or nil is returned.  If
// NoStandardValidation() is provided, the message is not validated against the
// standard validators, but any user-provided validators are still run.
func (u *Unknown) Validate(validators ...Processor) error {
	var msg Message
	u.to(&msg)
	return validate(&msg, validators...)
}

func (u *Unknown) to(msg *Message) {
	msg.Type = UnknownMessageType
}

// -----------------------------------------------------------------------------

// Authorization is a message type that represents an authorization message.
type Authorization struct {
	Status int64 `required:""`
}

var _ Union = (*Authorization)(nil)

// MsgType returns the message type for the struct.
func (a *Authorization) MsgType() MessageType {
	return AuthorizationMessageType
}

// From converts a Message struct to the struct.  The Message struct is
// validated before being converted.  If the Message struct is invalid, an error
// is returned.  If all of the Processors return ErrNotHandled, the resulting
// message is considered valid, and no error is returned.  Otherwise the first
// error encountered is returned, or nil is returned.  If NoStandardValidation()
// is provided, the message is not validated against the standard validators,
// but any user-provided validators are still run.
func (a *Authorization) From(msg *Message, validators ...Processor) error {
	if msg.Type != AuthorizationMessageType {
		return ErrInvalidMessageType
	}

	if err := validate(msg, validators...); err != nil {
		return err
	}

	a.from(msg)
	return nil
}

// To converts the struct into a Message struct.  The Message struct is validated
// before being returned.  If the Message struct is invalid, an error is returned.
// If all of the Processors return ErrNotHandled, the resulting message is
// considered valid, and no error is returned.  Otherwise the first error
// encountered is returned, or nil is returned.  If NoStandardValidation() is
// provided, the message is not validated against the standard validators, but
// any user-provided validators are still run.
func (a *Authorization) To(msg *Message, validators ...Processor) error {
	var tmp Message
	a.to(&tmp)
	if err := validate(&tmp, validators...); err != nil {
		return err
	}

	a.to(msg)
	return nil
}

// Validate checks the struct for correctness.  The check is performed on the
// *Message form of the struct using the provided validators.  If the struct is
// invalid, an error is returned.  If all of the Processors return ErrNotHandled,
// the resulting message is considered valid, and no error is returned.  Otherwise
// the first error encountered is returned, or nil is returned.  If
// NoStandardValidation() is provided, the message is not validated against the
// standard validators, but any user-provided validators are still run.
func (a *Authorization) Validate(validators ...Processor) error {
	var msg Message
	a.to(&msg)
	return validate(&msg, validators...)
}

func (a *Authorization) from(msg *Message) {
	a.Status = *msg.Status
}

func (a *Authorization) to(msg *Message) {
	msg.Type = AuthorizationMessageType
	msg.Status = &a.Status
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
