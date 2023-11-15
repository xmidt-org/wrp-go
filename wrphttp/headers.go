// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrphttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/xmidt-org/wrp-go/v3"
)

// Constant HTTP header strings representing WRP fields
const (
	MessageTypeHeader             = "X-Xmidt-Message-Type"
	TransactionUuidHeader         = "X-Xmidt-Transaction-Uuid" // nolint:gosec
	StatusHeader                  = "X-Xmidt-Status"
	RequestDeliveryResponseHeader = "X-Xmidt-Request-Delivery-Response"
	IncludeSpansHeader            = "X-Xmidt-Include-Spans"
	SpanHeader                    = "X-Xmidt-Span"
	PathHeader                    = "X-Xmidt-Path"
	SourceHeader                  = "X-Xmidt-Source"
	DestinationHeader             = "X-Webpa-Device-Name"
	AcceptHeader                  = "X-Xmidt-Accept"
	MetadataHeader                = "X-Xmidt-Metadata"
	PartnerIdHeader               = "X-Xmidt-Partner-Id"
	SessionIdHeader               = "X-Xmidt-Session-Id"
	HeadersHeader                 = "X-Xmidt-Headers"
	ServiceNameHeader             = "X-Xmidt-Service-Name"
	URLHeader                     = "X-Xmidt-Url"
)

// X-midt-* headers are deprecated and will stop being supported
// Please use X-Xmidt-* headers instead
const (
	msgTypeHeader         = "X-Midt-Msg-Type"
	transactionUuidHeader = "X-Midt-Transaction-Uuid"
	statusHeader          = "X-Midt-Status"
	rDRHeader             = "X-Midt-Request-Delivery-Response"
	headersArrHeader      = "X-Midt-Headers"
	includeSpansHeader    = "X-Midt-Include-Spans"
	spansHeader           = "X-Midt-Spans"
	pathHeader            = "X-Midt-Path"
	sourceHeader          = "X-Midt-Source"
	destinationHeader     = "X-Webpa-Device-Name"
	acceptHeader          = "X-Midt-Accept"
	metadataHeader        = "X-Midt-Metadata"
	partnerIdHeader       = "X-Midt-Partner-Id"
	sessionIdHeader       = "X-Midt-Session-Id"
	headersHeader         = "X-Midt-Headers"
	serviceNameHeader     = "X-Midt-Service-Name"
	urlHeader             = "X-Midt-Url"
)

var (
	errMissingMessageTypeHeader = fmt.Errorf("missing %s header", MessageTypeHeader)
)

// getMessageType extracts the wrp.MessageType from header.  This is a required field.
//
// This function panics if the message type header is missing or invalid.
func getMessageType(h http.Header) wrp.MessageType {
	value := h.Get(MessageTypeHeader)
	if len(value) == 0 {
		// check alternative header prefix
		value = h.Get(msgTypeHeader)
		if len(value) == 0 {
			panic(errMissingMessageTypeHeader)
		}
	}

	messageType := wrp.StringToMessageType(value)
	return messageType
}

// getIntHeader returns the header as a int64, or returns nil if the header is absent.
// This function panics if the header is present but not a valid integer.
func getIntHeader(h http.Header, n string) *int64 {
	value := h.Get(n)
	if len(value) == 0 {
		return nil
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		panic(err)
	}

	return &i
}

func getBoolHeader(h http.Header, n string) *bool {
	value := h.Get(n)
	if len(value) == 0 {
		return nil
	}

	b, err := strconv.ParseBool(value)
	if err != nil {
		panic(err)
	}

	return &b
}

func getSpans(h http.Header) [][]string {
	header := SpanHeader
	if len(h[SpanHeader]) == 0 {
		if len(h[spansHeader]) == 0 {
			return nil
		}
		header = spansHeader // alternative header version
	}

	spans := make([][]string, len(h[header]))

	for i, value := range h[header] {
		fields := strings.Split(value, ",")
		if len(fields) != 3 {
			panic(fmt.Errorf("invalid %s header: %s", header, value))
		}

		for j := 0; j < len(fields); j++ {
			fields[j] = strings.TrimSpace(fields[j])
		}

		spans[i] = fields
	}

	return spans
}

