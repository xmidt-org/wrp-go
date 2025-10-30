// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrp_test

import (
	"fmt"
	"strings"

	"github.com/xmidt-org/wrp-go/v5"
)

type Extended struct {
	wrp.SimpleEvent
	Extra string
}

var _ wrp.Union = (*Extended)(nil)

// Implement the wrp.Union interface for the Extended type.
func (e *Extended) MsgType() wrp.MessageType {
	return wrp.SimpleEventMessageType
}

// Implement the wrp.Union interface for the Extended type.
func (e *Extended) To(msg *wrp.Message, v ...wrp.Processor) error {
	e.Headers = []string{"extra: " + e.Extra}
	return e.SimpleEvent.To(msg, v...)
}

// Implement the wrp.Union interface for the Extended type.
func (e *Extended) From(msg *wrp.Message, v ...wrp.Processor) error {
	err := e.SimpleEvent.From(msg, v...)
	if err != nil {
		return err
	}

	e.Extra = ""
	if len(e.Headers) > 0 {
		after, found := strings.CutPrefix(e.Headers[0], "extra: ")
		if found {
			e.Extra = after
		}
	}

	return nil
}

// Implement the wrp.Union interface for the Extended type.
func (e *Extended) Validate(v ...wrp.Processor) error {
	return e.SimpleEvent.Validate(v...)
}

// Is reports whether the msg is the same type as the target, or is convertible
// to the target, without the need to validate the message.  This function is not
// part of the wrp.Union interface, but it is a useful function to have when working
// with the wrp.Union interface.
func (e *Extended) Is(msg wrp.Union) bool {
	var tmp wrp.SimpleEvent
	if wrp.As(msg, &tmp, wrp.NoStandardValidation()) == nil {
		if len(tmp.Headers) > 0 && strings.HasPrefix(tmp.Headers[0], "extra: ") {
			return true
		}
	}

	return false
}

func Example_extend() {
	isExtended := wrp.SimpleEvent{
		Headers: []string{"extra: something"},
	}
	isNotExtended := wrp.SimpleEvent{}

	// The Extended type is a wrp.Union, so it can be used in the wrp.Is function.
	// The wrp.Is function is useful for determining if the type matches, even if
	// the message doesn't validate.
	fmt.Printf("isExtended:    %t\n", wrp.Is(&isExtended, &Extended{}))
	fmt.Printf("isNotExtended: %t\n", wrp.Is(&isNotExtended, &Extended{}))

	// Output: isExtended:    true
	// isNotExtended: false
}
