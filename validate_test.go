// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		msg     Union
		wantErr bool
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
		},
		{
			name:    "empty message",
			msg:     &Message{},
			wantErr: true,
		},
		{
			name:    "valid message",
			msg:     &Message{Type: SimpleEventMessageType, Source: "dns:example.com", Destination: "mac:FFEEDDCCBBAA"},
			wantErr: false,
		},
		{
			name:    "invalid message, missing fields",
			msg:     &Message{Type: SimpleEventMessageType, Source: "dns:example.com"},
			wantErr: true,
		},
		{
			name:    "invalid message, specific, but missing fields",
			msg:     &SimpleEvent{Destination: "mac:FFEEDDCCBBAA"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.msg); (err != nil) != tt.wantErr {
				t.Errorf("Message.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
