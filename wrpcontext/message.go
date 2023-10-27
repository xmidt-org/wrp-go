// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpcontext

import (
	"context"

	"github.com/xmidt-org/wrp-go/v3"
)

type contextWRPMessageKey struct{}

// Set provides a standard way to add a wrp message to a context.Context. This supports not only wrp.Message
// but also all the other message types, such as wrp.SimpleRequestResponse
func SetMessage(ctx context.Context, msg any) context.Context {
	return context.WithValue(ctx, contextWRPMessageKey{}, msg)
}

// Get a message from a context and return it as type T
func GetMessage(ctx context.Context) (*wrp.Message, bool) {
	src := ctx.Value(contextWRPMessageKey{})
	if src == nil {
		return nil, false
	}

	return get[*wrp.Message](ctx, src)
}
