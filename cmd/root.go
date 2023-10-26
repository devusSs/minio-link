package cmd

import (
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/devusSs/minio-link/pkg/system"
	"github.com/spf13/cobra"
)

var (
	supportedOS []string = []string{"macOS", "Windows", "Linux"}

	rootCmd = &cobra.Command{
		Use:   "minio-link",
		Short: "File management via MinIO and shortening via YOURLS",
		Long: `Minio-Link is a CLI tool to upload files via MinIO and shorten the share urls via YOURLS.
It also provides a possibility to download the files via the shortened url.`,
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(checkOS)
}

func checkOS() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	osV, err := system.GetOS(ctx)
	cobra.CheckErr(err)

	if !slices.Contains(supportedOS, osV) {
		cobra.CheckErr(fmt.Sprintf("os %s not supported", osV))
	}
}
