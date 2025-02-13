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
		msg        *wrp.Message
		target     any
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
			target: &wrp.Message{Type: 3},
			want:   false,
		}, {
			desc:   "non-nil msg, type is invalid",
			msg:    &wrp.Message{Type: 0},
			target: &wrp.Message{},
			want:   false,
		}, {
			desc:   "non-nil msg and nil target",
			msg:    &wrp.Message{Type: 3},
			target: nil,
			want:   false,
		}, {
			desc:   "same msg and target type",
			msg:    &wrp.Message{Type: 3},
			target: &wrp.Message{Type: 3},
			want:   true,
		}, {
			desc:   "different msg and target type",
			msg:    &wrp.Message{Type: 3},
			target: &wrp.Message{Type: 4},
			want:   true,
		}, {
			desc:   "msg type matches exact type",
			msg:    &wrp.Message{Type: 3},
			target: &wrp.Message{},
			want:   true,
		}, {
			desc:   "msg type matches exact type",
			msg:    &wrp.Message{Type: 3},
			target: &wrp.SimpleRequestResponse{},
			want:   true,
		}, {
			desc:   "msg type is not the same type",
			msg:    &wrp.Message{Type: 3},
			target: &wrp.SimpleEvent{},
			want:   false,
		}, {
			desc: "msg type matches a field in the target struct",
			msg:  &wrp.Message{Type: 3},
			target: struct {
				foo string
				Bar wrp.Message
			}{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := wrp.Is(tt.msg, tt.target, tt.validators...)
			assert.Equal(t, tt.want, got)
		})
	}
}
