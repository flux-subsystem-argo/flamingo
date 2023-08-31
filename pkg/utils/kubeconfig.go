package utils

import (
	"fmt"

	"github.com/fluxcd/pkg/runtime/client"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

func KubeConfig(rcg genericclioptions.RESTClientGetter, opts *client.Options) (*rest.Config, error) {
	cfg, err := rcg.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("kubernetes configuration load failed: %w", err)
	}

	// avoid throttling request when some Flux CRDs are not registered
	cfg.QPS = opts.QPS
	cfg.Burst = opts.Burst

	return cfg, nil
}
