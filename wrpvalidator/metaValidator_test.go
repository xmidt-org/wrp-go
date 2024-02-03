// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpvalidator

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/sallust"
	"github.com/xmidt-org/touchstone"
	"github.com/xmidt-org/wrp-go/v3"
)

func TestMetaValidatorUnmarshal(t *testing.T) {
	tests := []struct {
		description string
		config      []byte
		expectedErr error
	}{
		{
			description: "Unmarshalling success",
			config: []byte(`{
				"type": "utf8",
				"level": "warning",
				"disable": true
			}`),
		},
		{
			description: "Empty configuration unmarshalling success",
		},
		{
			description: "Json unmarshalling error",
			config:      []byte(`[]`),
			expectedErr: ErrValidatorUnmarshalling,
		},
		{
			description: "Unknown validator type unmarshalling failure",
			config: []byte(`{
				"type": "FOOBAR",
				"level": "warning",
				"disable": true
			}`),
			expectedErr: ErrValidatorUnmarshalling,
		},
		{
			description: "Unknown validator type unmarshalling failure",
			config: []byte(`{
				"type": "unknown",
				"level": "warning",
				"disable": true
			}`),
			expectedErr: ErrValidatorUnmarshalling,
		},
		{
			description: "Default validator type unmarshalling failure",
			config: []byte(`{
				"level": "warning",
				"disable": true
			}`),
			expectedErr: ErrValidatorUnmarshalling,
		},
		{
			description: "Invalid validator configuration failure",
			config: []byte(`{
				"type": "utf8",
				"disable": true
			}`),
			expectedErr: ErrValidatorInvalidConfig,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			var valMeta ValidatorWithMetadata
			assert := assert.New(t)

			err := valMeta.UnmarshalJSON(tc.config)
			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
			} else {
				assert.NoError(err)
				if len(tc.config) != 0 {
					assert.False(valMeta.IsEmpty())
					assert.True(valMeta.IsValid())
					assert.True(valMeta.Disabled())
					assert.Equal(UTF8Type, valMeta.Type())
					assert.Equal(WarningLevel, valMeta.Level())
				} else {
					assert.True(valMeta.IsEmpty())
					assert.False(valMeta.IsValid())
					assert.False(valMeta.Disabled())
					assert.Equal(UnknownType, valMeta.Type())
					assert.Equal(UnknownLevel, valMeta.Level())
				}
			}
		})
	}
}

func TestMetaValidatorAddMetric(t *testing.T) {
	tests := []struct {
		description string
		config      []byte
		expectedErr error
	}{
		{
			description: "Add metric validator always_valid",
			config: []byte(`[
				{
					"type": "always_valid",
					"level": "warning"
				}
			]`),
		},
		{
			description: "Add metric validator utf8",
			config: []byte(`[
				{
					"type": "utf8",
					"level": "warning"
				}
			]`),
		},
		{
			description: "Add metric validator msg_type",
			config: []byte(`[
				{
					"type": "msg_type",
					"level": "warning"
				}
			]`),
		},
		{
			description: "Add metric validator source",
			config: []byte(`[
				{
					"type": "source",
					"level": "warning"
				}
			]`),
		},
		{
			description: "Add metric validator destination",
			config: []byte(`[
				{
					"type": "destination",
					"level": "warning"
				}
			]`),
		},
		{
			description: "Add metric validator simple_res_req",
			config: []byte(`[
				{
					"type": "simple_res_req",
					"level": "warning"
				}
			]`),
		},
		{
			description: "Add metric validator simple_event",
			config: []byte(`[
				{
					"type": "simple_event",
					"level": "warning"
				}
			]`),
		},
		{
			description: "Add metric validator spans",
			config: []byte(`[
				{
					"type": "spans",
					"level": "warning"
				}
			]`),
		},
		{
			description: "Add metric validator always_invalid",
			config: []byte(`[
				{
					"type": "always_invalid",
					"level": "warning"
				}
			]`),
		},
		{
			description: "Disabled validator success",
			config: []byte(`[
				{
					"type": "utf8",
					"level": "warning",
					"disable": true
				}
			]`),
		},
		{
			description: "Invalid validators failure",
			expectedErr: ErrValidatorInvalidConfig,
		},
		{
			description: "Duplicate validators failure",
			config: []byte(`[
				{
					"type": "utf8",
					"level": "warning"
				},
				{
					"type": "utf8",
					"level": "warning"
				}
			]`),
			expectedErr: ErrValidatorAddMetric,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			var valMeta []ValidatorWithMetadata
			assert := assert.New(t)
			require := require.New(t)
			if len(tc.config) != 0 {
				require.NoError(json.Unmarshal(tc.config, &valMeta))
			} else {
				valMeta = append(valMeta, ValidatorWithMetadata{})
			}

			cfg := touchstone.Config{
				DefaultNamespace: "n",
				DefaultSubsystem: "s",
			}
			_, pr, err := touchstone.New(cfg)
			require.NoError(err)

			tf := touchstone.NewFactory(cfg, sallust.Default(), pr)
			if len(valMeta) < 2 {
				err := valMeta[0].AddMetric(tf)
				if tc.expectedErr != nil {
					assert.ErrorIs(err, tc.expectedErr)
				} else {
					assert.NoError(err)
				}
			} else {
				if tc.expectedErr != nil {
					assert.NoError(valMeta[0].AddMetric(tf))
					assert.ErrorIs(valMeta[1].AddMetric(tf), tc.expectedErr)
				} else {
					assert.NoError(errors.New("Unknown test state"))
				}
			}

		})
	}
}

