/**
 * Copyright 2021 Comcast Cable Communications Management, LLC
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
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/xmidt-org/wrp-go/v3"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type XmidtResponse struct {

	//Code is the HTTP Status code received from the transaction
	Code int

	//ForwardedHeaders contains all the headers tr1d1um keeps from the transaction
	ForwardedHeaders http.Header

	//Body represents the full data off the XMiDT http.Response body
	Body []byte
}

type Client struct {

	// URL is the full location for the serverside wrp endpoint.
	// If unset, use talaria's URI at localhost, which the port used in talaria's docker image
	URL string

	// RequestFormat would be the wrp Format to use for all requests, which specifies the wrp.Encoder.
	// If unset, defaults to JSON, which is what the wrp package defaults to.
	RequestFormat wrp.Format

	// If unset, defaults to net/http.DefaultClient
	HTTPClient HTTPClient

	WrpMsg wrp.Message

	authHeaderValue string
}

func (c *Client) SendWRP(ctx context.Context) (result *XmidtResponse, err error) {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{}
	}
	// (1) create an *http.Request, using c.RequestFormat to marshal the body and the client URL
	var payload []byte
	err = wrp.NewEncoderBytes(&payload, c.RequestFormat).Encode(c.WrpMsg)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	// (2) use c.HTTPClient or http.DefaultClient to execute the HTTP transaction
	var response *http.Response
	if response, err = c.HTTPClient.Do(request.WithContext(ctx)); err == nil {

		// (3) translate the response using the wrp package and the response as the target of unmarshaling
		result = &XmidtResponse{
			ForwardedHeaders: make(http.Header),
			Body:             []byte{},
		}

		result.Code = response.StatusCode

		defer response.Body.Close()

		result.Body, err = ioutil.ReadAll(response.Body)
		return
	}

	return result, nil
}
