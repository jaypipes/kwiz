// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package node

import (
	"context"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kconnect "github.com/jaypipes/kwiz/pkg/kube/connect"
	"github.com/jaypipes/kwiz/pkg/types"
	"github.com/jaypipes/kwiz/pkg/unit"
)

var (
	nodeGVK = schema.GroupVersionKind{
		Kind: "Node",
	}
)

// Get returns a slice of `Node` objects contained in a Kubernetes cluster.
func Get(
	ctx context.Context,
	c *kconnect.Connection,
) ([]*types.Node, error) {
	gvrNode, err := c.GVR(nodeGVK)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	list, err := c.Client().Resource(gvrNode).List(
		ctx, opts,
	)
	if err != nil {
		return nil, err
	}
	nodes := make([]*types.Node, len(list.Items))
	for x, obj := range list.Items {
		cpuCap, err := resourceCapacityFromRaw(obj.Object, "cpu")
		if err != nil {
			return nil, err
		}
		cpuAlloc, err := resourceAllocatableFromRaw(obj.Object, "cpu")
		if err != nil {
			return nil, err
		}
		cpuReserved := cpuCap - cpuAlloc
		memCap, err := resourceCapacityFromRaw(obj.Object, "memory")
		if err != nil {
			return nil, err
		}
		memAlloc, err := resourceAllocatableFromRaw(obj.Object, "memory")
		if err != nil {
			return nil, err
		}
		memReserved := memCap - memAlloc
		podCap, err := resourceCapacityFromRaw(obj.Object, "pods")
		if err != nil {
			return nil, err
		}
		podAlloc, err := resourceAllocatableFromRaw(obj.Object, "pods")
		if err != nil {
			return nil, err
		}
		podReserved := podCap - podAlloc
		nodeRes := types.Resources{
			CPU: types.ResourceAmounts{
				Capacity:    cpuCap,
				Allocatable: cpuAlloc,
				Reserved:    cpuReserved,
			},
			Memory: types.ResourceAmounts{
				Capacity:    memCap,
				Allocatable: memAlloc,
				Reserved:    memReserved,
			},
			Pods: types.ResourceAmounts{
				Capacity:    podCap,
				Allocatable: podAlloc,
				Reserved:    podReserved,
			},
		}
		name, _, _ := unstructured.NestedString(obj.Object, "metadata", "name")
		node := &types.Node{
			Cluster:   "default",
			Name:      name,
			Resources: nodeRes,
			NUMACells: []types.NUMACell{},
		}
		nodes[x] = node
	}
	return nodes, nil
}

// resourceCapacityFromRaw accepts a raw map of Kubernetes object fields and
// returns the capacity of a requested resource type.
func resourceCapacityFromRaw(
	obj map[string]interface{},
	resType string,
) (float64, error) {
	return resourceAmountFromRaw(obj, "capacity", resType)
}

// resourceAllocatableFromRaw accepts a raw map of Kubernetes object fields and
// returns the allocatable amount of a requested resource type.
func resourceAllocatableFromRaw(
	obj map[string]interface{},
	resType string,
) (float64, error) {
	return resourceAmountFromRaw(obj, "allocatable", resType)
}

// resourceAmountFromRaw accepts a raw map of Kubernetes object fields and
// returns the amount of a category (capacity, allocatable, etc) of a requested
// resource type.
func resourceAmountFromRaw(
	obj map[string]interface{},
	category string,
	resType string,
) (float64, error) {
	amountStr, _, err := unstructured.NestedString(
		obj, "status", category, resType,
	)
	if err != nil {
		return 0, err
	}
	if resType == "memory" {
		// We need to convert any size strings for memory...
		return unit.SizeStringToBytes(amountStr), nil
	}
	amountInt, err := strconv.Atoi(amountStr)
	if err != nil {
		return 0, err
	}
	return float64(amountInt), nil
}
