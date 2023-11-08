// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package types

// NUMACell represents a single NUMA node/cell within a host. The host will
// typically be a baremetal machine, however a virtual machine may be
// configured to emulate multiple NUMA cells.
type NUMACell struct {
	// Resources contains the capacity, reserved amount and used amount of
	// various system resources in this NUMACell
	Resources Resources
}
