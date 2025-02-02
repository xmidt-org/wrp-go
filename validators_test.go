// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp_test

import (
	"context"
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/wrp-go/v4"
)

func TestInvalidUtf8Decoding(t *testing.T) {
	assert := assert.New(t)

	/*
		"\x85"  - 5 name value pairs
			"\xa8""msg_type"         : "\x03" // 3
			"\xa4""dest"             : "\xac""\xed\xbf\xbft-address"
			"\xa7""payload"          : "\xc4""\x03" - len 3
											 "123"
			"\xa6""source"           : "\xae""source-address"
			"\xb0""transaction_uuid" : "\xd9\x24""c07ee5e1-70be-444c-a156-097c767ad8aa"
	*/
	invalid := []byte{
		0x85,
		0xa8, 'm', 's', 'g', '_', 't', 'y', 'p', 'e', 0x03,
		0xa4, 'd', 'e', 's', 't', 0xac /* \xed\xbf\xbf is invalid */, 0xed, 0xbf, 0xbf, 't', '-', 'a', 'd', 'd', 'r', 'e', 's', 's',
		0xa7, 'p', 'a', 'y', 'l', 'o', 'a', 'd', 0xc4, 0x03, '1', '2', '3',
		0xa6, 's', 'o', 'u', 'r', 'c', 'e', 0xae, 's', 'o', 'u', 'r', 'c', 'e', '-', 'a', 'd', 'd', 'r', 'e', 's', 's',
		0xb0, 't', 'r', 'a', 'n', 's', 'a', 'c', 't', 'i', 'o', 'n', '_', 'u', 'u', 'i', 'd', 0xd9, 0x24, 'c', '0', '7', 'e', 'e', '5', 'e', '1', '-', '7', '0', 'b', 'e', '-', '4', '4', '4', 'c', '-', 'a', '1', '5', '6', '-', '0', '9', '7', 'c', '7', '6', '7', 'a', 'd', '8', 'a', 'a',
	}

	decoder := wrp.NewDecoderBytes(invalid, wrp.Msgpack)
	msg := new(wrp.Message)
	err := decoder.Decode(msg)
	assert.Nil(err)
	assert.True(utf8.ValidString(msg.Source))

	assert.False(utf8.ValidString(msg.Destination))

	ctx := context.Background()
	err = wrp.ValidateUTF8().ProcessWRP(ctx, *msg)
	assert.ErrorIs(err, wrp.ErrNotUTF8)
}

func TestUTF8(t *testing.T) {
	valid := wrp.Message{
		Source:          "valid string",
		Destination:     "valid string",
		TransactionUUID: "valid string",
		ContentType:     "valid string",
		Accept:          "valid string",
		Headers:         []string{"valid string"},
		Metadata:        map[string]string{"valid": "string"},
		Path:            "valid string",
		PartnerIDs:      []string{"valid string"},
		ServiceName:     "valid string",
		URL:             "valid string",
		SessionID:       "valid string",
	}

	invalid := wrp.Message{
		Source:          string([]byte{0xbf}),
		Destination:     string([]byte{0xbf}),
		TransactionUUID: string([]byte{0xbf}),
		ContentType:     string([]byte{0xbf}),
		Accept:          string([]byte{0xbf}),
		Headers:         []string{string([]byte{0xbf})},
		Metadata:        map[string]string{"invalid": string([]byte{0xbf})},
		Path:            string([]byte{0xbf}),
		PartnerIDs:      []string{string([]byte{0xbf})},
		ServiceName:     string([]byte{0xbf}),
		URL:             string([]byte{0xbf}),
		SessionID:       string([]byte{0xbf}),
	}

	tests := []struct {
		desc   string
		value  wrp.Message
		vador  wrp.Processor
		errHas string
	}{
		{
			desc:  "Success",
			value: valid,
			vador: wrp.ValidateUTF8(),
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8Source(),
			errHas: "source",
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8Destination(),
			errHas: "destination",
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8TransactionUUID(),
			errHas: "transaction_uuid",
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8ContentType(),
			errHas: "content_type",
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8Accept(),
			errHas: "accept",
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8Headers(),
			errHas: "headers",
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8Metadata(),
			errHas: "metadata",
		}, {
			desc: "invalid key ",
			value: wrp.Message{
				Metadata: map[string]string{string([]byte{0xbf}): "invalid"},
			},
			vador:  wrp.ValidateUTF8Metadata(),
			errHas: "metadata",
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8Path(),
			errHas: "path",
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8Partners(),
			errHas: "partner_ids",
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8ServiceName(),
			errHas: "service_name",
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8URL(),
			errHas: "url",
		}, {
			value:  invalid,
			vador:  wrp.ValidateUTF8SessionID(),
			errHas: "session_id",
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc+tc.errHas, func(t *testing.T) {
			assert := assert.New(t)
			ctx := context.Background()
			err := tc.vador.ProcessWRP(ctx, tc.value)
			if tc.errHas != "" {
				assert.Error(err)
				assert.Contains(err.Error(), tc.errHas)
				return
			}
		})
	}
}

