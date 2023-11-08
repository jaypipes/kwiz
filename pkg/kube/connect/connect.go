// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package connect

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	discocached "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	kerrors "github.com/jaypipes/kwiz/pkg/kube/errors"
)

// Connection is a struct containing a discovery client and a dynamic client
// that kwiz uses to communicate with Kubernetes.
type Connection struct {
	mapper meta.RESTMapper
	disco  discovery.CachedDiscoveryInterface
	client dynamic.Interface
}

// Client() returns the Connection's dynamic Kubernetes client interface
func (c *Connection) Client() dynamic.Interface {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client
}

// mappingFor returns a RESTMapper for a given resource type or kind
func (c *Connection) mappingFor(typeOrKind string) (*meta.RESTMapping, error) {
	fullySpecifiedGVR, groupResource := schema.ParseResourceArg(typeOrKind)
	gvk := schema.GroupVersionKind{}

	if fullySpecifiedGVR != nil {
		gvk, _ = c.mapper.KindFor(*fullySpecifiedGVR)
	}
	if gvk.Empty() {
		gvk, _ = c.mapper.KindFor(groupResource.WithVersion(""))
	}
	if !gvk.Empty() {
		return c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	}

	fullySpecifiedGVK, groupKind := schema.ParseKindArg(typeOrKind)
	if fullySpecifiedGVK == nil {
		gvk := groupKind.WithVersion("")
		fullySpecifiedGVK = &gvk
	}

	if !fullySpecifiedGVK.Empty() {
		if mapping, err := c.mapper.RESTMapping(fullySpecifiedGVK.GroupKind(), fullySpecifiedGVK.Version); err == nil {
			return mapping, nil
		}
	}

	mapping, err := c.mapper.RESTMapping(groupKind, gvk.Version)
	if err != nil {
		// if we error out here, it is because we could not match a resource or a kind
		// for the given argument. To maintain consistency with previous behavior,
		// announce that a resource type could not be found.
		// if the error is _not_ a *meta.NoKindMatchError, then we had trouble doing discovery,
		// so we should return the original error since it may help a user diagnose what is actually wrong
		if meta.IsNoMatchError(err) {
			return nil, fmt.Errorf("the server doesn't have a resource type %q", groupResource.Resource)
		}
		return nil, err
	}

	return mapping, nil
}

// GVR returns a GroupVersionResource from a GroupVersionKind, using the
// discovery client to look up the resource name (the plural of the kind). The
// returned GroupVersionResource will have the proper Group and Version filled
// in (as opposed to an APIResource which has empty Group and Version strings
// because it "inherits" its APIResourceList's GroupVersion ... ugh.)
func (c *Connection) GVR(
	gvk schema.GroupVersionKind,
) (schema.GroupVersionResource, error) {
	empty := schema.GroupVersionResource{}
	r, err := c.mappingFor(gvk.Kind)
	if err != nil {
		return empty, kerrors.ResourceUnknown(gvk)
	}

	return r.Resource, nil
}

// Connect returns a connection with a discovery client and a Kubernetes
// client-go DynamicClient to use in communicating with the Kubernetes API
func Connect(cfg *rest.Config) (*Connection, error) {
	c, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	discoverer, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	disco := discocached.NewMemCacheClient(discoverer)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(disco)
	expander := restmapper.NewShortcutExpander(mapper, disco)

	return &Connection{
		mapper: expander,
		disco:  disco,
		client: c,
	}, nil
}
