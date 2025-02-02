// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    Message
		output   http.Header
		onlyTo   bool
		onlyFrom bool
		errFrom  bool
	}{
		{
			name: "empty",
		}, {
			name: "simple",
			input: Message{
				Type:                    SimpleEventMessageType,
				Source:                  "mac:112233445566",
				Destination:             "event:device-status/foo",
				TransactionUUID:         "1234",
				ContentType:             "application/json",
				Accept:                  "application/json",
				Status:                  int64Ptr(200),
				RequestDeliveryResponse: int64Ptr(1),
				Headers: []string{
					"key1:value1",
					"key2:value2",
				},
				Metadata: map[string]string{
					"/key/1": "value1",
					"/key/2": "value2",
				},
				Path:             "/api/v1/device-status/foo",
				ServiceName:      "device-status",
				URL:              "http://localhost:8080/api/v1/device-status/foo",
				PartnerIDs:       []string{"partner1", "partner2"},
				SessionID:        "1234",
				QualityOfService: 12,
			},
			output: http.Header{
				"X-Xmidt-Message-Type":              []string{"4"},
				"X-Xmidt-Source":                    []string{"mac:112233445566"},
				"X-Webpa-Device-Name":               []string{"event:device-status/foo"},
				"X-Xmidt-Transaction-Uuid":          []string{"1234"},
				"Content-Type":                      []string{"application/json"},
				"X-Xmidt-Accept":                    []string{"application/json"},
				"X-Xmidt-Status":                    []string{"200"},
				"X-Xmidt-Request-Delivery-Response": []string{"1"},
				"X-Xmidt-Headers":                   []string{"key1:value1", "key2:value2"},
				"X-Xmidt-Metadata":                  []string{"/key/1=value1", "/key/2=value2"},
				"X-Xmidt-Path":                      []string{"/api/v1/device-status/foo"},
				"X-Xmidt-Service-Name":              []string{"device-status"},
				"X-Xmidt-Url":                       []string{"http://localhost:8080/api/v1/device-status/foo"},
				"X-Xmidt-Partner-Id":                []string{"partner1,partner2"},
				"X-Xmidt-Session-Id":                []string{"1234"},
				"X-Xmidt-Qos":                       []string{"12"},
			},
		}, {
			name: "invalid status",
			output: http.Header{
				"X-Xmidt-Status": []string{"invalid"},
			},
			errFrom:  true,
			onlyFrom: true,
		}, {
			name: "invalid qos",
			output: http.Header{
				"X-Xmidt-Qos": []string{"invalid"},
			},
			errFrom:  true,
			onlyFrom: true,
		},
	}

	for _, tt := range tests {
		if !tt.onlyFrom {
			t.Run("toHeaders: "+tt.name, func(t *testing.T) {
				got := http.Header{}
				msg := tt.input
				toHeaders(&msg, got)

				for k, v := range tt.output {
					assert.ElementsMatch(t, v, got[k])
				}
			})
		}

		if !tt.onlyTo {
			t.Run("fromHeaders: "+tt.name, func(t *testing.T) {
				var got Message
				// Populate the new instance from the environment variables
				err := fromHeaders(tt.output, &got)
				if tt.errFrom {
					assert.Error(t, err)
					return
				}

				assert.NoError(t, err)

				// Ensure the new instance matches the original input
				assert.Equal(t, tt.input, got)
			})
		}
	}
}
