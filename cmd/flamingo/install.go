package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/flux-subsystem-argo/cli/pkg/utils"
	"github.com/fluxcd/flux2/v2/pkg/status"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Flux Subsystem for Argo",
	Long: fmt.Sprintf(`
# Install the Flux Subsystem for Argo
flamingo install --version=%s

# Install the Flux Subsystem for Argo with the read-only mode enabled
flamingo install --version=%s --read-only-mode
`, installFlags.version, installFlags.version),
	RunE: installCmdRun,
}

var installFlags struct {
	version      string
	dev          bool
	readOnlyMode bool
	export       bool
}

func init() {
	installCmd.Flags().StringVarP(&installFlags.version, "version", "v", ServerVersion, "version of Flamingo to install")
	installCmd.Flags().BoolVar(&installFlags.readOnlyMode, "read-only-mode", false, "enable read-only mode")
	installCmd.Flags().BoolVar(&installFlags.dev, "dev", false, "install development version")
	installCmd.Flags().BoolVar(&installFlags.export, "export", false, "export manifests instead of installing")

	rootCmd.AddCommand(installCmd)
}

func installCmdRun(cmd *cobra.Command, args []string) error {
	if installFlags.export {
		logger.stderr = io.Discard
	}

	if installFlags.version == "" {
		return cmd.Help()
	}
	logger.Actionf("obtaining version info")

	client := &http.Client{}

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Set headers to prevent caching
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the entire response body into a byte buffer
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var candidates CandidateList
	if err := json.Unmarshal(buf, &candidates); err != nil {
		return err
	}

	// filter out .Flamingo that ends with -dev if --dev flag is not set
	if !installFlags.dev {
		var filteredCandidates []Candidate
		for _, candidate := range candidates.Candidates {
			if !isDev(candidate) {
				filteredCandidates = append(filteredCandidates, candidate)
			}
		}
		candidates.Candidates = filteredCandidates
	}

	var candidate *Candidate
	for i, c := range candidates.Candidates {
		if c.Flamingo == installFlags.version {
			candidate = &candidates.Candidates[i]
			break
		}
	}
	if candidate == nil {
		return fmt.Errorf("version %s not found", installFlags.version)
	}

	if err := installFluxSubsystemForArgo(*candidate, installFlags.readOnlyMode, installFlags.export); err != nil {
		return err
	}

	if !installFlags.export {
		if err := verifyTheInstallation(); err != nil {
			return err
		}
	}

	return nil
}

func buildComponentObjectRefs(namespace string, components ...string) ([]object.ObjMetadata, error) {
	var objRefs []object.ObjMetadata
	for _, deployment := range components {
		objRefs = append(objRefs, object.ObjMetadata{
			Namespace: namespace,
			Name:      deployment,
			GroupKind: schema.GroupKind{Group: "apps", Kind: "Deployment"},
		})
	}
	return objRefs, nil
}

func verifyTheInstallation() error {
	logger.Waitingf("verifying installation")

	kubeConfig, err := utils.KubeConfig(kubeconfigArgs, kubeclientOptions)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	statusChecker, err := status.NewStatusChecker(kubeConfig, 5*time.Second, rootArgs.timeout, logger)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	// we install Argo CD components in the application namespace (argocd), not the flux-system namespace
	objectRefs, err := buildComponentObjectRefs(
		rootArgs.applicationNamespace,
		"argocd-redis",
		"argocd-dex-server",
		"argocd-repo-server",
		"argocd-server",
		"argocd-notifications-controller",
		"argocd-applicationset-controller",
	)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	if err := statusChecker.Assess(objectRefs...); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	logger.Successf("install finished")
	return nil
}

var namespaceTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  name: %s
`

func installFluxSubsystemForArgo(candidate Candidate, readOnlyMode bool, export bool) error {
	logger.Generatef("generating manifests")

	tmpl := defaultTemplate
	if readOnlyMode {
		tmpl = readonlyTemplate
	}

	var tpl bytes.Buffer
	t, err := template.New("template").Parse(tmpl)
	if err != nil {
		return err
	}

	if err := t.Execute(&tpl, struct {
		Flamingo  string
		ArgoCD    string
		Image     string
		Namespace string
	}{
		Flamingo:  candidate.Flamingo,
		ArgoCD:    candidate.ArgoCD,
		Image:     candidate.Image,
		Namespace: rootArgs.applicationNamespace,
	}); err != nil {
		return err
	}

	// Use Kustomize (krusty) to build the kustomization
	fSys := filesys.MakeFsInMemory()
	kustomizationPath := "/app/kustomization.yaml"
	fSys.WriteFile(kustomizationPath, tpl.Bytes())
	namespacePath := "/app/namespace.yaml"
	fSys.WriteFile(namespacePath, []byte(fmt.Sprintf(namespaceTemplate, rootArgs.applicationNamespace)))

	opts := krusty.MakeDefaultOptions()
	opts.Reorder = krusty.ReorderOptionLegacy
	k := krusty.MakeKustomizer(opts)

	m, err := k.Run(fSys, "/app")
	if err != nil {
		return err
	}

	yamlOutput, err := m.AsYaml()
	if err != nil {
		return err
	}
	logger.Successf("manifests build completed")

	if export {
		fmt.Println(string(yamlOutput))
		return nil
	}

	logger.Actionf("installing components in %s namespace", rootArgs.applicationNamespace)
	applyOutput, err := utils.Apply(context.Background(), kubeconfigArgs, kubeclientOptions, yamlOutput)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}
	fmt.Fprintln(os.Stderr, applyOutput)

	return nil
}
