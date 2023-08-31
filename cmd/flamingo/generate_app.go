package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	helmv2b1 "github.com/fluxcd/helm-controller/api/v2beta1"
	"os"
	"strings"
	"text/template"

	"github.com/flux-subsystem-argo/cli/pkg/utils"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	runtimeclient "github.com/fluxcd/pkg/runtime/client"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var generateAppCmd = &cobra.Command{
	Use:   "generate-app NAME",
	Args:  cobra.ExactArgs(1),
	Short: "Create an application in ArgoCD from Flux resources",
	Long: `
# Generate a Flamingo application from a Flux Kustomization podinfo in the current namespace (flux-system).
# The generated application is put in the argocd namespace by default.
flamingo generate-app ks/podinfo

# Generate a Flamingo application from a Flux Kustomization podinfo in the podinfo namespace.
# The generated application is put in the argocd namespace by default.
flamingo generate-app -n podinfo ks/podinfo

# Create a Flamingo application from a HelmRelease podinfo in the current namespace (flux-system).
# The generated application is put in the argocd namespace by default.
flamingo generate-app hr/podinfo
`,
	RunE: generateAppCmdRun,
}

var generateAppFlags struct {
	export bool
}

func init() {
	generateAppCmd.Flags().BoolVar(&generateAppFlags.export, "export", false, "export the generated application to stdout")

	rootCmd.AddCommand(generateAppCmd)
}

const ksAppTemplate = `
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/managed-by: flamingo
    flamingo/app-type: "{{ .AppType }}"
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
    targetRevision: {{ .SourceRevision }}
  syncPolicy:
    syncOptions:
    - ApplyOutOfSyncOnly=true
    - FluxSubsystem=true
`

func generateAppCmdRun(cmd *cobra.Command, args []string) error {
	kubeConfig, err := utils.KubeConfig(kubeconfigArgs, kubeclientOptions)
	if err != nil {
		return err
	}

	restMapper, err := runtimeclient.NewDynamicRESTMapper(kubeConfig)
	if err != nil {
		return err
	}
	c, err := client.New(kubeConfig, client.Options{Mapper: restMapper, Scheme: utils.NewScheme()})
	if err != nil {
		return err
	}

	kindSlashName := args[0]
	isValid := false
	kindName := ""
	objectName := ""

	// Define a map for valid kinds with their short and full names
	var validKinds = map[string]string{
		"ks":             kustomizev1.KustomizationKind,
		"kustomization":  kustomizev1.KustomizationKind,
		"kustomizations": kustomizev1.KustomizationKind,
		"hr":             helmv2b1.HelmReleaseKind,
		"helmrelease":    helmv2b1.HelmReleaseKind,
		"helmreleases":   helmv2b1.HelmReleaseKind,
	}
	// Check if the kindSlashName starts with any of the valid prefixes
	for shortName, fullName := range validKinds {
		if strings.HasPrefix(kindSlashName, shortName+"/") {
			isValid = true
			kindName = fullName
			break
		} else if strings.HasPrefix(kindSlashName, fullName+"/") {
			isValid = true
			kindName = fullName
			break
		}
	}

	if !isValid {
		return fmt.Errorf("not a valid Kustomization or HelmRelease resource")
	}

	objectName = strings.Split(kindSlashName, "/")[1]

	var tpl bytes.Buffer

	if kindName == kustomizev1.KustomizationKind {
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
			DestinationNamespace string
			Path                 string
			SourceURL            string
			SourceRevision       string
			AppType              string
			SourceType           string
		}

		params.Name = object.Name
		params.Namespace = rootArgs.applicationNamespace
		params.Path = object.Spec.Path
		params.DestinationNamespace = *kubeconfigArgs.Namespace
		params.AppType = kindName
		params.SourceType = sourceKind

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

		if err := t.Execute(&tpl, params); err != nil {
			return err
		}

	} else if kindName == "helmrelease" {
		return errors.New("helmrelease not supported yet")
	}

	if generateAppFlags.export {
		fmt.Print(tpl.String())
		return nil
	} else {
		logger.Actionf("applying generated application %s in %s namespace", objectName, rootArgs.applicationNamespace)
		applyOutput, err := utils.Apply(context.Background(), kubeconfigArgs, kubeclientOptions, tpl.Bytes())
		if err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
		fmt.Fprintln(os.Stderr, applyOutput)
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
