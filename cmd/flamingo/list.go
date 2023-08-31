package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

const url = "https://raw.githubusercontent.com/flux-subsystem-argo/index/main/index.json"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List candidates",
	RunE:  listCmdRun,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listCmdRun(cmd *cobra.Command, args []string) error {
	resp, err := http.Get(url)
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

	// Use tabwriter to print in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)

	fmt.Fprintln(w, "FLAMINGO\tFSA-IMAGE\tSUPPORTED FLUX")
	for _, candidate := range candidates.Candidates {
		fmt.Fprintf(w, "%s\t%s\t%s\n", candidate.ArgoCD, candidate.Fsa, candidate.Flux)
	}
	w.Flush()

	return nil
}
