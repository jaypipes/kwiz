// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package errors

import (
	"fmt"

	kwerrors "github.com/jaypipes/kwiz/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// ErrResourceUnknown is returned when an unknown resource kind is
	// specified. This is a runtime error because we rely on the discovery
	// client to determine whether a resource kind is valid.
	ErrResourceUnknown = fmt.Errorf(
		"%w: resource unknown",
		kwerrors.RuntimeError,
	)
)

// ResourceUnknown returns ErrRuntimeResourceUnknown for a given kind
func ResourceUnknown(gvk schema.GroupVersionKind) error {
	return fmt.Errorf("%w: %s", ErrResourceUnknown, gvk)
}
