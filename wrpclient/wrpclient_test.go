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
	"github.com/xmidt-org/wrp-go/v3"
)

func TestCheckClientConfig(t *testing.T) {
	defaultClient := Client{
		URL:           "http://localhost:6200",
		HTTPClient:    &http.Client{},
		RequestFormat: 1,
	}
	tcs := []struct {
		desc           string
		client         Client
		expectedClient Client
	}{
		{
			desc:           "Empty Client",
			client:         Client{},
			expectedClient: defaultClient,
		},
		{
			desc: "Happy Input Client",
			client: Client{
				URL: "url",
				HTTPClient: &http.Client{
					Timeout: 2,
				},
				RequestFormat: 2,
			},
			expectedClient: Client{
				URL: "url",
				HTTPClient: &http.Client{
					Timeout: 2,
				},
				RequestFormat: 2,
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			assert := assert.New(t)
			tc.client.checkClientConfig()
			assert.EqualValues(tc.client, tc.expectedClient)
		})
	}
}

func TestSendWRP(t *testing.T) {

	tcs := []struct {
		desc        string
		client      Client
		response    interface{}
		request     interface{}
		nilContext  bool
		expectedErr error
	}{
		{
			desc: "Invalid RequestFormat failure",
			client: Client{
				RequestFormat: 8,
			},
			response:    wrp.Message{},
			expectedErr: errInvalidRequestFormat,
		},
		{
			desc: "Non 200 Response failure",
			client: Client{
				HTTPClient: &mockHTTPClientFailureCode{},
			},
			expectedErr: errNonSucessfulResponse,
		},
		{
			desc:        "Request Creation failure",
			nilContext:  true,
			expectedErr: errCreateRequest,
		},
		{
			desc:        "Encode failure",
			request:     wrp.Message{},
			expectedErr: errEncoding,
		},
		{
			desc: "HTTPClient Transaction failure",
			client: Client{
				HTTPClient: &mockHTTPClientReturnErr{},
			},
			expectedErr: errHTTPTransaction,
		},
		{
			desc: "Decode failure",
			client: Client{
				HTTPClient: &mockHTTPClientBodyFailure{},
			},
			response:    &wrp.Message{},
			request:     &wrp.Message{},
			expectedErr: errDecoding,
		},
		{
			desc: "Happy Client and Path success",
			client: Client{
				HTTPClient: &mockHTTPClientSuccess{},
			},
			response:    &wrp.Message{},
			request:     &wrp.Message{},
			expectedErr: nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			assert := assert.New(t)
			var ctx context.Context
			if tc.nilContext {
				ctx = nil
			} else {
				ctx = context.TODO()
			}
			err := tc.client.SendWRP(ctx, &tc.response, &tc.request)
			assert.True(errors.Is(err, tc.expectedErr),
				fmt.Errorf("error [%v] doesn't contain error [%v] in its err chain",
					err, tc.expectedErr),
			)
		})
	}
}
