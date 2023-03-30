/**
 * Copyright 2023 Comcast Cable Communications Management, LLC
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

package wrpcontext

import (
	"context"
	"encoding/json"
)

type contextKey string

func (c contextKey) String() string {
	return "wrpcontext context key " + string(c)
}

var (
	contextKeyWrpMessage = contextKey("wrp-message")
)

// Set provides a standard way to add a wrp message to a context.Context. This supports not only wrp.Message
// but also all the other message types, such as wrp.SimpleRequestResponse
func Set(ctx context.Context, msg any) context.Context {
	return context.WithValue(ctx, contextKeyWrpMessage, msg)
}

// Get a wrp.Message from a context, nil means no value was associated with the key in the context
func Get(ctx context.Context) any {
	return ctx.Value(contextKeyWrpMessage)
}

// Get a message from a context and store it in the value pointed to by msg
func GetAs(ctx context.Context, msg any) bool {
	msgVal := Get(ctx)
	if msgVal == nil {
		return false
	}
	jsonBytes, err := json.Marshal(msgVal)
	if err != nil {
		return false
	}
	if err := json.Unmarshal(jsonBytes, msg); err != nil {
		return false
	}
	return msg != nil
}
