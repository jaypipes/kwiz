// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package types

// Node represents a Kubernetes node in the cluster
type Node struct {
	// Cluster is the name of the Kubernetes cluster
	Cluster string
	// Name is the name of the Kubernetes node
	Name string
	// Resources contains the capacity, reserved amount and used amount of
	// various system resources on the Node. If the Node is representing a
	// machine with multiple NUMA cells, Resources contains ALL resources,
	// regardless of NUMA cell.
	Resources Resources
	// NUMACells contains the NUMACell structs for each NUMA node/cell in the
	// host machine.
	NUMACells []NUMACell
}
