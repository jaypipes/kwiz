// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package pod

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
	podGVK = schema.GroupVersionKind{
		Kind: "Pod",
	}
)

// Get returns a slice of `Pod` objects contained in a Kubernetes cluster.
func Get(
	ctx context.Context,
	c *kconnect.Connection,
) ([]*types.Pod, error) {
	gvrPod, err := c.GVR(podGVK)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	list, err := c.Client().Resource(gvrPod).List(
		ctx, opts,
	)
	if err != nil {
		return nil, err
	}
	pods := make([]*types.Pod, len(list.Items))
	for x, obj := range list.Items {
		cpuFloor, cpuCeil, err := resourceFloorCeilingFromRaw(obj.Object, "cpu")
		if err != nil {
			return nil, err
		}
		memFloor, memCeil, err := resourceFloorCeilingFromRaw(obj.Object, "memory")
		if err != nil {
			return nil, err
		}
		podResReq := types.ResourceRequests{
			CPU: types.ResourceRequest{
				Floor:   cpuFloor,
				Ceiling: cpuCeil,
			},
			Memory: types.ResourceRequest{
				Floor:   memFloor,
				Ceiling: memCeil,
			},
		}
		name, _, _ := unstructured.NestedString(obj.Object, "metadata", "name")
		nodeName, _, _ := unstructured.NestedString(obj.Object, "spec", "nodeName")
		ns, _, _ := unstructured.NestedString(obj.Object, "metadata", "namespace")
		pod := &types.Pod{
			Cluster:          "default",
			Name:             name,
			Node:             nodeName,
			Namespace:        ns,
			ResourceRequests: podResReq,
		}
		pods[x] = pod
	}
	return pods, nil
}

// resourceFloorCeilingFromRaw accepts a raw map of Kubernetes object fields
// and returns the floor and ceiling of a resource type's requests.
func resourceFloorCeilingFromRaw(
	obj map[string]interface{},
	resType string,
) (float64, float64, error) {
	floor, ceil := float64(0), float64(-1)
	ctrs, _, _ := unstructured.NestedSlice(obj, "spec", "containers")
	if len(ctrs) == 0 {
		return 0, 0, nil
	}
	for _, ctr := range ctrs {
		// The container's "requests" is the floor of requested resources.
		reqs, found, err := unstructured.NestedMap(ctr.(map[string]interface{}), "requests")
		if err != nil {
			return -1, -1, err
		}
		if found || len(reqs) > 0 {
			if amt, ok := reqs[resType]; ok {
				if resType == "memory" {
					// We need to convert any size strings for memory...
					amtFloat := unit.SizeStringToBytes(amt.(string))
					if err != nil {
						return -1, -1, err
					}
					floor += amtFloat
				} else {
					amountInt, err := strconv.Atoi(amt.(string))
					if err != nil {
						return -1, -1, err
					}
					floor += float64(amountInt)
				}
			}
		}
		// The container's "limits" is the ceiling of requested resources.
		limits, found, err := unstructured.NestedMap(ctr.(map[string]interface{}), "limits")
		if err != nil {
			return -1, -1, err
		}
		if found || len(limits) > 0 {
			if amt, ok := limits[resType]; ok {
				if resType == "memory" {
					// We need to convert any size strings for memory...
					amtFloat := unit.SizeStringToBytes(amt.(string))
					if err != nil {
						return -1, -1, err
					}
					ceil += amtFloat
				} else {
					amountInt, err := strconv.Atoi(amt.(string))
					if err != nil {
						return -1, -1, err
					}
					ceil += float64(amountInt)
				}
			}
		}
	}
	return floor, ceil, nil
}