// getMetadata returns the map that represents the metadata fields that were
// passed in as headers.  This function handles multiple duplicate headers.
// This function panics if the header contains data that is not a name=value
// pair.
func getMetadata(h http.Header) map[string]string {
	headers, ok := h[MetadataHeader]
	if !ok {
		// Check alternative header version
		headers, ok = h[metadataHeader]
		if !ok {
			return nil
		}
	}

	meta := make(map[string]string)

	for _, value := range headers {
		fields := strings.Split(value, ",")
		for _, v := range fields {
			kv := strings.Split(v, "=")
			if 0 < len(kv) {
				key := strings.TrimSpace(kv[0])
				kv = append(kv, "")
				meta[key] = strings.Join(kv[1:], "")
			}
		}
	}
	return meta
}

// getPartnerIDs returns the array that represents the partner-ids that were
// passed in as headers.  This function handles multiple duplicate headers.
func getPartnerIDs(h http.Header) []string {
	headers, ok := h[PartnerIdHeader]
	if !ok || len(headers) == 0 {
		// Check alternative header version
		headers, ok = h[partnerIdHeader]

		if !ok || len(headers) == 0 {
			return nil
		}
	}

	partners := []string{}

	for _, value := range headers {
		fields := strings.Split(value, ",")
		for i := 0; i < len(fields); i++ {
			fields[i] = strings.TrimSpace(fields[i])
		}
		partners = append(partners, fields...)
	}

	return partners
}

func readPayload(h http.Header, p io.Reader) ([]byte, string) {
	if p == nil {
		return nil, ""
	}

	payload, err := io.ReadAll(p)
	if err != nil {
		panic(err)
	}

	if len(payload) == 0 {
		return nil, ""
	}

	contentType := h.Get("Content-Type")
	if len(contentType) == 0 && len(payload) > 0 {
		contentType = wrp.MimeTypeOctetStream
	}

	return payload, contentType
}

// getHeaders returns the array that represents the headers that were
// passed in as headers.  This function handles multiple duplicate headers.
func getHeaders(h http.Header) []string {
	headers, ok := h[HeadersHeader]
	if !ok || len(headers) == 0 {
		// Check alternative header version
		headers, ok = h[headersHeader]

		if !ok || len(headers) == 0 {
			return nil
		}
	}

	hlist := []string{}

	for _, value := range headers {
		fields := strings.Split(value, ",")
		for i := 0; i < len(fields); i++ {
			fields[i] = strings.TrimSpace(fields[i])
		}
		hlist = append(hlist, fields...)
	}

	return hlist
}

// NewMessageFromHeaders extracts a WRP message from a set of HTTP headers.  If supplied, the
// given io.Reader is assumed to contain the payload of the WRP message.
func NewMessageFromHeaders(h http.Header, p io.Reader) (message *wrp.Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			message = nil
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("unable to create WRP message: %s", v)
			}
		}
	}()

	payload, contentType := readPayload(h, p)
	message = new(wrp.Message)
	err = SetMessageFromHeaders(h, message)
	if err != nil {
		message = nil
	}

	message.Payload = payload
	message.ContentType = contentType
	return
}

// SetMessageFromHeaders transfers header fields onto the given WRP message.  The payload is not
// handled by this method.
func SetMessageFromHeaders(h http.Header, m *wrp.Message) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("Unable to create WRP message: %s", v)
			}
		}
	}()

	m.Type = getMessageType(h)
	m.Source = h.Get(SourceHeader)
	if m.Source == "" {
		m.Source = h.Get(sourceHeader)
	}
	m.Destination = h.Get(DestinationHeader)
	if m.Destination == "" {
		m.Destination = h.Get(destinationHeader)
	}
	m.TransactionUUID = h.Get(TransactionUuidHeader)
	if m.TransactionUUID == "" {
		m.TransactionUUID = h.Get(transactionUuidHeader)
	}
	m.Status = getIntHeader(h, StatusHeader)
	if m.Status == nil {
		m.Status = getIntHeader(h, statusHeader)
	}
	m.RequestDeliveryResponse = getIntHeader(h, RequestDeliveryResponseHeader)
	if m.RequestDeliveryResponse == nil {
		m.RequestDeliveryResponse = getIntHeader(h, rDRHeader)
	}
	// TODO Remove along with `IncludeSpans`
	// nolint:staticcheck
	m.IncludeSpans = getBoolHeader(h, IncludeSpansHeader)
	// nolint:staticcheck
	if m.IncludeSpans == nil {
		// nolint:staticcheck
		m.IncludeSpans = getBoolHeader(h, includeSpansHeader)
	}
	m.Spans = getSpans(h)
	m.ContentType = h.Get("Content-Type")
	m.Accept = h.Get(AcceptHeader)
	if m.Accept == "" {
		m.Accept = h.Get(acceptHeader)
	}
	m.Path = h.Get(PathHeader)
	if m.Path == "" {
		m.Path = h.Get(pathHeader)
	}
	m.Metadata = getMetadata(h)
	m.PartnerIDs = getPartnerIDs(h)
	m.SessionID = h.Get(SessionIdHeader)
	if m.SessionID == "" {
		m.SessionID = h.Get(sessionIdHeader)
	}
	m.Headers = getHeaders(h)
	m.ServiceName = h.Get(ServiceNameHeader)
	if m.ServiceName == "" {
		m.ServiceName = h.Get(serviceNameHeader)
	}
	m.URL = h.Get(URLHeader)
	if m.URL == "" {
		m.URL = h.Get(urlHeader)
	}
	return
}

