// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp

import (
	"context"
	"errors"
)

// NoStandardValidation returns a Processor that prevents standard validation.
func NoStandardValidation() Processor {
	return noStandardValidation{}
}

type noStandardValidation struct{}

func (noStandardValidation) ProcessWRP(context.Context, Message) error {
	return nil
}

// StdValidator returns a Processor that validates messages based on their type.
// If the message type is not recognized, it will return ErrInvalidMessageType.
// If the message type is recognized, it will return an error if the message is
// invalid, or nil if the message is valid.
func StdValidator() Processor {
	return ProcessorFunc(func(ctx context.Context, msg Message) error {
		err := ErrInvalidMessageType

		p := mtValidatorMap[msg.Type]
		if p != nil {
			err = p.ProcessWRP(ctx, msg)
			if errors.Is(err, ErrNotHandled) {
				return ErrInvalidMessageType
			}
		}
		return err
	})
}
