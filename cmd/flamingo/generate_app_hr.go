package main

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	helmv2b1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const hrAppTemplate = `---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/managed-by: flamingo
    flamingo/workload-name: {{ .WorkloadName }}
    flamingo/workload-type: "{{ .WorkloadType }}"
    flamingo/source-type: "{{ .SourceType }}"
    flamingo/destination-namespace: "{{ .DestinationNamespace }}"
    flamingo/cluster-name: "{{ .ClusterName }}"
  annotations:
    weave.gitops.flamingo/base-url: "http://localhost:9001"
    weave.gitops.flamingo/cluster-name: "Default"
spec:
  destination:
    namespace: {{ .DestinationNamespace }}
    server: {{ .Server }}
  project: default
  source:
    chart: {{ .ChartName }}
    repoURL: {{ .ChartURL }}
    targetRevision: "{{ .ChartRevision }}"
    helm:
      releaseName: {{ .ReleaseName }}
  syncPolicy:
    syncOptions:
    - ApplyOutOfSyncOnly=true
    - FluxSubsystem=true
`

func generateHelmReleaseApp(c client.Client, appName string, objectName string, kindName string, clusterName string, server string, tpl *bytes.Buffer) error {
	object := helmv2b1.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: *kubeconfigArgs.Namespace,
		},
	}
	key := client.ObjectKeyFromObject(&object)
	if err := c.Get(context.Background(), key, &object); err != nil {
		return fmt.Errorf("%w in namespace %q", err, object.Namespace)
	}

	sourceKind := object.Spec.Chart.Spec.SourceRef.Kind
	sourceName := object.Spec.Chart.Spec.SourceRef.Name
	sourceNamespace := object.Spec.Chart.Spec.SourceRef.Namespace
	if sourceNamespace == "" {
		sourceNamespace = *kubeconfigArgs.Namespace
	}

	t, err := template.New("template").Parse(hrAppTemplate)
	if err != nil {
		return err
	}

	var params struct {
		Name                 string
		Namespace            string
		WorkloadName         string
		WorkloadType         string
		SourceType           string
		DestinationNamespace string
		ClusterName          string

		Server string

		ChartName     string
		ChartURL      string
		ChartRevision string
		ReleaseName   string
	}

	params.Name = appName
	params.Namespace = rootArgs.applicationNamespace
	params.DestinationNamespace = *kubeconfigArgs.Namespace

	params.WorkloadName = object.Name
	params.WorkloadType = kindName
	params.SourceType = sourceKind
	params.ClusterName = clusterName

	params.Server = server

	params.ChartName = object.Spec.Chart.Spec.Chart
	params.ChartRevision = object.Spec.Chart.Spec.Version
	params.ReleaseName = object.Spec.ReleaseName
	if params.ReleaseName == "" {
		params.ReleaseName = object.Name
	}

	switch sourceKind {
	case sourcev1b2.HelmRepositoryKind:
		sourceObj := sourcev1b2.HelmRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sourceName,
				Namespace: sourceNamespace,
			},
		}
		sourceKey := client.ObjectKeyFromObject(&sourceObj)
		if err := c.Get(context.Background(), sourceKey, &sourceObj); err != nil {
			return err
		}
		params.ChartURL = sourceObj.Spec.URL
	}

	if err := t.Execute(tpl, params); err != nil {
		return err
	}
	return nil
}
