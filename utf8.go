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

package wrp

import (
	"strconv"
	"strings"
	"unicode"
)

type utf8Fixer struct{}

// WriteExt converts a value to a []byte.
//
// Note: v is a pointer iff the registered extension type is a struct or array kind.
func (f utf8Fixer) WriteExt(v interface{}) []byte {
	return []byte(
		strings.ToValidUTF8(
			*v.(*string), // this works due to how we register
			strconv.QuoteRune(unicode.ReplacementChar), // may have to convert this to a string, not sure
		),
	)
}

// ReadExt updates a value from a []byte.
//
// Note: dst is always a pointer kind to the registered extension type.
func (f utf8Fixer) ReadExt(dst interface{}, src []byte) {
	*(dst.(*string)) = strings.ToValidUTF8(string(src), strconv.QuoteRune(unicode.ReplacementChar))
}
