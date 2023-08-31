package main

import "github.com/spf13/cobra"

var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Flamingo",
	Long: `
# Print the version number of Flamingo
flamingo version
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("Flamingo version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
