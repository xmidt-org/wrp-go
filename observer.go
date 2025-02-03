// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"context"
	"errors"
)

var (
	// ErrNotHandled is returned to indicate the message was not handled by the
	// Processor, or Modifier.
	ErrNotHandled = errors.New("message not handled")
)

// Observer interface is used to observe wrp messages.
type Observer interface {
	// ObserveWRP is called to observe a message.
	ObserveWRP(context.Context, Message)
}

// ObserverFunc is a convenience type to define an Observer using a function.
type ObserverFunc func(context.Context, Message)

func (f ObserverFunc) ObserveWRP(ctx context.Context, msg Message) {
	f(ctx, msg)
}

// Processor interface is used to handle wrp messages in a consistent way.
type Processor interface {
	// ProcessWRP is called to handle a message.  The return value indicates the
	// outcome of processing the message.
	//
	// The returned error value can be:
	//	- nil indicates the message was handled successfully.
	//	- ErrNotHandled indicates the message was not handled.
	//	- Any other error indicates the message was handled, but there was an error.
	//
	// The caller shall inspect the error using errors.Is(err, ErrNotHandled) to
	// determine if the message was not handled.  This ensures that the method
	// can return an error without ambiguity.
	ProcessWRP(context.Context, Message) error
}

// ProcessorFunc is a convenience type to define a Processor using a function.
type ProcessorFunc func(context.Context, Message) error

func (f ProcessorFunc) ProcessWRP(ctx context.Context, msg Message) error {
	return f(ctx, msg)
}

// Modifier interface is used to optionally modify a message and return the
// modified message.
type Modifier interface {
	// ModifyWRP is called to optionally modify a message.  The return value is
	// the modified message and an error.
	//
	// The returned error value can be:
	//	- nil indicates the message was handled successfully.
	//	- ErrNotHandled indicates the message was not handled.
	//	- Any other error indicates the message was handled, but there was an error.
	//
	// The caller shall inspect the error using errors.Is(err, ErrNotHandled) to
	// determine if the message was not handled.  This ensures that the method
	// can return an error without ambiguity.
	//
	// If the message was not handled, the message value is returned unmodified.
	ModifyWRP(context.Context, Message) (Message, error)
}

// ModifierFunc is a convenience type to define a Modifier using a function.
type ModifierFunc func(context.Context, Message) (Message, error)

func (f ModifierFunc) ModifyWRP(ctx context.Context, msg Message) (Message, error) {
	return f(ctx, msg)
}

// ObserverAsProcessor returns a Processor that wraps an Observer.
//
// This allows an Observer to be used as a Processor, which might be useful
// in such applications as logging or metrics where the message is observed
// but not modified.
//
// The Processor will always return ErrNotHandled to indicate that the message
// was not handled.
func ObserverAsProcessor(o Observer) Processor {
	return ProcessorFunc(func(ctx context.Context, msg Message) error {
		o.ObserveWRP(ctx, msg)
		return ErrNotHandled
	})
}

// ObserverAsModifier returns a Modifier that wraps an Observer.
//
// This allows an Observer to be used as a Modifier, which might be useful
// in such applications as logging or metrics where the message is observed
// but not modified.
//
// The Processor will always return ErrNotHandled to indicate that the message
// was not handled.  The original message is returned.
func ObserverAsModifier(o Observer) Modifier {
	return ModifierFunc(func(ctx context.Context, msg Message) (Message, error) {
		o.ObserveWRP(ctx, msg)
		return msg, ErrNotHandled
	})
}

// ProcessorAsModifier returns a Modifier that wraps a Processor.
//
// This allows a Processor to be used as a Modifier and not need to modify
// the message.  The error value is used to indicate if the message was
// handled or not is returned to the caller.
//
// The Processor will always return ErrNotHandled to indicate that the message
// was not handled.  The original message is returned.
func ProcessorAsModifier(p Processor) Modifier {
	return ModifierFunc(func(ctx context.Context, msg Message) (Message, error) {
		return msg, p.ProcessWRP(ctx, msg)
	})
}

// Observers is a collection of Observers that can be used to observe a message.
type Observers []Observer

// ObserveWRP iterates over the Observers, sequentially calling each Observer of
// the message.
func (o Observers) ObserveWRP(ctx context.Context, msg Message) {
	for _, obs := range o {
		if ctx.Err() != nil {
			return
		}

		if obs == nil {
			continue
		}
		obs.ObserveWRP(ctx, msg)
	}
}

// Processors is a collection of Processors that can be used to process a message.
type Processors []Processor

// ProcessWRP iterates over the Processors, sequentially calling each Processor
// of the message.  The first Processor to return an error that is not ErrNotHandled
// will stop the iteration and return the error.  If all Processors return ErrNotHandled,
// then ErrNotHandled is returned.  If the context is canceled, the iteration stops
// and the error value is returned.
func (p Processors) ProcessWRP(ctx context.Context, msg Message) error {
	e := ErrNotHandled
	for _, proc := range p {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if proc == nil {
			continue
		}

		if err := proc.ProcessWRP(ctx, msg); err != nil {
			if errors.Is(err, ErrNotHandled) {
				continue
			}
			return err
		}
		e = nil
	}
	return e
}

// Modifiers is a collection of Modifiers that can be used to modify a message.
type Modifiers []Modifier

// ModifyWRP iterates over the Modifiers, sequentially applying each Modifier
// to the message.  The first Modifier to return an error that is not ErrNotHandled
// will stop the iteration and return the error.  The modified message prior to
// the error is returned. If all Modifiers return ErrNotHandled, then the
// latest version of the message is returned along with ErrNotHandled.  If the
// context is canceled, the iteration stops and the modified message prior to
// the context being closed is returned.
func (m Modifiers) ModifyWRP(ctx context.Context, msg Message) (Message, error) {
	e := ErrNotHandled
	for _, mod := range m {
		if ctx.Err() != nil {
			return msg, ctx.Err()
		}

		if mod == nil {
			continue
		}

		next, err := mod.ModifyWRP(ctx, msg)
		if err != nil {
			if errors.Is(err, ErrNotHandled) {
				msg = next
				continue
			}
			return msg, err
		}
		e = nil
		msg = next
	}
	return msg, e
}
