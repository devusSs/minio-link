package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	BuildVersion   string
	BuildDate      string
	BuildGitCommit string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version and build information and exits",
	Long: `Build and version information will usually be set on compile time.
If you do not have version or build information please make sure to
either download an already compiled release or build the application
yourself properly.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("minio-link by devusSs")
		fmt.Println("---------------------")
		fmt.Printf("Build version:\t\t%s\n", BuildVersion)
		fmt.Printf("Build date:\t\t%s\n", BuildDate)
		fmt.Printf("Build Git commit:\t%s\n", BuildGitCommit)
		fmt.Println("")
		fmt.Printf("Build Go OS:\t\t%s\n", runtime.GOOS)
		fmt.Printf("Build Go arch:\t\t%s\n", runtime.GOARCH)
		fmt.Printf("Build Go version:\t%s\n", runtime.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	if BuildVersion == "" {
		BuildVersion = "unknown"
	}
	if BuildDate == "" {
		BuildDate = "unknown"
	}
	if BuildGitCommit == "" {
		BuildGitCommit = "unknown"
	}
}
