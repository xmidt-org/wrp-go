// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObservers_ObserveWRP(t *testing.T) {
	var a, b int

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		desc      string
		ctx       context.Context
		observers Observers
		a         int
		b         int
	}{
		{
			desc: "simple",
			observers: Observers{
				ObserverFunc(func(_ context.Context, _ Message) {
					a++
				}),
				nil,
				ObserverFunc(func(_ context.Context, _ Message) {
					b++
				}),
			},
			a: 1,
			b: 1,
		}, {
			desc: "canceled context",
			ctx:  cancelledCtx,
			observers: Observers{
				ObserverFunc(func(_ context.Context, _ Message) {
					a++
				}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			if tc.ctx != nil {
				ctx = tc.ctx
			}

			a = 0
			b = 0

			tc.observers.ObserveWRP(ctx, Message{})

			assert.Equal(t, tc.a, a)
			assert.Equal(t, tc.b, b)
		})
	}
}

func TestProcessors_ProcessWRP(t *testing.T) {
	var a, b int

	unknownErr := errors.New("unknown error")

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		desc       string
		ctx        context.Context
		processors Processors
		a          int
		b          int
		err        error
	}{
		{
			desc: "simple processed",
			processors: Processors{
				ObserverAsProcessor(ObserverFunc(func(_ context.Context, _ Message) {
					a++
				})),
				nil,
				ProcessorFunc(func(_ context.Context, _ Message) error {
					b++
					return nil
				}),
			},
			a: 1,
			b: 1,
		}, {
			desc: "canceled context",
			ctx:  cancelledCtx,
			processors: Processors{
				ProcessorFunc(func(_ context.Context, _ Message) error {
					a++
					return nil
				}),
			},
			err: context.Canceled,
		}, {
			desc: "simple not handled",
			processors: Processors{
				ProcessorFunc(func(_ context.Context, _ Message) error {
					a++
					return ErrNotHandled
				}),
			},
			a:   1,
			err: ErrNotHandled,
		}, {
			desc: "error encoutered",
			processors: Processors{
				ProcessorFunc(func(_ context.Context, _ Message) error {
					a++
					return unknownErr
				}),
				ObserverAsProcessor(ObserverFunc(func(_ context.Context, _ Message) {
					b++
				})),
			},
			a:   1,
			err: unknownErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			if tc.ctx != nil {
				ctx = tc.ctx
			}

			a = 0
			b = 0

			err := tc.processors.ProcessWRP(ctx, Message{})

			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.a, a)
			assert.Equal(t, tc.b, b)
		})
	}
}

func TestModifiers_ModifyWRP(t *testing.T) {
	var a, b, c int

	unknownErr := errors.New("unknown error")

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		desc      string
		ctx       context.Context
		modifiers Modifiers
		a         int
		b         int
		c         int
		want      Message
		err       error
	}{
		{
			desc: "simple modified",
			modifiers: Modifiers{
				ObserverAsModifier(ObserverFunc(func(_ context.Context, _ Message) {
					a++
				})),
				nil,
				ModifierFunc(func(_ context.Context, _ Message) (Message, error) {
					b++
					return Message{
						Accept: "anything",
					}, nil
				}),
				ModifierFunc(func(_ context.Context, _ Message) (Message, error) {
					c++
					return Message{
						ContentType: MimeTypeJson,
					}, nil
				}),
				ProcessorAsModifier(ProcessorFunc(func(_ context.Context, m Message) error {
					if m.Accept != "anything" &&
						m.ContentType != MimeTypeJson {
						return unknownErr
					}
					return ErrNotHandled
				})),
			},
			a: 1,
			b: 1,
			c: 1,
			want: Message{
				Accept:      "anything",
				ContentType: MimeTypeJson,
			},
		}, {
			desc: "canceled context",
			ctx:  cancelledCtx,
			modifiers: Modifiers{
				ModifierFunc(func(_ context.Context, _ Message) (Message, error) {
					a++
					return Message{}, nil
				}),
			},
			err: context.Canceled,
		}, {
			desc: "simple not handled",
			modifiers: Modifiers{
				ModifierFunc(func(_ context.Context, _ Message) (Message, error) {
					a++
					return Message{}, ErrNotHandled
				}),
			},
			a:   1,
			err: ErrNotHandled,
		}, {
			desc: "error encoutered",
			modifiers: Modifiers{
				ModifierFunc(func(_ context.Context, _ Message) (Message, error) {
					a++
					return Message{}, unknownErr
				}),
				ObserverAsModifier(ObserverFunc(func(_ context.Context, _ Message) {
					b++
				})),
			},
			a:   1,
			err: unknownErr,
		}, {
			desc: "observers as modifiers",
			modifiers: Modifiers{
				ObserverAsModifier(Observers{
					ObserverFunc(func(_ context.Context, _ Message) {
						a++
					}),
					ObserverFunc(func(_ context.Context, _ Message) {
						b++
					}),
				}),
			},
			a:   1,
			b:   1,
			err: ErrNotHandled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			if tc.ctx != nil {
				ctx = tc.ctx
			}

			a = 0
			b = 0

			_, err := tc.modifiers.ModifyWRP(ctx, Message{})

			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.a, a)
			assert.Equal(t, tc.b, b)
		})
	}
}
