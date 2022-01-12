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

	"github.com/go-kit/kit/endpoint"
)

// New constructs a go-kit endpoint for the given WRP service.  This endpoint enforces
// the constraint that ctx must be the context associated with the Request.
func New(s Service) endpoint.Endpoint {
	return func(ctx context.Context, value interface{}) (interface{}, error) {
		return s.ServeWRP(ctx, value.(Request))
	}
}

// Wrap does the opposite of New: it takes a go-kit endpoint and returns a Service
// that invokes it.
func Wrap(e endpoint.Endpoint) Service {
	return ServiceFunc(func(ctx context.Context, request Request) (Response, error) {
		response, err := e(ctx, request)
		return response.(Response), err
	})
}
