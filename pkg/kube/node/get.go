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
	kpod "github.com/jaypipes/kwiz/pkg/kube/pod"
	"github.com/jaypipes/kwiz/pkg/types"
	"github.com/jaypipes/kwiz/pkg/unit"
)

var (
	nodeGVK = schema.GroupVersionKind{
		Kind: "Node",
	}
)

type NodeGetOptions struct {
	// LabelSelector (label query) to filter on, supports '=', '==', and
	// '!='.(e.g. -l key1=value1,key2=value2). Matching objects must satisfy
	// all of the specified label constraints.
	LabelSelector string
}

// Get returns a slice of `Node` objects contained in a Kubernetes cluster.
func Get(
	ctx context.Context,
	c *kconnect.Connection,
	opts *NodeGetOptions,
) ([]*types.Node, error) {
	gvrNode, err := c.GVR(nodeGVK)
	if err != nil {
		return nil, err
	}
	lopts := metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
	}
	list, err := c.Client().Resource(gvrNode).List(
		ctx, lopts,
	)
	if err != nil {
		return nil, err
	}
	nodes := make([]*types.Node, len(list.Items))
	// Grab the entire set of Pods in the cluster and create a map, keyed by
	// node name, of Pod structs.
	pods, err := kpod.Get(ctx, c)
	if err != nil {
		return nil, err
	}
	nodePods := make(map[string][]*types.Pod, len(nodes))
	for _, p := range pods {
		np, ok := nodePods[p.Node]
		if !ok {
			np = []*types.Pod{}
		}
		np = append(np, p)
		nodePods[p.Node] = np
	}

	for x, obj := range list.Items {
		name, _, _ := unstructured.NestedString(obj.Object, "metadata", "name")
		podsOnNode, hasPods := nodePods[name]
		cpuCap, err := resourceCapacityFromRaw(obj.Object, "cpu")
		if err != nil {
			return nil, err
		}
		cpuAlloc, err := resourceAllocatableFromRaw(obj.Object, "cpu")
		if err != nil {
			return nil, err
		}
		cpuReserved := cpuCap - cpuAlloc
		var cpuReqFloor float64 = 0
		var cpuReqCeil float64 = 0
		if hasPods {
			for _, p := range podsOnNode {
				cpuReqFloor += p.ResourceRequests.CPU.Floor
				cpuReqCeil += p.ResourceRequests.CPU.Ceiling
			}
		}
		memCap, err := resourceCapacityFromRaw(obj.Object, "memory")
		if err != nil {
			return nil, err
		}
		memAlloc, err := resourceAllocatableFromRaw(obj.Object, "memory")
		if err != nil {
			return nil, err
		}
		memReserved := memCap - memAlloc
		var memReqFloor float64 = 0
		var memReqCeil float64 = 0
		if hasPods {
			for _, p := range podsOnNode {
				memReqFloor += p.ResourceRequests.Memory.Floor
				memReqCeil += p.ResourceRequests.Memory.Ceiling
			}
		}
		podCap, err := resourceCapacityFromRaw(obj.Object, "pods")
		if err != nil {
			return nil, err
		}
		podAlloc, err := resourceAllocatableFromRaw(obj.Object, "pods")
		if err != nil {
			return nil, err
		}
		podReserved := podCap - podAlloc
		var podCount float64 = 0
		if hasPods {
			podCount = float64(len(podsOnNode))
		}
		nodeRes := types.Resources{
			CPU: types.ResourceAmounts{
				Capacity:         cpuCap,
				Allocatable:      cpuAlloc,
				Reserved:         cpuReserved,
				RequestedFloor:   cpuReqFloor,
				RequestedCeiling: cpuReqCeil,
			},
			Memory: types.ResourceAmounts{
				Capacity:         memCap,
				Allocatable:      memAlloc,
				Reserved:         memReserved,
				RequestedFloor:   memReqFloor,
				RequestedCeiling: memReqCeil,
			},
			Pods: types.ResourceAmounts{
				Capacity:         podCap,
				Allocatable:      podAlloc,
				Reserved:         podReserved,
				RequestedFloor:   podCount,
				RequestedCeiling: podCount,
				Used:             podCount,
			},
		}
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
