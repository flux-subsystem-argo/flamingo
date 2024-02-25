package main

import (
	ctx "context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/flux-subsystem-argo/flamingo/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var addClusterCmd = &cobra.Command{
	Use:   "add-cluster CONTEXT_NAME",
	Short: "Add a cluster to Flamingo",
	Long: `
# Add a cluster to Flamingo
flamingo add-cluster my-cluster

# Add cluster dev-1, override the name and address, and skip TLS verification
flamingo add-cluster dev-1 \
  --server-name=dev-1.example.com \
  --server-addr=https://dev-1.example.com:6443 \
  --insecure
`,
	Args: cobra.ExactArgs(1),
	RunE: addClusterCmdRun,
}

var addClusterFlags struct {
	insecureSkipTLSVerify bool
	serverName            string
	serverAddress         string
	export                bool
}

func init() {
	addClusterCmd.Flags().BoolVar(&addClusterFlags.insecureSkipTLSVerify, "insecure", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	addClusterCmd.Flags().StringVar(&addClusterFlags.serverName, "server-name", "", "If set, this overrides the hostname used to validate the server certificate")
	addClusterCmd.Flags().StringVar(&addClusterFlags.serverAddress, "server-addr", "", "If set, this overrides the server address used to connect to the cluster")
	addClusterCmd.Flags().BoolVar(&addClusterFlags.export, "export", false, "export manifests instead of installing")

	rootCmd.AddCommand(addClusterCmd)
}

func addClusterCmdRun(cmd *cobra.Command, args []string) error {
	leafClusterContext := args[0]

	kubeconfig := ""
	if *kubeconfigArgs.KubeConfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	} else {
		kubeconfig = *kubeconfigArgs.KubeConfig
	}

	// Load the kubeconfig file
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return err
	}

	// Specify the context name
	contextName := leafClusterContext

	// Get the context
	context, exists := config.Contexts[contextName]
	if !exists {
		return fmt.Errorf("context not found")
	}

	// Get the cluster info
	cluster, exists := config.Clusters[context.Cluster]
	if !exists {
		return fmt.Errorf("cluster not found")
	}

	// Get the user info
	user, exists := config.AuthInfos[context.AuthInfo]
	if !exists {
		return fmt.Errorf("user not found")
	}

	template := `apiVersion: v1
kind: Secret
metadata:
  name: %s-cluster
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: cluster
    flamingo/cluster: "true"
  annotations: 
    flamingo/external-address: "%s"
    flamingo/internal-address: "%s"
type: Opaque
stringData:
  name: %s
  server: %s
  config: |
    {
      "tlsClientConfig": {
        "insecure": %v,
        "certData": "%s",
        "keyData": "%s",
        "serverName": "%s"
      }
    }
`
	serverAddress := addClusterFlags.serverAddress
	if serverAddress == "" {
		serverAddress = cluster.Server
	}

	result := fmt.Sprintf(template,
		contextName,
		cluster.Server, // external address (known to the user via kubectl config view)
		serverAddress,  // internal address
		contextName,
		serverAddress,
		addClusterFlags.insecureSkipTLSVerify,
		base64.StdEncoding.EncodeToString(user.ClientCertificateData),
		base64.StdEncoding.EncodeToString(user.ClientKeyData),
		addClusterFlags.serverName,
	)

	if addClusterFlags.export {
		fmt.Print(result)
		return nil
	} else {
		logger.Actionf("applying generated cluster secret %s in %s namespace", contextName+"-cluster", rootArgs.applicationNamespace)
		applyOutput, err := utils.Apply(ctx.Background(), kubeconfigArgs, kubeclientOptions, []byte(result))
		if err != nil {
			return fmt.Errorf("apply failed: %w", err)
		}
		fmt.Fprintln(os.Stderr, applyOutput)
	}

	return nil
}
