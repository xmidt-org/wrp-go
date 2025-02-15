// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/wrp-go/v4"
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
			target: &wrp.Message{Type: 11},
			want:   false,
		}, {
			desc:   "non-nil msg, type is invalid",
			msg:    &wrp.Message{Type: 0},
			target: &wrp.Message{},
			want:   false,
		}, {
			desc:   "non-nil msg and nil target",
			msg:    &wrp.Message{Type: 11},
			target: nil,
			want:   false,
		}, {
			desc:   "same msg and target type",
			msg:    &wrp.Message{Type: 11},
			target: &wrp.Message{Type: 11},
			want:   true,
		}, {
			desc:   "different msg and target type",
			msg:    &wrp.Message{Type: 3},
			target: &wrp.Message{Type: 4},
			want:   false,
		}, {
			desc:   "msg type matches exact type",
			msg:    &wrp.Message{Type: 11},
			target: &wrp.Unknown{},
			want:   true,
		}, {
			desc:   "msg type matches exact type, inverse",
			msg:    &wrp.Unknown{},
			target: &wrp.Message{Type: 11},
			want:   true,
		}, {
			desc:   "msg type is not the same type",
			msg:    &wrp.Message{Type: 11},
			target: &wrp.SimpleEvent{},
			want:   false,
		}, {
			desc:   "msg type is not the same type, inverse",
			msg:    &wrp.SimpleEvent{},
			target: &wrp.Message{Type: 11},
			want:   false,
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
		msg       wrp.Union
		target    wrp.Union
		want      wrp.Union
		Processor wrp.Processor
		wantError bool
	}{
		{
			desc: "msg type matches exact type",
			msg: &wrp.Message{
				Type:            wrp.SimpleRequestResponseMessageType,
				Source:          "mac:112233445566",
				Destination:     "event:device-status",
				TransactionUUID: "12345678-1234-1234-1234-123456789012",
			},
			target: &wrp.SimpleRequestResponse{},
			want: &wrp.SimpleRequestResponse{
				Source:          "mac:112233445566",
				Destination:     "event:device-status",
				TransactionUUID: "12345678-1234-1234-1234-123456789012",
			},
			Processor: wrp.NoStandardValidation(),
		},
		{
			desc: "msg type matches exact type, inverse",
			msg: &wrp.SimpleRequestResponse{
				Source:          "mac:112233445566",
				Destination:     "event:device-status",
				TransactionUUID: "12345678-1234-1234-1234-123456789012",
			},
			target: &wrp.Message{},
			want: &wrp.Message{
				Type:            wrp.SimpleRequestResponseMessageType,
				Source:          "mac:112233445566",
				Destination:     "event:device-status",
				TransactionUUID: "12345678-1234-1234-1234-123456789012",
			},
			Processor: wrp.NoStandardValidation(),
		},
		/*
			{
				desc:      "msg type is not the same type",
				msg:       &wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
				target:    &wrp.SimpleEvent{},
				wantError: true,
			},
			{
				desc:      "msg type is not the same type, inverse",
				msg:       &wrp.SimpleEvent{},
				target:    &wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
				wantError: true,
			},
			{
				desc:      "invalid message type",
				msg:       &wrp.Message{Type: wrp.MessageType(999)},
				target:    &wrp.SimpleRequestResponse{},
				wantError: true,
			},
		*/
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			err := wrp.As(tt.msg, tt.target, tt.Processor)
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			var want wrp.Message
			tt.want.To(&want)

			var got wrp.Message
			tt.target.To(&got)

			assert.Equal(t, want, got)
		})
	}
}