type validatorTest struct {
	desc  string
	value wrp.Message
	vador wrp.Processor
	err   error // Set to nil if ErrNotHandled is expected.
}

var partnerTests = []validatorTest{
	// wrp.ValidateHasAPartner()
	{
		desc: "valid partner",
		value: wrp.Message{
			PartnerIDs: []string{"partner"},
		},
		vador: wrp.ValidatePartnerPresent(),
	}, {
		desc:  "empty partner list",
		vador: wrp.ValidatePartnerPresent(),
		err:   wrp.ErrValidationFailed,
	}, {
		desc:  "empty partner list with empty members",
		vador: wrp.ValidatePartnerPresent(),
		value: wrp.Message{
			PartnerIDs: []string{""},
		},
		err: wrp.ErrValidationFailed,
	},

	// wrp.ValidateContainsPartner()
	{
		desc: "contains valid partner",
		value: wrp.Message{
			PartnerIDs: []string{"partner", "other"},
		},
		vador: wrp.ValidatePartnerIsOneOf("some", "partner"),
	}, {
		desc: "contains valid partner",
		value: wrp.Message{
			PartnerIDs: []string{"partner", "other"},
		},
		vador: wrp.ValidatePartnerIsOneOf("missing"),
		err:   wrp.ErrValidationFailed,
	},

	// wrp.ValidatePartnerIs()
	{
		desc: "is partner",
		value: wrp.Message{
			PartnerIDs: []string{"partner"},
		},
		vador: wrp.ValidatePartnerIs("partner"),
	}, {
		desc: "contains valid partner, but isn't exclusive",
		value: wrp.Message{
			PartnerIDs: []string{"partner", "other"},
		},
		vador: wrp.ValidatePartnerIs("partner"),
		err:   wrp.ErrValidationFailed,
	},
}

var messageTypeTests = []validatorTest{
	// wrp.ValidateMessageType()
	{
		desc: "contains valid message type",
		value: wrp.Message{
			Type: wrp.SimpleEventMessageType,
		},
		vador: wrp.ValidateMessageType(),
	}, {
		desc: "contains invalid message type",
		value: wrp.Message{
			Type: 0,
		},
		vador: wrp.ValidateMessageType(),
		err:   wrp.ErrValidationFailed,
	},
	// wrp.ValidateMessageTypeIs()
	{
		desc: "valid message type",
		value: wrp.Message{
			Type: wrp.SimpleEventMessageType,
		},
		vador: wrp.ValidateMessageTypeIs(wrp.SimpleEventMessageType),
	}, {
		desc: "invalid message type",
		value: wrp.Message{
			Type: wrp.SimpleEventMessageType,
		},
		vador: wrp.ValidateMessageTypeIs(wrp.SimpleRequestResponseMessageType),
		err:   wrp.ErrValidationFailed,
	},

	// wrp.ValidateMessageTypeIsOneOf()
	{
		desc: "valid message type",
		value: wrp.Message{
			Type: wrp.SimpleEventMessageType,
		},
		vador: wrp.ValidateMessageTypeIsOneOf(wrp.SimpleEventMessageType, wrp.SimpleRequestResponseMessageType),
	}, {
		desc: "invalid message type",
		value: wrp.Message{
			Type: wrp.SimpleEventMessageType,
		},
		vador: wrp.ValidateMessageTypeIsOneOf(wrp.SimpleRequestResponseMessageType, wrp.CreateMessageType),
		err:   wrp.ErrValidationFailed,
	},
}

