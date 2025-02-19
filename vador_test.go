// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"context"
	"testing"
)

func TestVadorProcessWRP(t *testing.T) {
	tests := []struct {
		name    string
		msg     Message
		v       vador
		wantErr bool
	}{
		{
			name: "valid message",
			msg: Message{
				Type:        SimpleEventMessageType,
				Source:      "dns:example.com",
				Destination: "mac:FFEEDDCCBBAA",
			},
			v: vador{
				Type:        SimpleEventMessageType,
				Source:      required,
				Destination: required,
			},
			wantErr: false,
		}, {
			name: "optional locator",
			msg: Message{
				Type:        SimpleEventMessageType,
				Source:      "dns:example.com",
				Destination: "mac:FFEEDDCCBBAA",
			},
			v: vador{
				Type:        SimpleEventMessageType,
				Source:      required,
				Destination: optional,
			},
			wantErr: false,
		}, {
			name: "utf8 problem for a locator",
			msg: Message{
				Type:   SimpleEventMessageType,
				Source: string([]byte{0xff, 0xfe, 0xfd}),
			},
			v: vador{
				Type:   SimpleEventMessageType,
				Source: required,
			},
			wantErr: true,
		}, {
			name: "invalid message, types don't match",
			msg: Message{
				Type: SimpleEventMessageType,
			},
			v: vador{
				Type: SimpleRequestResponseMessageType,
			},
			wantErr: true,
		}, {
			name: "invalid message, required payload not present",
			msg: Message{
				Type: SimpleEventMessageType,
			},
			v: vador{
				Type:    SimpleEventMessageType,
				Payload: required,
			},
			wantErr: true,
		}, {
			name: "invalid message, required strings not present",
			msg: Message{
				Type: SimpleEventMessageType,
			},
			v: vador{
				Type:    SimpleEventMessageType,
				Headers: required,
			},
			wantErr: true,
		}, {
			name: "invalid message, required strings empty",
			msg: Message{
				Type:    SimpleEventMessageType,
				Headers: []string{},
			},
			v: vador{
				Type:    SimpleEventMessageType,
				Headers: required,
			},
			wantErr: true,
		}, {
			name: "invalid message, required strings utf8 problem",
			msg: Message{
				Type:    SimpleEventMessageType,
				Headers: []string{string([]byte{0xff, 0xfe, 0xfd})},
			},
			v: vador{
				Type:    SimpleEventMessageType,
				Headers: required,
			},
			wantErr: true,
		}, {
			name: "invalid message, required metadata utf8 problem",
			msg: Message{
				Type: SimpleEventMessageType,
				Metadata: map[string]string{
					"foo": string([]byte{0xff, 0xfe, 0xfd}),
				},
			},
			v: vador{
				Type:     SimpleEventMessageType,
				Metadata: required,
			},
			wantErr: true,
		}, {
			name: "invalid message, required metadata missing",
			msg: Message{
				Type: SimpleEventMessageType,
			},
			v: vador{
				Type:     SimpleEventMessageType,
				Metadata: required,
			},
			wantErr: true,
		}, {
			name: "invalid message, required metadata empty",
			msg: Message{
				Type:     SimpleEventMessageType,
				Metadata: map[string]string{},
			},
			v: vador{
				Type:     SimpleEventMessageType,
				Metadata: required,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.ProcessWRP(context.Background(), tt.msg); (err != nil) != tt.wantErr {
				t.Errorf("vador.ProcessWRP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
