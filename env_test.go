// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvMap(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		output   map[string]string
		onlyTo   bool
		onlyFrom bool
	}{
		{
			name: "Basic struct",
			input: &struct {
				Name string `env:"NAME"`
				Age  *int64 `env:"AGE"`
			}{
				Name: "John",
				Age:  int64Ptr(30),
			},
			output: map[string]string{
				"NAME": "John",
				"AGE":  "30",
			},
		}, {
			name: "Handle ignored things",
			input: &struct {
				Name     string `env:"NAME"`
				Age      *int64 `env:"AGE"`
				Ignored  string
				Ignored2 string `env:"-"`
				Ignored3 string `env:""`
				Ignored4 string `env:",omitempty"`
			}{
				Name:     "John",
				Age:      int64Ptr(30),
				Ignored:  "ignored",
				Ignored2: "ignored",
				Ignored3: "ignored",
				Ignored4: "ignored",
			},
			output: map[string]string{
				"NAME": "John",
				"AGE":  "30",
			},
			onlyTo: true,
		}, {
			name: "Omit empty fields",
			input: &struct {
				Name string `env:"NAME,omitempty"`
				Age  *int64 `env:"AGE,omitempty"`
			}{
				Name: "",
				Age:  nil,
			},
			output: map[string]string{},
		},
		{
			name: "Include empty fields",
			input: &struct {
				Name string `env:"NAME"`
				Age  *int64 `env:"AGE"`
			}{
				Name: "",
				Age:  nil,
			},
			output: map[string]string{
				"NAME": "",
				"AGE":  "",
			},
		}, {
			name: "Multiline slice",
			input: &struct {
				Lines []string `env:"LINE,multiline"`
			}{
				Lines: []string{"line1", "line2", "line3"},
			},
			output: map[string]string{
				"LINE_000": "line1",
				"LINE_001": "line2",
				"LINE_002": "line3",
			},
		}, {
			name: "Base64 encode []byte",
			input: &struct {
				Data []byte `env:"DATA"`
			}{
				Data: []byte("hello"),
			},
			output: map[string]string{
				"DATA": "aGVsbG8=",
			},
		}, {
			name: "Map of strings",
			input: &struct {
				Labels map[string]string `env:"LABEL"`
			}{
				Labels: map[string]string{"key1": "value1", "key2": "value2"},
			},
			output: map[string]string{
				"LABEL_key1": "key1=value1",
				"LABEL_key2": "key2=value2",
			},
		}, {
			name: "array of of strings",
			input: &struct {
				Labels []string `env:"LABEL"`
			}{
				Labels: []string{"hello world", "goodbye world"},
			},
			output: map[string]string{
				"LABEL": "hello world,goodbye world",
			},
		},
		{
			name: "wrp.Message",
			input: &Message{
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
				Payload:          []byte("hello world"),
				ServiceName:      "device-status",
				URL:              "http://localhost:8080/api/v1/device-status/foo",
				PartnerIDs:       []string{"partner1", "partner2"},
				SessionID:        "1234",
				QualityOfService: 12,
			},
			output: map[string]string{
				"WRP_MSG_TYPE":         "4",
				"WRP_SOURCE":           "mac:112233445566",
				"WRP_DEST":             "event:device-status/foo",
				"WRP_TRANSACTION_UUID": "1234",
				"WRP_CONTENT_TYPE":     "application/json",
				"WRP_ACCEPT":           "application/json",
				"WRP_STATUS":           "200",
				"WRP_RDR":              "1",
				"WRP_HEADERS_000":      "key1:value1",
				"WRP_HEADERS_001":      "key2:value2",
				"WRP_METADATA_key_1":   "/key/1=value1",
				"WRP_METADATA_key_2":   "/key/2=value2",
				"WRP_PATH":             "/api/v1/device-status/foo",
				"WRP_PAYLOAD":          "aGVsbG8gd29ybGQ=",
				"WRP_SERVICE_NAME":     "device-status",
				"WRP_URL":              "http://localhost:8080/api/v1/device-status/foo",
				"WRP_PARTNER_IDS":      "partner1,partner2",
				"WRP_SESSION_ID":       "1234",
				"WRP_QOS":              "12",
			},
		},
	}

	for _, tt := range tests {
		if !tt.onlyFrom {
			t.Run("toEnvMap: "+tt.name, func(t *testing.T) {
				got := toEnvMap(tt.input)
				assert.Equal(t, tt.output, got)
			})
		}

		if !tt.onlyTo {
			t.Run("fromEnvMap: "+tt.name, func(t *testing.T) {
				newInstance := reflect.New(reflect.TypeOf(tt.input).Elem()).Interface()

				var list []string
				for k, v := range tt.output {
					list = append(list, k+"="+v)
				}
				// Populate the new instance from the environment variables
				err := fromEnvMap(list, newInstance)
				assert.NoError(t, err)

				// Ensure the new instance matches the original input
				assert.Equal(t, tt.input, newInstance)

			})
		}
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}
