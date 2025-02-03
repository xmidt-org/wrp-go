// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestModifiers(t *testing.T) {
	tests := []struct {
		desc    string
		value   Message
		vador   Modifier
		want    Message
		altWant func(Message) bool
		err     error // Set to nil if ErrNotHandled is expected.
	}{
		// ValidateSelfDestinationLocator()
		{
			desc: "valid self destination",
			value: Message{
				Destination: "self:/service/ignored",
			},
			vador: ReplaceSelfDestinationLocator(Locator{
				Scheme:    "mac",
				Authority: "112233445566",
			}),
			want: Message{
				Destination: "mac:112233445566/service/ignored",
			},
		}, {
			desc: "valid non-self destination",
			value: Message{
				Destination: "mac:665544332211/service/ignored",
			},
			vador: ReplaceSelfDestinationLocator(Locator{
				Scheme:    "mac",
				Authority: "112233445566",
			}),
			want: Message{
				Destination: "mac:665544332211/service/ignored",
			},
		}, {
			desc: "invalid self destination",
			value: Message{
				Destination: "self",
			},
			vador: ReplaceSelfDestinationLocator(Locator{
				Scheme:    "mac",
				Authority: "112233445566",
			}),
			err: ErrorInvalidLocator,
		}, {
			desc: "empty self destination",
			vador: ReplaceSelfDestinationLocator(Locator{
				Scheme:    "mac",
				Authority: "112233445566",
			}),
		},

		// ValidateSelfSourceLocator()
		{
			desc: "valid self source",
			value: Message{
				Source: "self:/service/ignored",
			},
			vador: ReplaceSelfSourceLocator(Locator{
				Scheme:    "mac",
				Authority: "112233445566",
			}),
			want: Message{
				Source: "mac:112233445566/service/ignored",
			},
		}, {
			desc: "valid non-self source",
			value: Message{
				Source: "mac:665544332211/service/ignored",
			},
			vador: ReplaceSelfSourceLocator(Locator{
				Scheme:    "mac",
				Authority: "112233445566",
			}),
			want: Message{
				Source: "mac:665544332211/service/ignored",
			},
		}, {
			desc: "invalid self source",
			value: Message{
				Source: "self",
			},
			vador: ReplaceSelfSourceLocator(Locator{
				Scheme:    "mac",
				Authority: "112233445566",
			}),
			err: ErrorInvalidLocator,
		}, {
			desc: "empty self source",
			vador: ReplaceSelfSourceLocator(Locator{
				Scheme:    "mac",
				Authority: "112233445566",
			}),
		},

		// ValidateAnySelfLocator()
		{
			desc: "valid self destination",
			value: Message{
				Destination: "self:/service1/ignored",
				Source:      "self:/service2/ignored",
			},
			vador: ReplaceAnySelfLocator(Locator{
				Scheme:    "mac",
				Authority: "112233445566",
			}),
			want: Message{
				Destination: "mac:112233445566/service1/ignored",
				Source:      "mac:112233445566/service2/ignored",
			},
		},

		// EnsureTransactionUUID()
		{
			desc: "ensure transaction UUID when empty",
			value: Message{
				TransactionUUID: "",
			},
			vador: EnsureTransactionUUID(),
			altWant: func(m Message) bool {
				return m.TransactionUUID != ""
			},
		}, {
			desc: "ensure transaction UUID when already set",
			value: Message{
				TransactionUUID: "existing_uuid",
			},
			vador: EnsureTransactionUUID(),
			want: Message{
				TransactionUUID: "existing_uuid",
			},
		},

		// EnsurePartnerID()
		{
			desc:  "ensure partner ID when none are present",
			vador: EnsurePartnerID("partner2"),
			want: Message{
				PartnerIDs: []string{"partner2"},
			},
		}, {
			desc: "ensure partner ID when not present",
			value: Message{
				PartnerIDs: []string{"partner1"},
			},
			vador: EnsurePartnerID("partner2"),
			want: Message{
				PartnerIDs: []string{"partner1", "partner2"},
			},
		}, {
			desc: "ensure partner ID when already present",
			value: Message{
				PartnerIDs: []string{"partner1", "partner2"},
			},
			vador: EnsurePartnerID("partner2"),
			want: Message{
				PartnerIDs: []string{"partner1", "partner2"},
			},
		},

		// SetPartnerID()
		{
			desc: "set single partner ID",
			value: Message{
				PartnerIDs: []string{"partner1", "partner2"},
			},
			vador: SetPartnerID("partner3"),
			want: Message{
				PartnerIDs: []string{"partner3"},
			},
		},

		// SetPartnerIDs()
		{
			desc: "set multiple partner IDs",
			value: Message{
				PartnerIDs: []string{"partner1", "partner2"},
			},
			vador: SetPartnerIDs("partner3", "partner4"),
			want: Message{
				PartnerIDs: []string{"partner3", "partner4"},
			},
		},

		// SetSessionID()
		{
			desc: "set session ID",
			value: Message{
				SessionID: "old_session_id",
			},
			vador: SetSessionID("new_session_id"),
			want: Message{
				SessionID: "new_session_id",
			},
		},

		// ClampQualityOfService()
		{
			desc: "clamp QoS below range",
			value: Message{
				QualityOfService: -1,
			},
			vador: ClampQualityOfService(),
			want: Message{
				QualityOfService: 0,
			},
		}, {
			desc: "clamp QoS above range",
			value: Message{
				QualityOfService: 100,
			},
			vador: ClampQualityOfService(),
			want: Message{
				QualityOfService: 99,
			},
		}, {
			desc: "clamp QoS within range",
			value: Message{
				QualityOfService: 50,
			},
			vador: ClampQualityOfService(),
			want: Message{
				QualityOfService: 50,
			},
		},

		// EnsureMetadata()
		{
			desc:  "ensure metadata, when metadata is nil",
			vador: EnsureMetadata("key2", "value2"),
			want: Message{
				Metadata: map[string]string{"key2": "value2"},
			},
		}, {
			desc:  "ensure metadata bool",
			vador: EnsureMetadata("key2", true),
			want: Message{
				Metadata: map[string]string{"key2": "true"},
			},
		}, {
			desc: "ensure metadata string",
			value: Message{
				Metadata: map[string]string{"key1": "value1"},
			},
			vador: EnsureMetadata("key2", "value2"),
			want: Message{
				Metadata: map[string]string{"key1": "value1", "key2": "value2"},
			},
		}, {
			desc:  "ensure metadata int",
			vador: EnsureMetadata("key2", int(12)),
			want: Message{
				Metadata: map[string]string{"key2": "12"},
			},
		}, {
			desc:  "ensure metadata int8",
			vador: EnsureMetadata("key2", int8(-12)),
			want: Message{
				Metadata: map[string]string{"key2": "-12"},
			},
		}, {
			desc:  "ensure metadata int16",
			vador: EnsureMetadata("key2", int16(-12)),
			want: Message{
				Metadata: map[string]string{"key2": "-12"},
			},
		}, {
			desc:  "ensure metadata int32",
			vador: EnsureMetadata("key2", int32(12)),
			want: Message{
				Metadata: map[string]string{"key2": "12"},
			},
		}, {
			desc:  "ensure metadata int64",
			vador: EnsureMetadata("key2", int64(12)),
			want: Message{
				Metadata: map[string]string{"key2": "12"},
			},
		}, {
			desc:  "ensure metadata uint",
			vador: EnsureMetadata("key2", uint(12)),
			want: Message{
				Metadata: map[string]string{"key2": "12"},
			},
		}, {
			desc:  "ensure metadata uint8",
			vador: EnsureMetadata("key2", uint8(12)),
			want: Message{
				Metadata: map[string]string{"key2": "12"},
			},
		}, {
			desc:  "ensure metadata uint16",
			vador: EnsureMetadata("key2", uint16(12)),
			want: Message{
				Metadata: map[string]string{"key2": "12"},
			},
		}, {
			desc:  "ensure metadata uint32",
			vador: EnsureMetadata("key2", uint32(12)),
			want: Message{
				Metadata: map[string]string{"key2": "12"},
			},
		}, {
			desc:  "ensure metadata uint64",
			vador: EnsureMetadata("key2", uint64(12)),
			want: Message{
				Metadata: map[string]string{"key2": "12"},
			},
		}, {
			desc:  "ensure metadata uintptr",
			vador: EnsureMetadata("key2", uintptr(12)),
			want: Message{
				Metadata: map[string]string{"key2": "12"},
			},
		}, {
			desc:  "ensure metadata float32",
			vador: EnsureMetadata("key2", float32(12.01)),
			want: Message{
				Metadata: map[string]string{"key2": "12.01"},
			},
		}, {
			desc:  "ensure metadata float64",
			vador: EnsureMetadata("key2", float64(12.01)),
			want: Message{
				Metadata: map[string]string{"key2": "12.01"},
			},
		}, {
			desc:  "ensure metadata complex64",
			vador: EnsureMetadata("key2", complex64(12.01)),
			want: Message{
				Metadata: map[string]string{"key2": "(12.01+0i)"},
			},
		}, {
			desc:  "ensure metadata complex128",
			vador: EnsureMetadata("key2", complex128(12.01)),
			want: Message{
				Metadata: map[string]string{"key2": "(12.01+0i)"},
			},
		}, {
			desc:  "ensure metadata time",
			vador: EnsureMetadata("key2", time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)),
			want: Message{
				Metadata: map[string]string{"key2": "2023-01-01T00:00:00Z"},
			},
		}, {
			desc:  "ensure metadata duration",
			vador: EnsureMetadata("key2", time.Second),
			want: Message{
				Metadata: map[string]string{"key2": "1s"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			assert := assert.New(t)
			ctx := context.Background()
			want, err := tc.vador.ModifyWRP(ctx, tc.value)
			assert.Error(err)
			if tc.altWant != nil {
				assert.True(tc.altWant(want))
			} else {
				assert.Equal(tc.want, want)
			}
			if tc.err == nil {
				assert.ErrorIs(err, ErrNotHandled)
				return
			}

			assert.ErrorIs(err, tc.err)
		})
	}
}
