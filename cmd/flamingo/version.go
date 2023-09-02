package main

import "github.com/spf13/cobra"

var ServerVersion = "dev"
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Flamingo",
	Long: `
# Print the version numbers of Flamingo CLI and Server
flamingo version
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("Flamingo CLI version %s - Server version %s\n", Version, ServerVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