var destinationTests = []validatorTest{
	// wrp.ValidateDestination()
	{
		desc: "contains valid destination",
		value: wrp.Message{
			Destination: "mac:112233445566/service",
		},
		vador: wrp.ValidateDestination(),
	}, {
		desc:  "empty destination",
		vador: wrp.ValidateDestination(),
		err:   wrp.ErrValidationFailed,
	}, {
		desc: "contains invalid destination",
		value: wrp.Message{
			Destination: "invalid",
		},
		vador: wrp.ValidateDestination(),
		err:   wrp.ErrorInvalidLocator,
	},
}

var sourceTests = []validatorTest{
	// wrp.ValidateSource()
	{
		desc: "contains valid source",
		value: wrp.Message{
			Source: "mac:112233445566/service",
		},
		vador: wrp.ValidateSource(),
	}, {
		desc:  "empty source",
		vador: wrp.ValidateSource(),
		err:   wrp.ErrValidationFailed,
	}, {
		desc: "contains invalid source",
		value: wrp.Message{
			Source: "invalid",
		},
		vador: wrp.ValidateSource(),
		err:   wrp.ErrorInvalidLocator,
	},
}

var transactionUUIDTests = []validatorTest{
	// wrp.ValidateTransactionUUID()
	{
		desc: "valid transaction uuid",
		value: wrp.Message{
			TransactionUUID: "a6f8711d-6a7e-43ff-b5bf-5f5b58c9f622",
		},
		vador: wrp.ValidateTransactionUUID(),
	}, {
		desc:  "empty transaction uuid",
		vador: wrp.ValidateTransactionUUID(),
		err:   wrp.ErrValidationFailed,
	},
}

var serviceNameTests = []validatorTest{
	// wrp.ValidateServiceName()
	{
		desc: "valid service name",
		value: wrp.Message{
			ServiceName: "service",
		},
		vador: wrp.ValidateServiceName(),
	}, {
		desc:  "empty service name",
		vador: wrp.ValidateServiceName(),
		err:   wrp.ErrValidationFailed,
	},
}

var rdrTests = []validatorTest{
	// wrp.ValidateRequestDeliveryResponse()
	{
		desc: "valid rdr",
		value: wrp.Message{
			RequestDeliveryResponse: i64ptr(200),
		},
		vador: wrp.ValidateRequestDeliveryResponse(),
	}, {
		desc:  "empty rdr",
		vador: wrp.ValidateRequestDeliveryResponse(),
		err:   wrp.ErrValidationFailed,
	},
}

var urlTests = []validatorTest{
	// wrp.ValidateURL()
	{
		desc: "valid url",
		value: wrp.Message{
			URL: "http://example.com",
		},
		vador: wrp.ValidateURL(),
	}, {
		desc:  "empty url",
		vador: wrp.ValidateURL(),
		err:   wrp.ErrValidationFailed,
	},
}

var qosTests = []validatorTest{
	// wrp.ValidateQOS()
	{
		desc: "valid qos",
		value: wrp.Message{
			QualityOfService: 50,
		},
		vador: wrp.ValidateQualityOfService(),
	}, {
		desc: "invalid (too large) qos",
		value: wrp.Message{
			QualityOfService: 1000,
		},
		vador: wrp.ValidateQualityOfService(),
		err:   wrp.ErrValidationFailed,
	}, {
		desc: "invalid (negative) qos",
		value: wrp.Message{
			QualityOfService: -1,
		},
		vador: wrp.ValidateQualityOfService(),
		err:   wrp.ErrValidationFailed,
	},
}

var typeIsAuthorizationTests = []validatorTest{
	// wrp.ValidateTypeIsAuthorization()
	{
		desc: "valid auth type",
		value: wrp.Message{
			Type:   wrp.AuthorizationMessageType,
			Status: i64ptr(200),
		},
		vador: wrp.ValidateTypeIsAuthorization(),
	}, {
		desc: "invalid auth type",
		value: wrp.Message{
			Type: wrp.SimpleEventMessageType,
		},
		vador: wrp.ValidateTypeIsAuthorization(),
		err:   wrp.ErrValidationFailed,
	}, {
		desc: "invalid auth type - missing status",
		value: wrp.Message{
			Type: wrp.AuthorizationMessageType,
		},
		vador: wrp.ValidateTypeIsAuthorization(),
		err:   wrp.ErrValidationFailed,
	},
}

