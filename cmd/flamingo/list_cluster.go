package main

import (
	ctx "context"
	"fmt"
	"github.com/flux-subsystem-argo/flamingo/pkg/utils"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"text/tabwriter"
)

var listClusterCmd = &cobra.Command{
	Use:     "list-clusters",
	Aliases: []string{"list-cluster", "lc"},
	Short:   "List clusters",
	Long: `
# List clusters
flamingo list-clusters
`,
	Args: cobra.NoArgs,
	RunE: listClusterCmdRun,
}

func init() {
	rootCmd.AddCommand(listClusterCmd)
}

func listClusterCmdRun(_ *cobra.Command, _ []string) error {
	cli, err := utils.KubeClient(kubeconfigArgs, kubeclientOptions)
	if err != nil {
		return err
	}
	// list clusters
	list := &corev1.SecretList{}
	opts := &client.ListOptions{
		Namespace: rootArgs.applicationNamespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"flamingo/cluster": "true",
		}),
	}
	if err := cli.List(ctx.Background(), list, opts); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tEXTERNAL ADDRESS\tINTERNAL ADDRESS")
	// hard code in-cluster info
	fmt.Fprintf(w, "%s\t%s\t%s\n", "in-cluster", "-", "https://kubernetes.default.svc")
	for _, s := range list.Items {
		fmt.Fprintf(w, "%s\t%s\t%s\n", string(s.Data["name"]), s.Annotations["flamingo/external-address"], s.Annotations["flamingo/internal-address"])
	}
	w.Flush()

	return nil
}
