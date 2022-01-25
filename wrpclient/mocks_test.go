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
	"io/ioutil"
	"net/http"

	"github.com/stretchr/testify/mock"
)

type mockHTTPClient struct {
	mock.Mock
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return &http.Response{
		StatusCode: args.Int(0),
		Body:       ioutil.NopCloser(bytes.NewBuffer(args.Get(1).([]byte))),
	}, nil
}

// type mockHTTPClientSuccess struct{}

// func (m *mockHTTPClientSuccess) Do(_ *http.Request) (*http.Response, error) {
// 	var payload []byte
// 	wrp.NewEncoderBytes(&payload, 1).Encode(&wrp.Message{Type: 4})
// 	return &http.Response{
// 		StatusCode: 200,
// 		Body:       ioutil.NopCloser(bytes.NewBuffer(payload)),
// 	}, nil
// }

// type mockHTTPClientFailureCode struct{}

// func (m *mockHTTPClientFailureCode) Do(_ *http.Request) (*http.Response, error) {
// 	var payload []byte
// 	wrp.NewEncoderBytes(&payload, 1).Encode(&wrp.Message{Type: 4})
// 	return &http.Response{
// 		StatusCode: 400,
// 		Body:       ioutil.NopCloser(bytes.NewBuffer(payload)),
// 	}, nil
// }

// type mockHTTPClientBodyFailure struct{}

// func (m *mockHTTPClientBodyFailure) Do(_ *http.Request) (*http.Response, error) {
// 	return &http.Response{
// 		StatusCode: 200,
// 		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
// 	}, nil
// }

// type mockHTTPClientReturnErr struct{}

// func (m *mockHTTPClientReturnErr) Do(_ *http.Request) (*http.Response, error) {
// 	return &http.Response{}, errors.New("test")
// }
