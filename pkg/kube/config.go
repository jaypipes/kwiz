// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package kube

import (
	"context"

	kwcontext "github.com/jaypipes/kwiz/pkg/context"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config returns a Kubernetes client-go `rest.Config` after evaluating
// locations to find a Kubernetes config and context.
//
// We evaluate where to retrieve the Kubernetes config from by looking at the
// following things, in this order:
//
// 1) A `gviz.kube.config_path` context value key if present
// 2) KUBECONFIG environment variable pointing at a file.
// 3) In-cluster config if running in cluster.
// 4) $HOME/.kube/config if exists.
func Config(ctx context.Context) (*rest.Config, error) {
	kcfgPath := kwcontext.KubeConfigPath(ctx)
	kctx := kwcontext.KubeContext(ctx)
	overrides := &clientcmd.ConfigOverrides{}
	if kctx != "" {
		overrides.CurrentContext = kctx
	}
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kcfgPath != "" {
		rules.ExplicitPath = kcfgPath
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules, overrides,
	).ClientConfig()
}