func TestMetaValidatorValidate(t *testing.T) {
	tests := []struct {
		description string
		config      []byte
		msg         wrp.Message
		expectedErr error
	}{
		{
			description: "Disabled validate success",
			config: []byte(`[
				{
					"type": "utf8",
					"level": "warning",
					"disable": true
				}
			]`),
			msg: wrp.Message{Destination: "MAC:\xed\xbf\xbf"},
		},
		{
			description: "Validate success validator always_valid",
			config: []byte(`[
				{
					"type": "always_valid",
					"level": "warning"
				}
			]`),
			msg: wrp.Message{},
		},
		{
			description: "Validate success validator utf8",
			config: []byte(`[
				{
					"type": "utf8",
					"level": "warning"
				}
			]`),
			msg: wrp.Message{Destination: "MAC:11:22:33:44:55:66"},
		},
		{
			description: "Validate success validator msg_type",
			config: []byte(`[
				{
					"type": "msg_type",
					"level": "warning"
				}
			]`),
			msg: wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
		},
		{
			description: "Validate success validator source",
			config: []byte(`[
				{
					"type": "source",
					"level": "warning"
				}
			]`),
			msg: wrp.Message{Source: "MAC:11:22:33:44:55:66"},
		},
		{
			description: "Validate success validator destination",
			config: []byte(`[
				{
					"type": "destination",
					"level": "warning"
				}
			]`),
			msg: wrp.Message{Destination: "MAC:11:22:33:44:55:66"},
		},
		{
			description: "Validate success validator simple_res_req",
			config: []byte(`[
				{
					"type": "simple_res_req",
					"level": "warning"
				}
			]`),
			msg: wrp.Message{Type: wrp.SimpleRequestResponseMessageType},
		},
		{
			description: "Validate success validator simple_event",
			config: []byte(`[
				{
					"type": "simple_event",
					"level": "warning"
				}
			]`),
			msg: wrp.Message{Type: wrp.SimpleEventMessageType},
		},
		{
			description: "Validate success validator spans",
			config: []byte(`[
				{
					"type": "spans",
					"level": "warning"
				}
			]`),
			msg: wrp.Message{Spans: [][]string{{"parent", "name", "1234", "1234", "1234"}}},
		},
		{
			description: "Validate failure validator always_invalid",
			config: []byte(`[
				{
					"type": "always_invalid",
					"level": "warning"
				}
			]`),
			expectedErr: ErrorInvalidMsgType.Err,
		},
		{
			description: "Invalid validators failure",
			msg:         wrp.Message{},
			expectedErr: ErrValidatorInvalidConfig,
		},
		{
			description: "Not UTF8 validate failure",
			config: []byte(`[
				{
					"type": "utf8",
					"level": "warning"
				}
			]`),
			msg: wrp.Message{
				// Not UFT8 Destination string
				Destination: "MAC:\xed\xbf\xbf",
			},
			expectedErr: ErrorInvalidMessageEncoding.Err,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			var valMeta []ValidatorWithMetadata
			assert := assert.New(t)
			require := require.New(t)
			if len(tc.config) != 0 {
				require.NoError(json.Unmarshal(tc.config, &valMeta))
			} else {
				valMeta = append(valMeta, ValidatorWithMetadata{})
			}

			require.Len(valMeta, 1)

			err := valMeta[0].Validate(tc.msg, prometheus.Labels{})
			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
			} else {
				assert.NoError(err)
			}
		})
	}
}
