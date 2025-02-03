// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	ErrNotUTF8          = errors.New("field contains non-utf-8 characters")
	ErrValidationFailed = errors.New("validation failed")
	/*
		ErrInvalidMessageType = errors.New("invalid message type")
		ErrSourceInvalid      = errors.New("source is invalid")
		ErrDestinationInvalid = errors.New("destination is invalid")
		ErrNoTransactionUUID  = errors.New("no transaction UUID")
		ErrInvalidContentType = errors.New("invalid content type")
		ErrInvalidPartner     = errors.New("invalid partner ID")
		ErrNoServiceName      = errors.New("no service name")
		ErrNoURL              = errors.New("no URL")
	*/
)

// ValidateUTF8 provides a processor that validates that all fields in a message
// are valid UTF-8, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateUTF8() Processor {
	return Processors{
		ValidateUTF8Source(),
		ValidateUTF8Destination(),
		ValidateUTF8TransactionUUID(),
		ValidateUTF8ContentType(),
		ValidateUTF8Accept(),
		ValidateUTF8Headers(),
		ValidateUTF8Metadata(),
		ValidateUTF8Path(),
		ValidateUTF8ServiceName(),
		ValidateUTF8URL(),
		ValidateUTF8Partners(),
		ValidateUTF8SessionID(),
	}
}

func strUTF8Vador(s, field string) error {
	if !utf8.ValidString(s) {
		return errors.Join(ErrValidationFailed, ErrNotUTF8, errors.New("invalid "+field))
	}

	return ErrNotHandled
}

func sArrayUTF8Vador(list []string, field string) error {
	for _, s := range list {
		if !utf8.ValidString(s) {
			return errors.Join(ErrValidationFailed, ErrNotUTF8, errors.New("invalid "+field))
		}
	}

	return ErrNotHandled
}

// ValidateUTF8Source provides a processor that validates that the source field
// is valid UTF-8, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateUTF8Source() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		return strUTF8Vador(msg.Source, "source")
	})
}

// ValidateUTF8Destination provides a processor that validates that the
// destination field is valid UTF-8, or returns an error.  If the message is
// valid, the processor returns ErrNotHandled.
func ValidateUTF8Destination() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		return strUTF8Vador(msg.Destination, "destination")
	})
}

// ValidateUTF8TransactionUUID provides a processor that validates that the
// transaction UUID field is valid UTF-8, or returns an error.  If the message
// is valid, the processor returns ErrNotHandled.
func ValidateUTF8TransactionUUID() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		return strUTF8Vador(msg.TransactionUUID, "transaction_uuid")
	})
}

// ValidateUTF8ContentType provides a processor that validates that the content
// type field is valid UTF-8, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateUTF8ContentType() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		return strUTF8Vador(msg.ContentType, "content_type")
	})
}

// ValidateUTF8Accept provides a processor that validates that the accept field
// is valid UTF-8, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateUTF8Accept() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		return strUTF8Vador(msg.Accept, "accept")
	})
}

// ValidateUTF8Headers provides a processor that validates that all headers are
// valid UTF-8, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateUTF8Headers() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		return sArrayUTF8Vador(msg.Headers, "headers")
	})
}

// ValidateUTF8Metadata provides a processor that validates that all metadata
// keys and values are valid UTF-8, or returns an error.  If the message is
// valid, the processor returns ErrNotHandled.
func ValidateUTF8Metadata() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		for k, v := range msg.Metadata {
			if !utf8.ValidString(k) || !utf8.ValidString(v) {
				return errors.Join(ErrNotUTF8, errors.New("invalid metadata"))
			}
		}

		return ErrNotHandled
	})
}

// ValidateUTF8Path provides a processor that validates that the path field is
// valid UTF-8, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateUTF8Path() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		return strUTF8Vador(msg.Path, "path")
	})
}

// ValidateUTF8ServiceName provides a processor that validates that the service
// name field is valid UTF-8, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateUTF8ServiceName() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		return strUTF8Vador(msg.ServiceName, "service_name")
	})
}

