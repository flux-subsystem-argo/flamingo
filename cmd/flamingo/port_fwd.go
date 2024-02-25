package main

import (
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var portFwdCmd = &cobra.Command{
	Use:     "port-fwd",
	Args:    cobra.NoArgs,
	Aliases: []string{"port-forward", "pf"},
	Short:   "Port forward to the Flamingo service",
	RunE:    portFwdCmdRun,
}

var portFwdFlags struct {
	port string
}

func init() {
	portFwdCmd.Flags().StringVar(&portFwdFlags.port, "port", "8080", "port to forward to")

	rootCmd.AddCommand(portFwdCmd)
}

func portFwdCmdRun(_ *cobra.Command, _ []string) error {
	cmd := exec.Command("kubectl", "-n", rootArgs.applicationNamespace, "port-forward", "svc/argocd-server", portFwdFlags.port+":443")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	logger.Actionf("running kubectl -n %s port-forward svc/argocd-server %s:443", rootArgs.applicationNamespace, portFwdFlags.port)
	logger.Waitingf("address: https://localhost:%s", portFwdFlags.port)
	return cmd.Run()
}
