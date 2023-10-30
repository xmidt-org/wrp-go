// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpcontext

import (
	"context"
)

type contextContentsKey struct{}

// Set provides a standard way to add a wrp message to a context.Context. This supports not only wrp.Message
// but also all the other message types, such as wrp.SimpleRequestResponse
func SetContents(ctx context.Context, b []byte) context.Context {
	return context.WithValue(ctx, contextContentsKey{}, b)
}

// Get a message from a context and return it as type T
func GetContents(ctx context.Context) ([]byte, bool) {
	src := ctx.Value(contextContentsKey{})
	if src == nil {
		return []byte{}, false
	}

	return get[[]byte](ctx, src)
}