// ValidateUTF8URL provides a processor that validates that the URL field is
// valid UTF-8, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateUTF8URL() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		return strUTF8Vador(msg.URL, "url")
	})
}

// ValidateUTF8Partners provides a processor that validates that all partner
// IDs are valid UTF-8, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateUTF8Partners() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		return sArrayUTF8Vador(msg.PartnerIDs, "partner_ids")
	})
}

// ValidateUTF8SessionID provides a processor that validates that the session ID
// field is valid UTF-8, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateUTF8SessionID() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		return strUTF8Vador(msg.SessionID, "session_id")
	})
}

//------------------------------------------------------------------------------

// ValidateMessageType provides a processor that validates that the message type
// is valid, or returns an error.  If the message is valid, the processor returns
// ErrNotHandled.
func ValidateMessageType() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if !msg.Type.IsValid() {
			return errors.Join(
				ErrValidationFailed,
				errors.New("invalid message type"),
			)
		}

		return ErrNotHandled
	})
}

// ValidateMessageTypeIs provides a processor that validates that the message
// type is the provided type, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateMessageTypeIs(typ MessageType) Processor {
	return ValidateMessageTypeIsOneOf(typ)
}

// ValidateMessageTypeIsOneOf provides a processor that validates that the message
// type is one of the provided types, or returns an error.  If the message is
// valid, the processor returns ErrNotHandled.
func ValidateMessageTypeIsOneOf(types ...MessageType) Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		for _, typ := range types {
			if msg.Type == typ {
				return ErrNotHandled
			}
		}

		return errors.Join(
			ErrValidationFailed,
			errors.New("invalid message type"),
		)
	})
}

//------------------------------------------------------------------------------

// ValidateSource provides a processor that validates that the source locator is
// valid, or returns an error.  If the message is valid, the processor returns
// ErrNotHandled.
func ValidateSource() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.Source == "" {
			return errors.Join(ErrValidationFailed, errors.New("missing source"))
		}

		if _, err := ParseLocator(msg.Source); err != nil {
			return errors.Join(ErrValidationFailed, errors.New("invalid source"), err)
		}
		return ErrNotHandled
	})
}

// ValidateNoSource provides a processor that validates that the message has no
// source locator, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateNoSource() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.Source == "" {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected source"))
	})
}

//------------------------------------------------------------------------------

// ValidateDestination provides a processor that validates that the destination
// locator is valid, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateDestination() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.Destination == "" {
			return errors.Join(ErrValidationFailed, errors.New("missing destination"))
		}

		if _, err := ParseLocator(msg.Destination); err != nil {
			return errors.Join(ErrValidationFailed, errors.New("invalid destination"), err)
		}
		return ErrNotHandled
	})
}

// ValidateNoDestination provides a processor that validates that the message has
// no destination locator, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateNoDestination() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.Destination == "" {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected destination"))
	})
}

//------------------------------------------------------------------------------

// ValidateTransactionUUID provides a processor that validates that the message
// has a transaction UUID, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateTransactionUUID() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.TransactionUUID == "" {
			return errors.Join(ErrValidationFailed, errors.New("missing transaction UUID"))
		}
		return ErrNotHandled
	})
}

// ValidateNoTransactionUUID provides a processor that validates that the message
// has no transaction UUID, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateNoTransactionUUID() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.TransactionUUID == "" {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected transaction UUID"))
	})
}

//------------------------------------------------------------------------------

// ValidateNoContentType provides a processor that validates that the message has
// no content type, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateNoContentType() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.ContentType == "" {
			return ErrNotHandled
		}
		return errors.Join(ErrValidationFailed, errors.New("unexpected content type"))
	})
}

//------------------------------------------------------------------------------

// ValidateNoAccept provides a processor that validates that the message has no
// accept field, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateNoAccept() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.Accept == "" {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected accept"))
	})
}

//------------------------------------------------------------------------------

