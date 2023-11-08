// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package types

// Resources contains the capacity, reserved amount and used amount of various
// system resources on the provider of resources (either Node or NUMA cell)
type Resources struct {
	// CPU contains CPU resource amounts
	CPU ResourceAmounts
	// Memory contains RAM resource amounts
	Memory ResourceAmounts
	// Pods contains the amounts of Pod resources
	Pods ResourceAmounts
}

// ResourceAmounts contains a single resource's capacity, reserved amount and
// used amount.
type ResourceAmounts struct {
	// Capacity is the total amount of this resource
	Capacity float64
	// Allocatable is the amount of this resource that may be allocated to
	// consumers
	Allocatable float64
	// Reserved is the amount of this resource reserved for the system
	Reserved float64
	// RequestedFloor is the floor amount of this resource that has been
	// requested by consumers
	RequestedFloor float64
	// RequestedCeiling is the maximum amount of this resource that has been
	// requested by consumers
	RequestedCeiling float64
	// Used is the reported actual amount of this resource being actively
	// consumed (includes system usage)
	Used float64
}

// ResourceRequests contains the floor and ceiling requests of various system
// resources by a single consumer (Pod)
type ResourceRequests struct {
	// CPU contains CPU resource request
	CPU ResourceRequest
	// Memory contains RAM resource request
	Memory ResourceRequest
}

// ResourceRequests contains the floor and ceiling request for a particular
// resource
type ResourceRequest struct {
	// Floor is the floor amount of this resource that has been requested by
	// the consumer.
	Floor float64
	// Ceiling is the max/ceiling amount of this resource that has been
	// requested by the consumer. -1.0 means there is no ceiling.
	Ceiling float64
}
