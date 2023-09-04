package main

import (
	"context"
	"fmt"
	"github.com/flux-subsystem-argo/cli/pkg/utils"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var showInitPasswordCmd = &cobra.Command{
	Use:   "show-init-password",
	Short: "Show the init password",
	Long: `
# Show the init password
flamingo show-init-password
`,
	RunE: runShowInitPassword,
}

func init() {
	rootCmd.AddCommand(showInitPasswordCmd)
}

func runShowInitPassword(cmd *cobra.Command, args []string) error {
	return showInitPassword()
}

// showInitPassword shows the init password
func showInitPassword() error {
	cli, err := utils.KubeClient(kubeconfigArgs, kubeclientOptions)
	if err != nil {
		return err
	}
	secret := corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      "argocd-initial-admin-secret",
			Namespace: rootArgs.applicationNamespace,
		},
	}

	if err := cli.Get(context.Background(), client.ObjectKeyFromObject(&secret), &secret); err != nil {
		return err
	}
	fmt.Println(string(secret.Data["password"]))

	return nil
}