// ValidateStatus provides a processor that validates that the message has a
// status, or returns an error.  If the message is valid, the processor returns
// ErrNotHandled.
func ValidateStatus() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.Status == nil {
			return errors.Join(ErrValidationFailed, errors.New("missing status"))
		}

		return ErrNotHandled
	})
}

// ValidateNoStatus provides a processor that validates that the message has no
// status, or returns an error.  If the message is valid, the processor returns
// ErrNotHandled.
func ValidateNoStatus() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.Status == nil {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected status"))
	})
}

//------------------------------------------------------------------------------

// ValidateRequestDeliveryResponse provides a processor that validates that the
// message has a request delivery response, or returns an error.  If the message
// is valid, the processor returns ErrNotHandled.
func ValidateRequestDeliveryResponse() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.RequestDeliveryResponse == nil {
			return errors.Join(ErrValidationFailed, errors.New("missing request delivery response"))
		}

		return ErrNotHandled
	})
}

// ValidateNoRequestDeliveryResponse provides a processor that validates that the
// message has no request delivery response, or returns an error.  If the message
// is valid, the processor returns ErrNotHandled.
func ValidateNoRequestDeliveryResponse() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.RequestDeliveryResponse == nil {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected request delivery response"))
	})
}

//------------------------------------------------------------------------------

// ValidateNoHeaders provides a processor that validates that the message has no
// headers, or returns an error.  If the message is valid, the processor returns
// ErrNotHandled.
func ValidateNoHeaders() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if len(msg.Headers) == 0 {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected headers"))
	})
}

//------------------------------------------------------------------------------

// ValidateNoMetadata provides a processor that validates that the message has no
// metadata, or returns an error.  If the message is valid, the processor returns
// ErrNotHandled.
func ValidateNoMetadata() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if len(msg.Metadata) == 0 {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected metadata"))
	})
}

//------------------------------------------------------------------------------

// ValidateNoPath provides a processor that validates that the message has no path,
// or returns an error.  If the message is valid, the processor returns ErrNotHandled.
func ValidateNoPath() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.Path == "" {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected path"))
	})
}

//------------------------------------------------------------------------------

// ValidateNoPayload provides a processor that validates that the message has no
// payload, or returns an error.  If the message is valid, the processor returns
// ErrNotHandled.
func ValidateNoPayload() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if len(msg.Payload) == 0 {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected payload"))
	})
}

//------------------------------------------------------------------------------

// ValidateServiceName provides a processor that validates that the message has a
// service name, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateServiceName() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.ServiceName == "" {
			return errors.Join(ErrValidationFailed, errors.New("missing service name"))
		}

		return ErrNotHandled
	})
}

// ValidateNoServiceName provides a processor that validates that the message
// has no service name, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateNoServiceName() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.ServiceName == "" {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected service name"))
	})
}

//------------------------------------------------------------------------------

// ValidateURL provides a processor that validates that the message has a URL, or
// returns an error.  If the message is valid, the processor returns ErrNotHandled.
func ValidateURL() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.URL == "" {
			return errors.Join(ErrValidationFailed, errors.New("missing url"))
		}

		return ErrNotHandled
	})
}

// ValidateNoURL provides a processor that validates that the message has a URL, or
// returns an error.  If the message is valid, the processor returns ErrNotHandled.
func ValidateNoURL() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.URL == "" {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected url"))
	})
}

//------------------------------------------------------------------------------

// ValidatePartnerPresent provides a processor that validates that a message has
// at least one partner ID, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidatePartnerPresent() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if 0 < len(msg.PartnerIDs) {
			// Ensure that all partner IDs are not empty
			for _, id := range msg.PartnerIDs {
				if id != "" {
					return ErrNotHandled
				}
			}
		}

		return errors.Join(
			ErrValidationFailed,
			errors.New("missing partner ID"),
		)
	})
}

