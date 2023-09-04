package utils

import (
	runclient "github.com/fluxcd/pkg/runtime/client"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func KubeClient(rcg genericclioptions.RESTClientGetter, opts *runclient.Options) (client.Client, error) {
	cfg, err := KubeConfig(rcg, opts)
	if err != nil {
		return nil, err
	}
	restMapper, err := rcg.ToRESTMapper()
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{Mapper: restMapper, Scheme: NewScheme()})
}
