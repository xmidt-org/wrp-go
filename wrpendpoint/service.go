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
)

// Service represents a component which processes WRP transactions.
type Service interface {
	// ServeWRP processes a WRP request.  Either the Response or the error
	// returned from this method will be nil, but not both.
	ServeWRP(context.Context, Request) (Response, error)
}

// ServiceFunc is a function type that implements Service
type ServiceFunc func(context.Context, Request) (Response, error)

func (sf ServiceFunc) ServeWRP(ctx context.Context, r Request) (Response, error) {
	return sf(ctx, r)
}