// ValidatePartnerIsOneOf provides a processor that validates that a message
// has at least one partner ID that matches one of the provided partners.  If
// the message is valid, the processor returns ErrNotHandled.
func ValidatePartnerIsOneOf(partners ...string) Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		ids := msg.TrimmedPartnerIDs()
		for _, partner := range partners {
			for _, id := range ids {
				if partner == id {
					// Found a match, return success
					return ErrNotHandled
				}
			}
		}

		return errors.Join(
			ErrValidationFailed,
			fmt.Errorf("expected one of: %s in the list: '%s'",
				strings.Join(partners, ", "),
				strings.Join(ids, ", ")),
		)
	})
}

// ValidatePartnerIs() provides a processor that validates that a message has
// exactly one partner ID, and that it matches the provided partner.  If the
// message is valid, the processor returns ErrNotHandled.
func ValidatePartnerIs(partner string) Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		ids := msg.TrimmedPartnerIDs()
		if len(ids) == 1 && ids[0] == partner {
			return ErrNotHandled
		}

		return errors.Join(
			ErrValidationFailed,
			fmt.Errorf("expected: %s in the list: '%s'", partner,
				strings.Join(ids, ", ")),
		)
	})
}

// ValidateNoPartners provides a processor that validates that a message has no
// partner IDs, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateNoPartners() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if len(msg.PartnerIDs) == 0 {
			return ErrNotHandled
		}

		return errors.Join(
			ErrValidationFailed,
			fmt.Errorf("unexpected partner IDs: %s",
				strings.Join(msg.PartnerIDs, ", ")),
		)
	})
}

//------------------------------------------------------------------------------

// ValidateNoSessionID provides a processor that validates that the message has no
// session ID, or returns an error.  If the message is valid, the processor returns
// ErrNotHandled.
func ValidateNoSessionID() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.SessionID == "" {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected session ID"))
	})
}

//------------------------------------------------------------------------------

func ValidateQualityOfService() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if !msg.QualityOfService.Valid() {
			return errors.Join(ErrValidationFailed, errors.New("missing quality of service"))
		}

		return ErrNotHandled
	})
}

func ValidateNoQualityOfService() Processor {
	return ProcessorFunc(func(_ context.Context, msg Message) error {
		if msg.QualityOfService == 0 {
			return ErrNotHandled
		}

		return errors.Join(ErrValidationFailed, errors.New("unexpected quality of service"))
	})
}

//------------------------------------------------------------------------------

// ValidateTypeIsAuthorization provides a processor that validates that the message is
// an authorization message type, or returns an error.  If the message is valid,
// the processor returns ErrNotHandled.
func ValidateTypeIsAuthorization() Processor {
	return Processors{
		// want
		ValidateMessageTypeIs(AuthorizationMessageType),
		ValidateStatus(),

		// do not want
		ValidateNoSource(),
		ValidateNoDestination(),
		ValidateNoTransactionUUID(),
		ValidateNoContentType(),
		ValidateNoAccept(),
		ValidateNoRequestDeliveryResponse(),
		ValidateNoHeaders(),
		ValidateNoMetadata(),
		ValidateNoPath(),
		ValidateNoPayload(),
		ValidateNoServiceName(),
		ValidateNoURL(),
		ValidateNoPartners(),
		ValidateNoSessionID(),
		ValidateNoQualityOfService(),
	}
}

// ValidateTypeIsSimpleRequestResponse provides a processor that validates that
// the message is a request/response, or returns an error.  If the message is
// valid, the processor returns ErrNotHandled.
func ValidateTypeIsSimpleRequestResponse() Processor {
	return Processors{
		// want
		ValidateMessageTypeIs(SimpleRequestResponseMessageType),
		ValidateUTF8(),
		ValidateSource(),
		ValidateDestination(),
		ValidateTransactionUUID(),
		ValidateQualityOfService(),

		// do not want
		ValidateNoPath(),
		ValidateNoServiceName(),
		ValidateNoURL(),
	}
}

// ValidateTypeIsSimpleEvent provides a processor that validates that the
// message is a simple event, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateTypeIsSimpleEvent() Processor {
	return Processors{
		// want
		ValidateMessageTypeIs(SimpleEventMessageType),
		ValidateUTF8(),
		ValidateSource(),
		ValidateDestination(),
		ValidateQualityOfService(),

		// do not want
		ValidateNoStatus(),
		ValidateNoAccept(),
		ValidateNoPath(),
		ValidateNoServiceName(),
		ValidateNoURL(),
	}
}

