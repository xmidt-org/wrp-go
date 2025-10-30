// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

// modality is used to indicate if a field is required, optional, or unwanted.
type modality int

const (
	unwanted modality = iota
	optional
	required
)

// vador is a struct that contains all the fields that need to be validated and
// the modality of each field.
type vador struct {
	Type                    MessageType
	Source                  modality
	Destination             modality
	TransactionUUID         modality
	ContentType             modality
	Accept                  modality
	RequestDeliveryResponse modality
	Status                  modality
	Headers                 modality
	Metadata                modality
	Path                    modality
	Payload                 modality
	ServiceName             modality
	URL                     modality
	PartnerIDs              modality
	SessionID               modality
	QualityOfService        modality
}

func (v vador) ProcessWRP(ctx context.Context, msg Message) error {
	if msg.Type != v.Type {
		return ErrNotHandled
	}

	var err error
	v.locator(&err, v.Source, msg.Source, "Source")
	v.locator(&err, v.Destination, msg.Destination, "Destination")
	v.string(&err, v.TransactionUUID, msg.TransactionUUID, "TransactionUUID")
	v.string(&err, v.ContentType, msg.ContentType, "ContentType")
	v.string(&err, v.Accept, msg.Accept, "Accept")
	v.int64p(&err, v.RequestDeliveryResponse, msg.RequestDeliveryResponse, "RequestDeliveryResponse")
	v.int64p(&err, v.Status, msg.Status, "Status")
	v.strings(&err, v.Headers, msg.Headers, "Headers")
	v.metadata(&err, &msg, v.Metadata)
	v.string(&err, v.Path, msg.Path, "Path")
	v.payload(&err, &msg, v.Payload)
	v.serviceName(&err, &msg, v.ServiceName)
	v.string(&err, v.URL, msg.URL, "URL")
	v.strings(&err, v.PartnerIDs, msg.PartnerIDs, "PartnerIDs")
	v.string(&err, v.SessionID, msg.SessionID, "SessionID")
	v.qos(&err, &msg, v.QualityOfService)

	return err
}

func (vador) locator(err *error, m modality, field, name string) {
	if *err != nil {
		return
	}

	switch m {
	case required:
		if field == "" {
			*err = errors.Join(ErrMessageIsInvalid, errors.New(name+" is required"))
			return
		}
	case optional:
	case unwanted:
		if field != "" {
			*err = errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
			return
		}
	}

	if field != "" {
		if !utf8.ValidString(field) {
			*err = errors.Join(ErrMessageIsInvalid, ErrNotUTF8, errors.New("invalid "+name))
			return
		}

		if _, er := ParseLocator(field); er != nil {
			*err = errors.Join(ErrMessageIsInvalid, er)
			return
		}
	}
}

func (v vador) serviceName(err *error, msg *Message, m modality) {
	v.string(err, m, msg.ServiceName, "ServiceName")
	if *err != nil {
		return
	}
	if strings.Contains(msg.ServiceName, "/") {
		*err = errors.Join(ErrMessageIsInvalid, errors.New("service_name cannot contain '/'"))
	}
}

func (vador) string(err *error, m modality, field, name string) {
	if *err != nil {
		return
	}

	switch m {
	case required:
		if field == "" {
			*err = errors.Join(ErrMessageIsInvalid, errors.New(name+" is required"))
			return
		}
	case optional:
	case unwanted:
		if field != "" {
			*err = errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
			return
		}
	}

	if field != "" {
		if !utf8.ValidString(field) {
			*err = errors.Join(ErrMessageIsInvalid, ErrNotUTF8, errors.New("invalid "+name))
			return
		}
	}
}

func (vador) int64p(err *error, m modality, field *int64, name string) {
	if *err != nil {
		return
	}

	switch m {
	case required:
		if field == nil {
			*err = errors.Join(ErrMessageIsInvalid, errors.New(name+" is required"))
			return
		}
	case optional:
	case unwanted:
		if field != nil {
			*err = errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
			return
		}
	}
}

func (vador) payload(err *error, msg *Message, m modality) {
	if *err != nil {
		return
	}

	switch m {
	case required:
		if len(msg.Payload) == 0 {
			*err = errors.Join(ErrMessageIsInvalid, errors.New("Payload is required")) // nolint:staticcheck
			return
		}
	case optional:
	case unwanted:
		if len(msg.Payload) != 0 {
			*err = errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
			return
		}
	}
}

func (vador) strings(err *error, m modality, field []string, name string) {
	if *err != nil {
		return
	}

	switch m {
	case required:
		if len(field) == 0 {
			*err = errors.Join(ErrMessageIsInvalid, errors.New(name+" is required"))
			return
		}
	case optional:
	case unwanted:
		if len(field) != 0 {
			*err = errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
			return
		}
	}

	for _, f := range field {
		if !utf8.ValidString(f) {
			*err = errors.Join(ErrMessageIsInvalid, ErrNotUTF8, errors.New("invalid "+name))
			return
		}
	}
}

func (vador) metadata(err *error, msg *Message, m modality) {
	if *err != nil {
		return
	}

	switch m {
	case required:
		if len(msg.Metadata) == 0 {
			*err = errors.Join(ErrMessageIsInvalid, errors.New("Payload is required")) // nolint:staticcheck
			return
		}
	case optional:
	case unwanted:
		if len(msg.Metadata) != 0 {
			*err = errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
			return
		}
	}

	for k, v := range msg.Metadata {
		if !utf8.ValidString(k) || !utf8.ValidString(v) {
			*err = errors.Join(ErrMessageIsInvalid, ErrNotUTF8, errors.New("invalid Metadata"))
			return
		}
	}
}

func (vador) qos(err *error, msg *Message, m modality) {
	if *err != nil {
		return
	}

	switch m {
	case required, optional:
		if !msg.QualityOfService.Valid() {
			*err = errors.Join(ErrMessageIsInvalid, errors.New("invalid QualityOfService"))
			return
		}
	case unwanted:
		if msg.QualityOfService != 0 {
			*err = errors.Join(ErrMessageIsInvalid, ErrUnsupportedFieldsSet)
			return
		}
	}
}
