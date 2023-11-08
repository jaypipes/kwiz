// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package errors

import (
	"errors"
)

var (
	// RuntimeError is the base error class for all errors occurring during
	// runtime (and not during the parsing of a config or processing arguments)
	RuntimeError = errors.New("runtime error")
)
