// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp_test

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/wrp-go/v4"
)

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

func TestNewMessageFromHeadersSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expectedStatus                  int64 = 928
		expectedRequestDeliveryResponse int64 = 1
		expectedIncludeSpans            bool  = true

		tests = []struct {
			desc     string
			header   http.Header
			payload  []byte
			expected wrp.Message
		}{
			{
				desc: "simple request response with normal headers",
				header: http.Header{
					MessageTypeHeader: []string{"SimpleRequestResponse"},
				},
				payload: nil,
				expected: wrp.Message{
					Type: wrp.SimpleRequestResponseMessageType,
				},
			},
			{
				desc: "simple request response with legacy headers",
				header: http.Header{
					msgTypeHeader: []string{"SimpleRequestResponse"},
				},
				payload: nil,
				expected: wrp.Message{
					Type: wrp.SimpleRequestResponseMessageType,
				},
			},
			{
				desc: "simple request response with empty payload",
				header: http.Header{
					MessageTypeHeader: []string{"SimpleRequestResponse"},
				},
				payload: []byte{},
				expected: wrp.Message{
					Type:    wrp.SimpleRequestResponseMessageType,
					Payload: []byte{},
				},
			},
			{
				desc: "full req/resp message with normal headers",
				header: http.Header{
					MessageTypeHeader:             []string{"SimpleRequestResponse"},
					TransactionUuidHeader:         []string{"1234"},
					SourceHeader:                  []string{"test"},
					DestinationHeader:             []string{"mac:111122223333"},
					StatusHeader:                  []string{strconv.FormatInt(expectedStatus, 10)},
					RequestDeliveryResponseHeader: []string{strconv.FormatInt(expectedRequestDeliveryResponse, 10)},
					IncludeSpansHeader:            []string{strconv.FormatBool(expectedIncludeSpans)},
					SpanHeader: []string{
						"foo, bar, moo",
						"goo, gar, hoo",
					},
					AcceptHeader:      []string{wrp.MimeTypeJson},
					PathHeader:        []string{"/foo/bar"},
					SessionIdHeader:   []string{"test123"},
					HeadersHeader:     []string{"head-1", "head-2"},
					ServiceNameHeader: []string{"service"},
					URLHeader:         []string{"anonspecialurl"},
				},
				payload: nil,
				expected: wrp.Message{
					Type:                    wrp.SimpleRequestResponseMessageType,
					TransactionUUID:         "1234",
					Source:                  "test",
					Destination:             "mac:111122223333",
					Status:                  &expectedStatus,
					RequestDeliveryResponse: &expectedRequestDeliveryResponse,
					Accept:                  wrp.MimeTypeJson,
					Path:                    "/foo/bar",
					SessionID:               "test123",
					Headers:                 []string{"head-1", "head-2"},
					ServiceName:             "service",
					URL:                     "anonspecialurl",
				},
			},
			{
				desc: "full req/resp message with legacy headers",
				header: http.Header{
					msgTypeHeader:         []string{"SimpleRequestResponse"},
					transactionUuidHeader: []string{"1234"},
					sourceHeader:          []string{"test"},
					destinationHeader:     []string{"mac:111122223333"},
					statusHeader:          []string{strconv.FormatInt(expectedStatus, 10)},
					rDRHeader:             []string{strconv.FormatInt(expectedRequestDeliveryResponse, 10)},
					acceptHeader:          []string{wrp.MimeTypeJson},
					pathHeader:            []string{"/foo/bar"},
					sessionIdHeader:       []string{"test123"},
					headersArrHeader:      []string{"head-1", "head-2"},
					serviceNameHeader:     []string{"service"},
					urlHeader:             []string{"anonspecialurl"},
				},
				payload: nil,
				expected: wrp.Message{
					Type:                    wrp.SimpleRequestResponseMessageType,
					TransactionUUID:         "1234",
					Source:                  "test",
					Destination:             "mac:111122223333",
					Status:                  &expectedStatus,
					RequestDeliveryResponse: &expectedRequestDeliveryResponse,
					Accept:                  wrp.MimeTypeJson,
					Path:                    "/foo/bar",
					SessionID:               "test123",
					Headers:                 []string{"head-1", "head-2"},
					ServiceName:             "service",
					URL:                     "anonspecialurl",
				},
			},
			{
				desc: "an event message with normal headers",
				header: http.Header{
					MessageTypeHeader: []string{"SimpleEvent"},
					SourceHeader:      []string{"test"},
					DestinationHeader: []string{"mac:111122223333"},
					"Content-Type":    []string{"text/plain"},
				},
				payload: []byte("payload"),
				expected: wrp.Message{
					Type:        wrp.SimpleEventMessageType,
					Source:      "test",
					Destination: "mac:111122223333",
					ContentType: "text/plain",
					Payload:     []byte("payload"),
				},
			},
			{
				desc: "an event message with default content type",
				header: http.Header{
					MessageTypeHeader: []string{"SimpleEvent"},
					SourceHeader:      []string{"test"},
					DestinationHeader: []string{"mac:111122223333"},
				},
				payload: []byte("payload"),
				expected: wrp.Message{
					Type:        wrp.SimpleEventMessageType,
					Source:      "test",
					Destination: "mac:111122223333",
					ContentType: wrp.MimeTypeOctetStream,
					Payload:     []byte("payload"),
				},
			},
			{
				desc: "an event message with extra headers",
				header: http.Header{
					MessageTypeHeader: []string{"SimpleEvent"},
					SourceHeader:      []string{"test"},
					DestinationHeader: []string{"mac:111122223333"},
					MetadataHeader:    []string{"/foo=bar,/goo=car", "/dog=food", "/tag", "/slag="},
				},
				payload: []byte("payload"),
				expected: wrp.Message{
					Type:        wrp.SimpleEventMessageType,
					Source:      "test",
					Destination: "mac:111122223333",
					ContentType: wrp.MimeTypeOctetStream,
					Payload:     []byte("payload"),
					Metadata: map[string]string{"/foo": "bar,/goo=car",
						"/dog":  "food",
						"/tag":  "",
						"/slag": ""},
				},
			},
			{
				desc: "an event message with extra partners",
				header: http.Header{
					MessageTypeHeader: []string{"SimpleEvent"},
					SourceHeader:      []string{"test"},
					DestinationHeader: []string{"mac:111122223333"},
					PartnerIdHeader:   []string{"partner-1 , partner-2,partner-3", " p4 "},
				},
				payload: []byte("payload"),
				expected: wrp.Message{
					Type:        wrp.SimpleEventMessageType,
					Source:      "test",
					Destination: "mac:111122223333",
					ContentType: wrp.MimeTypeOctetStream,
					Payload:     []byte("payload"),
					PartnerIDs:  []string{"partner-1", "partner-2", "partner-3", "p4"},
				},
			},
		}
	)

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			actual, err := wrp.NewMessageFromHeaders(tc.header, tc.payload)
			require.NotNil(actual)
			require.Equal(tc.expected, *actual)
			assert.NoError(err)
		})
	}
}
