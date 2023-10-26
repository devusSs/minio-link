package cmd

import (
	"github.com/devusSs/minio-link/internal/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates the application if there are updates available",
	Long:  "Updating will only work if you have proper build information setup.",
	Run: func(cmd *cobra.Command, args []string) {
		err := updater.CheckForUpdatesAndApply(BuildVersion)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
