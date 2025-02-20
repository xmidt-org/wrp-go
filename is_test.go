// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/wrp-go/v5"
)

func TestIs(t *testing.T) {
	tests := []struct {
		desc       string
		msg        wrp.Union
		target     wrp.Union
		validators []wrp.Processor
		want       bool
	}{
		{
			desc:   "nil msg and nil target",
			msg:    nil,
			target: nil,
			want:   true,
		}, {
			desc:   "nil msg and non-nil target",
			msg:    nil,
			target: &wrp.Message{Type: wrp.UnknownMessageType},
			want:   false,
		}, {
			desc:   "non-nil msg, type is invalid",
			msg:    &wrp.Message{Type: 0},
			target: &wrp.Message{},
			want:   false,
		}, {
			desc:   "non-nil msg and nil target",
			msg:    &wrp.Message{Type: wrp.UnknownMessageType},
			target: nil,
			want:   false,
		}, {
			desc:   "same msg and target type",
			msg:    &wrp.Message{Type: wrp.UnknownMessageType},
			target: &wrp.Message{Type: wrp.UnknownMessageType},
			want:   true,
		}, {
			desc:   "same msg and target type, using specific type",
			msg:    &wrp.Message{Type: wrp.UnknownMessageType},
			target: &wrp.Unknown{},
			want:   true,
		}, {
			desc:   "same msg and target type, using specific type, inverse",
			msg:    &wrp.Unknown{},
			target: &wrp.Message{Type: wrp.UnknownMessageType},
			want:   true,
		}, {
			desc:   "different msg and target type",
			msg:    &wrp.Message{Type: 3},
			target: &wrp.Message{Type: 4},
			want:   false,
		}, {
			desc:   "msg type matches exact type",
			msg:    &wrp.Message{Type: wrp.UnknownMessageType},
			target: &wrp.Unknown{},
			want:   true,
		}, {
			desc:   "msg type matches exact type, inverse",
			msg:    &wrp.Unknown{},
			target: &wrp.Message{Type: wrp.UnknownMessageType},
			want:   true,
		}, {
			desc:   "msg type is not the same type",
			msg:    &wrp.Message{Type: wrp.UnknownMessageType},
			target: &wrp.SimpleEvent{},
			want:   false,
		}, {
			desc:   "msg type is not the same type, inverse",
			msg:    &wrp.SimpleEvent{},
			target: &wrp.Message{Type: wrp.UnknownMessageType},
			want:   false,
		}, {
			desc:   "msg types match, but validation fails",
			msg:    &wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
			target: &wrp.SimpleRequestResponse{},
			want:   false,
		}, {
			desc:   "msg types match, but validation fails, inverse",
			msg:    &wrp.SimpleRequestResponse{},
			target: &wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
			want:   false,
		}, {
			desc:       "msg types match, but validation is skipped",
			msg:        &wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
			target:     &wrp.SimpleRequestResponse{},
			validators: []wrp.Processor{wrp.NoStandardValidation()},
			want:       true,
		}, {
			desc:       "msg types match, but validation is skipped, inverse",
			msg:        &wrp.SimpleRequestResponse{},
			target:     &wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
			validators: []wrp.Processor{wrp.NoStandardValidation()},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := wrp.Is(tt.msg, tt.target, tt.validators...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAs(t *testing.T) {
	tests := []struct {
		desc      string
		src       wrp.Union
		dst       wrp.Union
		want      wrp.Union
		Processor wrp.Processor
		wantError bool
	}{
		{
			desc:      "nil src and nil dst",
			src:       nil,
			dst:       nil,
			want:      nil,
			wantError: false,
		}, {
			desc:      "valid msg and nil target",
			src:       &wrp.Message{Type: 11},
			dst:       nil,
			want:      nil,
			wantError: true,
		}, {
			desc: "msg type matches exact type",
			src: &wrp.Message{
				Type:            wrp.SimpleRequestResponseMessageType,
				Source:          "mac:112233445566",
				Destination:     "event:device-status",
				TransactionUUID: "12345678-1234-1234-1234-123456789012",
			},
			dst: &wrp.SimpleRequestResponse{},
			want: &wrp.SimpleRequestResponse{
				Source:          "mac:112233445566",
				Destination:     "event:device-status",
				TransactionUUID: "12345678-1234-1234-1234-123456789012",
			},
			Processor: wrp.NoStandardValidation(),
		}, {
			desc: "msg type matches exact type, inverse",
			src: &wrp.SimpleRequestResponse{
				Source:          "mac:112233445566",
				Destination:     "event:device-status",
				TransactionUUID: "12345678-1234-1234-1234-123456789012",
			},
			dst: &wrp.Message{},
			want: &wrp.Message{
				Type:            wrp.SimpleRequestResponseMessageType,
				Source:          "mac:112233445566",
				Destination:     "event:device-status",
				TransactionUUID: "12345678-1234-1234-1234-123456789012",
			},
			Processor: wrp.NoStandardValidation(),
		}, {
			desc:      "msg type is not the same type",
			src:       &wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
			dst:       &wrp.SimpleEvent{},
			wantError: true,
		}, {
			desc:      "msg type is not the same type, inverse",
			src:       &wrp.SimpleEvent{},
			dst:       &wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
			wantError: true,
		}, {
			desc:      "invalid message type",
			src:       &wrp.Message{Type: wrp.MessageType(999)},
			dst:       &wrp.SimpleRequestResponse{},
			wantError: true,
		}, {
			desc: "from and to the same type ... or basically a clone with validation",
			src: &wrp.Authorization{
				Status: 200,
			},
			dst: &wrp.Authorization{},
			want: &wrp.Authorization{
				Status: 200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			err := wrp.As(tt.src, tt.dst, tt.Processor)
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.want == nil {
				return
			}

			var want wrp.Message
			tt.want.To(&want)

			var got wrp.Message
			tt.dst.To(&got)

			assert.Equal(t, want, got)
		})
	}
}
