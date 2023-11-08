// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package types

// Pod represents a Kubernetes pod
type Pod struct {
	// Cluster is the name of the Kubernetes cluster
	Cluster string
	// Node is the name of the Kubernetes node the Pod is on
	Node string
	// Namespace is the Kubernetes namesapce the Pod is in
	Namespace string
	// Name is the name of the Pod
	Name string
	// ResourceRequests contains the floor and ceiling amounts of resources
	// requested by all containers in the Pod
	ResourceRequests ResourceRequests
}
