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
	"errors"
	"fmt"
	"net/http"

	"github.com/xmidt-org/httpaux/erraux"
	"github.com/xmidt-org/wrp-go/v3"
)

var (
	errEncoding              = errors.New("encoding error")
	errCreateRequest         = errors.New("http request creation error")
	errHTTPTransaction       = errors.New("http transaction error")
	errDecoding              = errors.New("decoding response error")
	errNonSuccessfulResponse = errors.New("non-200 response")
	errInvalidRequestFormat  = errors.New("invalid client RequestFormat")
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {

	// URL is the full location for the serverside wrp endpoint.
	// If unset, use talaria's URI at localhost, which is the port used in talaria's docker image
	URL string

	// RequestFormat is the wrp Format to use for all requests, which specifies the wrp.Encoder.
	// If unset, defaults to JSON, which is what the wrp package defaults to.
	RequestFormat wrp.Format

	// If unset, defaults to net/http.DefaultClient
	HTTPClient HTTPClient
}

func (c *Client) checkClientConfig() error {
	if c.URL == "" {
		c.URL = "http://localhost:6200"
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{}
	}
	if c.RequestFormat == 0 {
		c.RequestFormat = wrp.JSON
	} else if c.RequestFormat > 2 || c.RequestFormat < 0 {
		return errInvalidRequestFormat
	}
	return nil
}

func (c *Client) SendWRP(ctx context.Context, response, request interface{}) error {
	err := c.checkClientConfig()
	if err != nil {
		return err
	}
	// Create an *http.Request, using c.RequestFormat to marshal the body and the client URL
	var payload []byte
	err = wrp.NewEncoderBytes(&payload, c.RequestFormat).Encode(request)
	if err != nil {
		return fmt.Errorf("%w: %v", errEncoding, err)
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("%w: %v", errCreateRequest, err)
	}

	// Use c.HTTPClient or http.DefaultClient to execute the HTTP transaction
	resp, err := c.HTTPClient.Do(r.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("%w: %v", errHTTPTransaction, err)
	} else if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		err := &erraux.Error{
			Err:     err,
			Code:    resp.StatusCode,
			Message: resp.Status,
			Header:  resp.Header,
		}
		return fmt.Errorf("%w: %v", errNonSuccessfulResponse, err)
	}

	// Translate the response using the wrp package and the response as the target of unmarshaling
	defer resp.Body.Close()
	err = wrp.NewDecoder(resp.Body, c.RequestFormat).Decode(response)
	if err != nil {
		return fmt.Errorf("%w: %v", errDecoding, err)
	}

	return nil
}
