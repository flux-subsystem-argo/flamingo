package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

const url = "https://raw.githubusercontent.com/flux-subsystem-argo/cli/main/index/index.json"

var listCandidates = &cobra.Command{
	Use:     "list-candidates",
	Short:   "List candidates",
	Aliases: []string{"lc"},
	RunE:    listCmdRun,
}

var listCandidatesFlags struct {
	dev bool
}

func init() {
	listCandidates.Flags().BoolVar(&listCandidatesFlags.dev, "dev", false, "list development candidates")

	rootCmd.AddCommand(listCandidates)
}

func listCmdRun(cmd *cobra.Command, args []string) error {
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

	// Use tabwriter to print in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)

	fmt.Fprintln(w, "FLAMINGO\tFSA-IMAGE\tSUPPORTED FLUX")
	for _, candidate := range candidates.Candidates {
		if !listCandidatesFlags.dev && isDev(candidate) {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", candidate.Flamingo, candidate.Image, candidate.Flux)
	}
	w.Flush()

	return nil
}

func isDev(candidate Candidate) bool {
	return strings.HasSuffix(candidate.Flamingo, "-dev")
}
