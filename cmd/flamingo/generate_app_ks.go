package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ksAppTemplate = `---
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
  annotations:
    weave.gitops.flamingo/base-url: "http://localhost:9001"
    weave.gitops.flamingo/cluster-name: "Default"
spec:
  destination:
    namespace: {{ .DestinationNamespace }}
    server: https://kubernetes.default.svc
  project: default
  source:
    path: {{ .Path }}
    repoURL: {{ .SourceURL }}
    targetRevision: "{{ .SourceRevision }}"
  syncPolicy:
    syncOptions:
    - ApplyOutOfSyncOnly=true
    - FluxSubsystem=true
`

func generateKustomizationApp(c client.Client, appName, objectName string, kindName string, tpl *bytes.Buffer) error {
	object := kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: *kubeconfigArgs.Namespace,
		},
	}
	key := client.ObjectKeyFromObject(&object)
	if err := c.Get(context.Background(), key, &object); err != nil {
		return fmt.Errorf("%w in namespace %q", err, object.Namespace)
	}

	sourceKind := object.Spec.SourceRef.Kind
	sourceName := object.Spec.SourceRef.Name
	sourceNamespace := object.Spec.SourceRef.Namespace
	if sourceNamespace == "" {
		sourceNamespace = *kubeconfigArgs.Namespace
	}

	t, err := template.New("template").Parse(ksAppTemplate)
	if err != nil {
		return err
	}

	var params struct {
		Name                 string
		Namespace            string
		WorkloadType         string
		WorkloadName         string
		SourceType           string
		DestinationNamespace string

		Path           string
		SourceURL      string
		SourceRevision string
	}

	params.Name = appName
	params.Namespace = rootArgs.applicationNamespace
	params.DestinationNamespace = *kubeconfigArgs.Namespace

	params.WorkloadName = object.Name
	params.WorkloadType = kindName
	params.SourceType = sourceKind

	params.Path = "."
	if object.Spec.Path != "" {
		params.Path = object.Spec.Path
	}

	switch sourceKind {
	case sourcev1.GitRepositoryKind:
		sourceObj := sourcev1.GitRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sourceName,
				Namespace: sourceNamespace,
			},
		}
		sourceKey := client.ObjectKeyFromObject(&sourceObj)
		if err := c.Get(context.Background(), sourceKey, &sourceObj); err != nil {
			return err
		}
		params.SourceURL = sourceObj.Spec.URL
		params.SourceRevision = getGitRepositorySourceRevision(sourceObj.Spec.Reference)
	case sourcev1b2.BucketKind:
		sourceObj := sourcev1b2.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sourceName,
				Namespace: sourceNamespace,
			},
		}
		sourceKey := client.ObjectKeyFromObject(&sourceObj)
		if err := c.Get(context.Background(), sourceKey, &sourceObj); err != nil {
			return err
		}
		params.SourceURL = getBucketURL(sourceObj.Spec)
		params.SourceRevision = "HEAD"
	case sourcev1b2.OCIRepositoryKind:
		sourceObj := sourcev1b2.OCIRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sourceName,
				Namespace: sourceNamespace,
			},
		}
		sourceKey := client.ObjectKeyFromObject(&sourceObj)
		if err := c.Get(context.Background(), sourceKey, &sourceObj); err != nil {
			return err
		}
		params.SourceURL = sourceObj.Spec.URL
		params.SourceRevision = getOCIRepositorySourceRevision(sourceObj.Spec.Reference)
	}

	if err := t.Execute(tpl, params); err != nil {
		return err
	}
	return nil
}

func getBucketURL(bs sourcev1b2.BucketSpec) string {
	switch bs.Provider {
	case "aws":
		if bs.Region == "" {
			return ""
		}
		return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", bs.BucketName, bs.Region)
	case "gcp":
		return fmt.Sprintf("https://storage.googleapis.com/%s/", bs.BucketName)
	case "azure":
		return fmt.Sprintf("https://%s.blob.core.windows.net/", bs.BucketName)
	case "generic":
		protocol := "https"
		if bs.Insecure {
			protocol = "http"
		}
		// Ensure the endpoint doesn't have a trailing slash.
		endpoint := strings.TrimRight(bs.Endpoint, "/")
		return fmt.Sprintf("%s://%s/%s/", protocol, endpoint, bs.BucketName)
	default:
		return ""
	}
}

func getGitRepositorySourceRevision(reference *sourcev1.GitRepositoryRef) string {
	if reference == nil {
		return "master"
	}

	if reference.Commit != "" {
		return reference.Commit
	}
	if reference.Name != "" {
		return reference.Name
	}
	if reference.SemVer != "" {
		return reference.SemVer
	}
	if reference.Tag != "" {
		return reference.Tag
	}
	if reference.Branch != "" {
		return reference.Branch
	}

	// the default value of the branch is master
	return "master"
}

func getOCIRepositorySourceRevision(reference *sourcev1b2.OCIRepositoryRef) string {
	if reference == nil {
		return "latest"
	}
	if reference.Digest != "" {
		return reference.Digest
	}
	if reference.SemVer != "" {
		return reference.SemVer
	}
	if reference.Tag != "" {
		return reference.Tag
	}

	return "latest"
}
