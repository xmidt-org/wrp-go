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

package wrpendpoint

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceFunc(t *testing.T) {
	var (
		assert = assert.New(t)

		request Request = &request{
			note: note{
				contents: []byte("expected request"),
			},
		}

		expectedResponse Response = &response{
			note: note{
				contents: []byte("expected response"),
			},
		}

		expectedCtx = context.WithValue(context.Background(), foo, "bar")

		serviceFuncCalled = false

		serviceFunc = ServiceFunc(func(ctx context.Context, r Request) (Response, error) {
			serviceFuncCalled = true
			assert.Equal(expectedCtx, ctx)
			return expectedResponse, nil
		})
	)

	actualResponse, err := serviceFunc.ServeWRP(expectedCtx, request)
	assert.Equal(expectedResponse, actualResponse)
	assert.NoError(err)
	assert.True(serviceFuncCalled)
}