// ValidateTypeIsCRUD provides a processor that validates that the message is a CRUD
// operation, or returns an error.  If the message is valid, the processor returns
// ErrNotHandled.
func ValidateTypeIsCRUD() Processor {
	return Processors{
		// want
		ValidateMessageTypeIsOneOf(CreateMessageType, RetrieveMessageType, UpdateMessageType, DeleteMessageType),
		ValidateUTF8(),
		ValidateSource(),
		ValidateDestination(),
		ValidateTransactionUUID(),
		ValidateQualityOfService(),

		// do not want
		ValidateNoServiceName(),
		ValidateNoURL(),
	}
}

// ValidateTypeIsServiceRegistration provides a processor that validates that the message
// is a service registration, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateTypeIsServiceRegistration() Processor {
	return Processors{
		// want
		ValidateMessageTypeIs(ServiceRegistrationMessageType),
		ValidateUTF8(),
		ValidateServiceName(),
		ValidateURL(),

		// do not want
		ValidateNoSource(),
		ValidateNoDestination(),
		ValidateNoTransactionUUID(),
		ValidateNoContentType(),
		ValidateNoAccept(),
		ValidateNoStatus(),
		ValidateNoRequestDeliveryResponse(),
		ValidateNoHeaders(),
		ValidateNoMetadata(),
		ValidateNoPath(),
		ValidateNoPayload(),
		ValidateNoPartners(),
		ValidateNoSessionID(),
		ValidateNoQualityOfService(),
	}
}

// ValidateTypeIsServiceAlive provides a processor that validates that the message is a
// service alive, or returns an error.  If the message is valid, the processor
// returns ErrNotHandled.
func ValidateTypeIsServiceAlive() Processor {
	return Processors{
		// want
		ValidateMessageTypeIs(ServiceAliveMessageType),

		// do not want
		ValidateNoSource(),
		ValidateNoDestination(),
		ValidateNoTransactionUUID(),
		ValidateNoContentType(),
		ValidateNoAccept(),
		ValidateNoStatus(),
		ValidateNoRequestDeliveryResponse(),
		ValidateNoHeaders(),
		ValidateNoMetadata(),
		ValidateNoPath(),
		ValidateNoServiceName(),
		ValidateNoURL(),
		ValidateNoPayload(),
		ValidateNoPartners(),
		ValidateNoSessionID(),
		ValidateNoQualityOfService(),
	}
}

// ValidateTypeIsUnknown provides a processor that validates that the message
// is an unknown message type, or returns an error.  If the message is valid, the
// processor returns ErrNotHandled.
func ValidateTypeIsUnknown() Processor {
	return Processors{
		// want
		ValidateMessageTypeIs(UnknownMessageType),

		// do not want
		ValidateNoSource(),
		ValidateNoDestination(),
		ValidateNoTransactionUUID(),
		ValidateNoContentType(),
		ValidateNoAccept(),
		ValidateNoStatus(),
		ValidateNoRequestDeliveryResponse(),
		ValidateNoHeaders(),
		ValidateNoMetadata(),
		ValidateNoPath(),
		ValidateNoServiceName(),
		ValidateNoURL(),
		ValidateNoPayload(),
		ValidateNoPartners(),
		ValidateNoSessionID(),
		ValidateNoQualityOfService(),
	}
}

// ValidateIsRoutable provides a processor that validates that the message is
// routable, or returns an error.  If the message is valid, the processor returns
// ErrNotHandled.
func ValidateIsRoutable() Processor {
	return Processors{
		ValidateMessageTypeIsOneOf(
			SimpleRequestResponseMessageType,
			SimpleEventMessageType,
			CreateMessageType,
			RetrieveMessageType,
			UpdateMessageType,
			DeleteMessageType,
		),
		ValidateUTF8(),
		ValidateSource(),
		ValidateDestination(),
		ValidateTransactionUUID(),
	}
}