// AddMessageHeaders adds the HTTP header representation of a given WRP message.
// This function does not handle the payload, to allow further headers to be written by
// calling code.
func AddMessageHeaders(h http.Header, m *wrp.Message) {
	h.Set(MessageTypeHeader, m.Type.FriendlyName())

	if len(m.Source) > 0 {
		h.Set(SourceHeader, m.Source)
	}

	if len(m.Destination) > 0 {
		h.Set(DestinationHeader, m.Destination)
	}

	if len(m.TransactionUUID) > 0 {
		h.Set(TransactionUuidHeader, m.TransactionUUID)
	}

	if m.Status != nil {
		h.Set(StatusHeader, strconv.FormatInt(*m.Status, 10))
	}

	if m.RequestDeliveryResponse != nil {
		h.Set(RequestDeliveryResponseHeader, strconv.FormatInt(*m.RequestDeliveryResponse, 10))
	}

	// TODO Remove along with `IncludeSpans`
	// nolint:staticcheck
	if m.IncludeSpans != nil {
		h.Set(IncludeSpansHeader, strconv.FormatBool(*m.IncludeSpans))
	}

	for _, s := range m.Spans {
		h.Add(SpanHeader, strings.Join(s, ","))
	}

	if len(m.Accept) > 0 {
		h.Set(AcceptHeader, m.Accept)
	}

	if len(m.Path) > 0 {
		h.Set(PathHeader, m.Path)
	}

	for k, v := range m.Metadata {
		// perform k + "=" + v more efficiently
		buf := bytes.Buffer{}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(v)
		h.Add(MetadataHeader, buf.String())
	}

	for _, v := range m.PartnerIDs {
		h.Add(PartnerIdHeader, v)
	}

	if len(m.SessionID) > 0 {
		h.Set(SessionIdHeader, m.SessionID)
	}

	for _, v := range m.Headers {
		h.Add(HeadersHeader, v)
	}

	if len(m.ServiceName) > 0 {
		h.Set(ServiceNameHeader, m.ServiceName)
	}

	if len(m.URL) > 0 {
		h.Set(URLHeader, m.URL)
	}
}

// ReadPayload extracts the payload from a reader, setting the appropriate
// fields on the given message.
func ReadPayload(h http.Header, p io.Reader, m *wrp.Message) (int, error) {
	contentType := h.Get("Content-Type")
	if len(contentType) == 0 {
		contentType = wrp.MimeTypeOctetStream
	}

	var err error
	m.Payload, err = io.ReadAll(p)
	if err != nil {
		return 0, err
	}

	m.ContentType = contentType
	return len(m.Payload), nil
}

// WritePayload writes the WRP payload to the given io.Writer.  If the message has no
// payload, this function does nothing.
//
// The http.Header is optional.  If supplied, the header's Content-Type and Content-Length
// will be set appropriately.
func WritePayload(h http.Header, p io.Writer, m *wrp.Message) (int, error) {
	if len(m.Payload) == 0 {
		return 0, nil
	}

	if h != nil {
		if len(m.ContentType) > 0 {
			h.Set("Content-Type", m.ContentType)
		} else {
			h.Set("Content-Type", wrp.MimeTypeOctetStream)
		}

		h.Set("Content-Length", strconv.Itoa(len(m.Payload)))
	}

	return p.Write(m.Payload)
}
