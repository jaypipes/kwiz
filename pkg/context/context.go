// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package context

import (
	"context"

	kconnect "github.com/jaypipes/kwiz/pkg/kube/connect"
)

type ContextKey string

var (
	kubeConfigPathKey = ContextKey("kwiz.kube.config_path")
	kubeContextKey    = ContextKey("kwiz.kube.context")
	connectionKey     = ContextKey("kwiz.connection")
)

// ContextModifier sets some value on the context
type ContextModifier func(context.Context) context.Context

// WithKubeConfigPath sets a context's kube config path
func WithKubeConfigPath(path string) ContextModifier {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, kubeConfigPathKey, path)
	}
}

// KubeConfigPath returns any Kubernetes config path saved in the context
func KubeConfigPath(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(kubeConfigPathKey); v != nil {
		return v.(string)
	}
	return ""
}

// WithKubeContext sets a context's kube config path
func WithKubeContext(path string) ContextModifier {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, kubeContextKey, path)
	}
}

// KubeContext returns any Kubernetes context saved in the context
func KubeContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(kubeContextKey); v != nil {
		return v.(string)
	}
	return ""
}

// Connection returns any Kubernetes context saved in the context
func Connection(ctx context.Context) *kconnect.Connection {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(connectionKey); v != nil {
		return v.(*kconnect.Connection)
	}
	return nil
}

// RegisterConnection registers a Connection struct with the context
func RegisterConnection(
	ctx context.Context,
	conn *kconnect.Connection,
) context.Context {
	return context.WithValue(ctx, connectionKey, conn)
}

// New returns a new Context
func New(mods ...ContextModifier) context.Context {
	ctx := context.TODO()
	for _, mod := range mods {
		ctx = mod(ctx)
	}
	return ctx
}