var typeIsSimpleRequestResponseTests = []validatorTest{
	// wrp.ValidateTypeIsRequestResponse()
	{
		desc: "valid type",
		value: wrp.Message{
			Type:            wrp.SimpleRequestResponseMessageType,
			Source:          "mac:112233445566/service",
			Destination:     "mac:112233445566/service",
			TransactionUUID: "a6f8711d-6a7e-43ff-b5bf-5f5b58c9f622",
		},
		vador: wrp.ValidateTypeIsSimpleRequestResponse(),
	}, {
		desc: "invalid type",
		value: wrp.Message{
			Type: wrp.SimpleEventMessageType,
		},
		vador: wrp.ValidateTypeIsSimpleRequestResponse(),
		err:   wrp.ErrValidationFailed,
	},
}

func generateTests[T any](model T, vador wrp.Processor, vadorName string) []validatorTest {
	tests := make([]validatorTest, 0)

	{
		inputs := generateManyValidTestCases(model)

		var count int
		for _, input := range inputs {
			tests = append(tests, validatorTest{
				desc:  fmt.Sprintf("valid %s %d", vadorName, count),
				value: input,
				vador: vador,
			})
			count++
		}
	}

	{
		changed := []string{}

		inputs := generateDisallowedFieldsTestCases(model, &changed)
		var count int
		for i, input := range inputs {
			tests = append(tests, validatorTest{
				desc:  fmt.Sprintf("invalid %s %s %d", vadorName, changed[i], count),
				value: input,
				vador: vador,
				err:   wrp.ErrValidationFailed,
			})
			count++
		}
	}

	return tests
}

func TestValidators(t *testing.T) {
	t.Parallel()

	tests := make([]validatorTest, 0, 100)

	tests = append(tests, messageTypeTests...)
	tests = append(tests, sourceTests...)
	tests = append(tests, destinationTests...)
	tests = append(tests, transactionUUIDTests...)
	tests = append(tests, partnerTests...)
	tests = append(tests, rdrTests...)
	tests = append(tests, serviceNameTests...)
	tests = append(tests, urlTests...)
	tests = append(tests, qosTests...)
	tests = append(tests, typeIsAuthorizationTests...)
	tests = append(tests, typeIsSimpleRequestResponseTests...)
	tests = append(tests, generateTests(
		authorization{},
		wrp.ValidateTypeIsAuthorization(),
		"ValidateTypeIsAuthorization")...)
	tests = append(tests, generateTests(
		simpleRequestResponse{},
		wrp.ValidateTypeIsSimpleRequestResponse(),
		"ValidateTypeIsSimpleRequestResponse")...)
	tests = append(tests, generateTests(
		simpleEvent{},
		wrp.ValidateTypeIsSimpleEvent(),
		"ValidateTypeIsSimpleEvent")...)
	tests = append(tests, generateTests(
		crud{},
		wrp.ValidateTypeIsCRUD(),
		"ValidateTypeIsCRUD")...)
	tests = append(tests, generateTests(
		serviceRegistration{},
		wrp.ValidateTypeIsServiceRegistration(),
		"ValidateTypeIsServiceRegistration")...)
	tests = append(tests, generateTests(
		serviceAlive{},
		wrp.ValidateTypeIsServiceAlive(),
		"ValidateTypeIsServiceAlive")...)
	tests = append(tests, generateTests(
		unknown{},
		wrp.ValidateTypeIsUnknown(),
		"ValidateTypeIsUnknown")...)

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			assert := assert.New(t)
			ctx := context.Background()
			err := tc.vador.ProcessWRP(ctx, tc.value)
			assert.Error(err)
			if tc.err == nil {
				if !assert.ErrorIs(err, wrp.ErrNotHandled, fmt.Sprintf("input: %+v", tc.value)) {
					fmt.Println("uncomment the line below for more help")
					//pp.Println(tc.value)
				}
				return
			}

			if !assert.ErrorIs(err, tc.err, fmt.Sprintf("input: %+v", tc.value)) {
				fmt.Println("uncomment the line below for more help")
				//pp.Println(tc.value)
			}
		})
	}
}

func i64ptr(i int64) *int64 {
	return &i
}
