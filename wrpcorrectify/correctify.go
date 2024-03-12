// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrpcorrectify

import (
	"context"

	"github.com/xmidt-org/wrp-go/v3"
)

// Correctifier applies a series of normalizing options to a WRP message.
type Correctifier struct {
	opts []Option
}

// Option is a functional option for normalizing a WRP message.
type Option interface {
	// Correctify applies the option to the given message.
	Correctify(context.Context, *wrp.Message) error
}

// OptionFunc is an adapter to allow the use of ordinary functions as
// normalizing options.
type OptionFunc func(context.Context, *wrp.Message) error

var _ Option = OptionFunc(nil)

func (f OptionFunc) Correctify(ctx context.Context, m *wrp.Message) error {
	return f(ctx, m)
}

// New creates a new Normalizer with the given options.
func New(opts ...Option) *Correctifier {
	return &Correctifier{
		opts: opts,
	}
}

// Correctify applies the normalizing options to the message or returns an error
// if any of the options fail.
func (n *Correctifier) Correctify(ctx context.Context, m *wrp.Message) error {
	for _, opt := range n.opts {
		if opt != nil {
			if err := opt.Correctify(ctx, m); err != nil {
				return err
			}
		}
	}
	return nil
}
