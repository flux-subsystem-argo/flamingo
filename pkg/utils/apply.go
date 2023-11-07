package utils

import (
	"bytes"
	"context"
	"fmt"
	helmv2b1 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	"time"

	"github.com/fluxcd/pkg/ssa"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"

	runclient "github.com/fluxcd/pkg/runtime/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Apply(ctx context.Context, rcg genericclioptions.RESTClientGetter, opts *runclient.Options, resources []byte) (string, error) {
	objs, err := ssa.ReadObjects(bytes.NewReader(resources))
	if err != nil {
		return "", err
	}

	if len(objs) == 0 {
		return "", fmt.Errorf("no Kubernetes objects found")
	}

	if err := ssa.SetNativeKindsDefaults(objs); err != nil {
		return "", err
	}

	changeSet := ssa.NewChangeSet()

	// contains only CRDs and Namespaces
	var stageOne []*unstructured.Unstructured

	// contains all objects except for CRDs and Namespaces
	var stageTwo []*unstructured.Unstructured

	for _, u := range objs {
		if ssa.IsClusterDefinition(u) {
			stageOne = append(stageOne, u)
		} else {
			stageTwo = append(stageTwo, u)
		}
	}

	if len(stageOne) > 0 {
		cs, err := applySet(ctx, rcg, opts, stageOne)
		if err != nil {
			return "", err
		}
		changeSet.Append(cs.Entries)
	}

	if len(changeSet.Entries) > 0 {
		if err := waitForSet(rcg, opts, changeSet); err != nil {
			return "", err
		}
	}

	if len(stageTwo) > 0 {
		cs, err := applySet(ctx, rcg, opts, stageTwo)
		if err != nil {
			return "", err
		}
		changeSet.Append(cs.Entries)
	}

	if len(changeSet.Entries) > 0 {
		if err := waitForSet(rcg, opts, changeSet); err != nil {
			return "", err
		}
	}

	return changeSet.String(), nil
}

func NewScheme() *apiruntime.Scheme {
	scheme := apiruntime.NewScheme()
	_ = apiextensionsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = sourcev1.AddToScheme(scheme)
	_ = sourcev1b2.AddToScheme(scheme)
	_ = kustomizev1.AddToScheme(scheme)
	_ = helmv2b1.AddToScheme(scheme)
	return scheme
}

func newManager(rcg genericclioptions.RESTClientGetter, opts *runclient.Options) (*ssa.ResourceManager, error) {
	cfg, err := KubeConfig(rcg, opts)
	if err != nil {
		return nil, err
	}
	restMapper, err := rcg.ToRESTMapper()
	if err != nil {
		return nil, err
	}
	kubeClient, err := client.New(cfg, client.Options{Mapper: restMapper, Scheme: NewScheme()})
	if err != nil {
		return nil, err
	}
	kubePoller := polling.NewStatusPoller(kubeClient, restMapper, polling.Options{})

	return ssa.NewResourceManager(kubeClient, kubePoller, ssa.Owner{
		Field: "flamingo",
		Group: "flamingo.io",
	}), nil

}

func applySet(ctx context.Context, rcg genericclioptions.RESTClientGetter, opts *runclient.Options, objects []*unstructured.Unstructured) (*ssa.ChangeSet, error) {
	man, err := newManager(rcg, opts)
	if err != nil {
		return nil, err
	}

	return man.ApplyAll(ctx, objects, ssa.DefaultApplyOptions())
}

func waitForSet(rcg genericclioptions.RESTClientGetter, opts *runclient.Options, changeSet *ssa.ChangeSet) error {
	man, err := newManager(rcg, opts)
	if err != nil {
		return err
	}
	return man.WaitForSet(changeSet.ToObjMetadataSet(), ssa.WaitOptions{Interval: 2 * time.Second, Timeout: 5 * time.Minute})
}
