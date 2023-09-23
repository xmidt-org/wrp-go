// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package wrphttp

import "errors"

type httpError struct {
	err  error
	code int
}

func (e httpError) Error() string {
	return e.err.Error()
}

func (e httpError) StatusCode() int {
	return e.code
}

// Is reports whether any error in e.err's chain matches target.
func (e httpError) Is(target error) bool {
	return errors.Is(e.err, target)
}
