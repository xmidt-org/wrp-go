// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0


package wrpclient

import (
	"bytes"
	"io"
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
		Body:       io.NopCloser(bytes.NewBuffer(args.Get(1).([]byte))),
	}, nil
}
