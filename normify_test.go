// SPDX-FileCopyrightText: 2024 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormifier(t *testing.T) {
	tests := []struct {
		description string
		opt         NormifierOption
		opts        []NormifierOption
		msg         Message
		want        Message
		wantFn      func(*assert.Assertions, *Message)
		expectedErr error
	}{
		{
			description: "empty is ok",
		}, {
			description: "ReplaceAnySelfLocator(mac:112233445566)",
			opt:         ReplaceAnySelfLocator("mac:112233445566"),
			msg: Message{
				Source:      "self:/service/ignored",
				Destination: "self:/place/ignored",
			},
			want: Message{
				Source:      "mac:112233445566/service/ignored",
				Destination: "mac:112233445566/place/ignored",
			},
		}, {
			description: "EnsureTranactionUUID(), new UUID",
			opt:         EnsureTransactionUUID(),
			wantFn: func(assert *assert.Assertions, m *Message) {
				assert.NotEmpty(m.TransactionUUID)
			},
		}, {
			description: "EnsureTranactionUUID() with existing UUID",
			opt:         EnsureTransactionUUID(),
			msg: Message{
				TransactionUUID: "123e4567-e89b-12d3-a456-426614174000",
			},
			want: Message{
				TransactionUUID: "123e4567-e89b-12d3-a456-426614174000",
			},
		}, {
			description: "EnsurePartnerID(partner) appending it to empty list",
			opt:         EnsurePartnerID("partner"),
			want: Message{
				PartnerIDs: []string{"partner"},
			},
		}, {
			description: "EnsurePartnerID(partner) appending it",
			opt:         EnsurePartnerID("partner"),
			msg: Message{
				PartnerIDs: []string{"mouse"},
			},
			want: Message{
				PartnerIDs: []string{"mouse", "partner"},
			},
		}, {
			description: "EnsurePartnerID(partner') existing",
			opt:         EnsurePartnerID("partner"),
			msg: Message{
				PartnerIDs: []string{"mouse", "partner", "cats"},
			},
			want: Message{
				PartnerIDs: []string{"mouse", "partner", "cats"},
			},
		}, {
			description: "SetPartnerID(partner)",
			opt:         SetPartnerID("partner"),
			msg: Message{
				PartnerIDs: []string{"mouse"},
			},
			want: Message{
				PartnerIDs: []string{"partner"},
			},
		}, {
			description: "SetSessionID(session)",
			opt:         SetSessionID("session"),
			want: Message{
				SessionID: "session",
			},
		}, {
			description: "ClampQualityOfService(), QualityOfService < 0",
			opt:         ClampQualityOfService(),
			msg: Message{
				QualityOfService: -1,
			},
			want: Message{
				QualityOfService: 0,
			},
		}, {
			description: "ClampQualityOfService(), QualityOfService > 99",
			opt:         ClampQualityOfService(),
			msg: Message{
				QualityOfService: 100,
			},
			want: Message{
				QualityOfService: 99,
			},
		}, {
			description: "EnsureMetadataString(key, value) add to empty",
			opt:         EnsureMetadataString("key", "value"),
			want: Message{
				Metadata: map[string]string{
					"key": "value",
				},
			},
		}, {
			description: "EnsureMetadataString(key, value) overwrite",
			opt:         EnsureMetadataString("key", "value"),
			msg: Message{
				Metadata: map[string]string{
					"key": "something",
				},
			},
			want: Message{
				Metadata: map[string]string{
					"key": "value",
				},
			},
		}, {
			description: "EnsureMetadataTime(key, 2022-01-01 00:00:00 +0000 UTC)",
			opt:         EnsureMetadataTime("key", time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			want: Message{
				Metadata: map[string]string{
					"key": "2022-01-01T00:00:00Z",
				},
			},
		}, {
			description: "EnsureMetadataInt64(key, 99)",
			opt:         EnsureMetadataInt64("key", 99),
			want: Message{
				Metadata: map[string]string{
					"key": "99",
				},
			},
		}, {
			description: "ValidateSource()",
			opt:         ValidateSource(),
			msg: Message{
				Source: "mac:112233445566/place/ignored",
			},
			want: Message{
				Source: "mac:112233445566/place/ignored",
			},
		}, {
			description: "ValidateDestination()",
			opt:         ValidateDestination(),
			msg: Message{
				Destination: "mac:112233445566/place/ignored",
			},
			want: Message{
				Destination: "mac:112233445566/place/ignored",
			},
		}, {
			description: "ValidateMessageType()",
			opt:         ValidateMessageType(),
			msg: Message{
				Type: SimpleEventMessageType,
			},
			want: Message{
				Type: SimpleEventMessageType,
			},
		}, {
			description: "ValidateOnlyUTF8Strings()",
			opt:         ValidateOnlyUTF8Strings(),
			msg: Message{
				ContentType: "text/plain",
			},
			want: Message{
				ContentType: "text/plain",
			},
		}, {
			description: "ValidateIsPartner(partner)",
			opt:         ValidateIsPartner("partner"),
			msg: Message{
				PartnerIDs: []string{"partner"},
			},
			want: Message{
				PartnerIDs: []string{"partner"},
			},
		}, {
			description: "ValidateHasPartner(cat, partner)",
			opt:         ValidateHasPartner("cat", "partner"),
			msg: Message{
				PartnerIDs: []string{"bob", "partner"},
			},
			want: Message{
				PartnerIDs: []string{"bob", "partner"},
			},
		}, {
			description: "ValidateHasPartner('', partner)",
			opt:         ValidateHasPartner("", "partner"),
			msg: Message{
				PartnerIDs: []string{"bob", "partner"},
			},
			want: Message{
				PartnerIDs: []string{"bob", "partner"},
			},
		},

		// Negative test cases
		{
			description: "ReplaceAnySelfLocator()",
			opt:         ReplaceAnySelfLocator(""),
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "ReplaceDestinationSelfLocator()",
			opt:         ReplaceDestinationSelfLocator(""),
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "ReplaceDestinationSelfLocator(mac:112233445566)",
			opt:         ReplaceDestinationSelfLocator("mac:112233445566"),
			msg: Message{
				Destination: "invalid:/place/ignored",
			},
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "ReplaceSourceSelfLocator()",
			opt:         ReplaceSourceSelfLocator(""),
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "ReplaceSourceSelfLocator(mac:112233445566)",
			opt:         ReplaceSourceSelfLocator("mac:112233445566"),
			msg: Message{
				Source: "invalid:/place/ignored",
			},
			expectedErr: ErrorInvalidLocator,
		}, {
			description: "ValidateSource() empty",
			opt:         ValidateSource(),
			expectedErr: ErrInvalidSource,
		}, {
			description: "ValidateSource() invalid",
			opt:         ValidateSource(),
			msg: Message{
				Source: "mac:invalid/place/ignored",
			},
			expectedErr: ErrInvalidSource,
		}, {
			description: "ValidateDestination() empty",
			opt:         ValidateDestination(),
			expectedErr: ErrInvalidDest,
		}, {
			description: "ValidateDestination() invalid",
			opt:         ValidateDestination(),
			msg: Message{
				Destination: "mac:invalid/place/ignored",
			},
			expectedErr: ErrInvalidDest,
		}, {
			description: "ValidateMessageType(), invalid as 0",
			opt:         ValidateMessageType(),
			expectedErr: ErrInvalidMessageType,
		}, {
			description: "ValidateMessageType(), invalid as really large",
			opt:         ValidateMessageType(),
			msg: Message{
				Type: MessageType(999999999999999999),
			},
			expectedErr: ErrInvalidMessageType,
		}, {
			description: "ValidateOnlyUTF8Strings() invalid",
			opt:         ValidateOnlyUTF8Strings(),
			msg: Message{
				ContentType: string([]byte{0xbf}),
			},
			expectedErr: ErrNotUTF8,
		}, {
			description: "ValidateIsPartner(partner), empty list",
			opt:         ValidateIsPartner("partner"),
			expectedErr: ErrInvalidPartnerID,
		}, {
			description: "ValidateIsPartner(partner), not there",
			opt:         ValidateIsPartner("partner"),
			msg: Message{
				PartnerIDs: []string{"mouse"},
			},
			expectedErr: ErrInvalidPartnerID,
		}, {
			description: "ValidateHasPartner('', partner) not there",
			opt:         ValidateHasPartner("", "partner"),
			msg: Message{
				PartnerIDs: []string{"bob", "cats"},
			},
			expectedErr: ErrInvalidPartnerID,
		}, {
			description: "ValidateHasPartner('', partner) not there, empty list",
			opt:         ValidateHasPartner("", "partner"),
			expectedErr: ErrInvalidPartnerID,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			opts := append(tc.opts, tc.opt)
			n := NewNormifier(opts...)
			require.NotNil(n)

			m := tc.msg
			err := n.Normify(&m)

			assert.ErrorIs(err, tc.expectedErr)
			if tc.expectedErr != nil {
				return
			}

			if tc.wantFn != nil {
				tc.wantFn(assert, &m)
			} else {
				assert.Equal(tc.want, m)
			}
		})
	}
}
