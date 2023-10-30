package main

import (
	"context"

	"github.com/flux-subsystem-argo/flamingo/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Args:  cobra.ExactArgs(1),
	Short: "Delete Flamingo applications",
	Long:  `Delete Flamingo applications`,
	RunE:  deleteCmdRun,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func deleteCmdRun(cmd *cobra.Command, args []string) error {
	cli, err := utils.KubeClient(kubeconfigArgs, kubeclientOptions)
	if err != nil {
		return err
	}
	gvk := schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Application",
	}

	namespace := *kubeconfigArgs.Namespace
	labelSelector := map[string]string{
		"app.kubernetes.io/managed-by":   "flamingo",
		"flamingo/destination-namespace": namespace,
	}
	if getCmdFlags.all {
		delete(labelSelector, "flamingo/destination-namespace")
	}
	app := &unstructured.Unstructured{}
	app.SetName(args[0])
	app.SetNamespace(namespace)
	app.SetGroupVersionKind(gvk)

	if err := cli.Delete(context.Background(), app); err != nil {
		return err
	}

	return nil
}
