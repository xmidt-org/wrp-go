/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package wrpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/wrp-go/v3"
)

func TestNew(t *testing.T) {
	defaultClient := Client{
		url:           "http://localhost:6200",
		httpClient:    &http.Client{},
		requestFormat: 1,
	}
	tcs := []struct {
		desc           string
		client         Client
		expectedClient Client
		expectedErr    error
	}{
		{
			desc: "Happy Input Client",
			client: Client{
				url: "http://random.com",
				httpClient: &http.Client{
					Timeout: 2,
				},
				requestFormat: 2,
			},
			expectedClient: Client{
				url: "http://random.com",
				httpClient: &http.Client{
					Timeout: 2,
				},
				requestFormat: 2,
			},
		},
		{
			desc:           "Empty Client",
			client:         Client{},
			expectedClient: defaultClient,
		},

		{
			desc: "Invalid requestFormat failure",
			client: Client{
				requestFormat: 8,
			},
			expectedErr: errInvalidRequestFormat,
		},
		{
			desc: "Invalid url failure",
			client: Client{
				url: "nope",
			},
			expectedErr: errInvalidURL,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			assert := assert.New(t)
			client, err := New(tc.client.url, tc.client.requestFormat, tc.client.httpClient)
			if tc.expectedErr != nil {
				assert.True(errors.Is(err, tc.expectedErr),
					fmt.Errorf("error [%v] doesn't contain error [%v] in its err chain",
						err, tc.expectedErr),
				)
			} else {
				require.NoError(t, err)
				assert.EqualValues(tc.expectedClient, *client)
			}

		})
	}
}

func TestSendWRP(t *testing.T) {
	simpleMessage := &wrp.Message{
		Type: wrp.SimpleRequestResponseMessageType,
	}

	tcs := []struct {
		desc              string
		client            Client
		response          interface{}
		request           interface{}
		nilContext        bool
		expectedErr       error
		HTTPReturnCode    int
		HTTPPayload       interface{}
		useMockHTTPClient bool
	}{
		{
			desc:              "Non 200 Response failure",
			useMockHTTPClient: true,
			expectedErr:       errNonSuccessfulResponse,
			HTTPReturnCode:    400,
			HTTPPayload:       simpleMessage,
		},
		{
			desc:        "Request Creation failure",
			nilContext:  true,
			expectedErr: errCreateRequest,
		},
		{
			desc:        "HTTPClient Transaction failure",
			expectedErr: errHTTPTransaction,
		},
		{
			desc:              "Decode failure",
			useMockHTTPClient: true,
			HTTPReturnCode:    200,
			HTTPPayload:       "",
			response:          &wrp.Message{},
			request:           &wrp.Message{},
			expectedErr:       errDecoding,
		},
		{
			desc:              "Happy Client and Path success",
			useMockHTTPClient: true,
			HTTPReturnCode:    200,
			HTTPPayload:       simpleMessage,
			response:          &wrp.Message{},
			request:           &wrp.Message{},
			expectedErr:       nil,
		},
	}

	for _, tc := range tcs {

		t.Run(tc.desc, func(t *testing.T) {
			assert := assert.New(t)

			var ctx context.Context
			if tc.nilContext {
				ctx = nil
			} else {
				ctx = context.Background()
			}

			var payload []byte
			err := wrp.NewEncoderBytes(&payload, wrp.JSON).Encode(tc.HTTPPayload)
			require.NoError(t, err)
			m := new(mockHTTPClient)
			m.On("Do", mock.AnythingOfType("*http.Request")).Return(tc.HTTPReturnCode, payload)
			client, err := New("", wrp.JSON, nil)
			require.NoError(t, err)
			if tc.useMockHTTPClient {
				client.httpClient = m
			}

			err = client.SendWRP(ctx, &tc.response, &tc.request)
			if tc.useMockHTTPClient {
				m.AssertExpectations(t)
			}

			assert.True(errors.Is(err, tc.expectedErr),
				fmt.Errorf("error [%v] doesn't contain error [%v] in its err chain",
					err, tc.expectedErr),
			)
		})
	}
}
