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
	"bytes"
	"context"
	"net/http"

	"github.com/xmidt-org/httpaux/erraux"
	"github.com/xmidt-org/wrp-go/v3"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
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
}

func (c *Client) SendWRP(ctx context.Context, response, request interface{}) error {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{}
	}
	// (1) create an *http.Request, using c.RequestFormat to marshal the body and the client URL
	var payload []byte
	err := wrp.NewEncoderBytes(&payload, c.RequestFormat).Encode(request)
	if err != nil {
		return err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	// (2) use c.HTTPClient or http.DefaultClient to execute the HTTP transaction
	resp, err := c.HTTPClient.Do(r.WithContext(ctx))
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return &erraux.Error{
			Err:     err,
			Code:    resp.StatusCode,
			Message: resp.Status,
			Header:  resp.Header,
		}
	} else if err != nil {
		return err
	}

	// (3) translate the response using the wrp package and the response as the target of unmarshaling

	defer resp.Body.Close()
	err = wrp.NewDecoder(resp.Body, c.RequestFormat).Decode(response)
	if err != nil {
		return err
	}

	return nil
}
