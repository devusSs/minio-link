package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/devusSs/minio-link/internal/config/environment"
	"github.com/devusSs/minio-link/internal/minio"
	"github.com/devusSs/minio-link/internal/yourls"
	"github.com/devusSs/minio-link/pkg/log"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download [link]",
	Short: "Downloads a file from MinIO via it's shortened YOURLS url",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()

		link := args[0]
		cfgPath := cmd.Flag("config").Value.String()
		logsPath := cmd.Flag("logs").Value.String()
		debug, err := cmd.Flags().GetBool("debug")
		cobra.CheckErr(err)
		filepath := cmd.Flag("filepath").Value.String()
		fmt.Println(filepath)

		downloadLogger := log.NewLogger().
			WithDirectory(logsPath).
			WithName("download").
			WithDebug(debug).
			WithConsoleOutput(debug)

		cfg, err := environment.Load(cfgPath)
		if err != nil {
			downloadLogger.Error(err.Error())
			os.Exit(1)
		}

		downloadLogger.Debug(fmt.Sprintf("loaded config: %v", cfg))

		if !cfg.MinioUseSSL {
			downloadLogger.Warn("minio not using SSL / TLS (INSECURE)")
		}

		if !strings.Contains(cfg.YourlsEndpoint, "https://") {
			downloadLogger.Warn("yourls not using SSL / TLS (INSECURE)")
		}

		stopChan := make(chan bool, 1)
		cancelChannel := make(chan os.Signal, 1)
		signal.Notify(cancelChannel, os.Interrupt, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			select {
			case sig := <-cancelChannel:
				downloadLogger.Debug(fmt.Sprintf("received sys signal: %s", sig.String()))
				cancel()
				return
			case <-stopChan:
				downloadLogger.Debug("received stop signal")
				return
			}
		}()

		yourlsClient := yourls.NewClient(logsPath, debug, cfg)
		originalURL, err := yourlsClient.ExpandURL(ctx, link)
		if err != nil {
			downloadLogger.Error(err.Error())
			os.Exit(1)
		}

		downloadLogger.Debug(fmt.Sprintf("original url: %s", originalURL))

		minioClient, err := minio.NewClient(logsPath, debug, cfg)
		if err != nil {
			downloadLogger.Error(err.Error())
			os.Exit(1)
		}

		if err := minioClient.DownloadFile(ctx, originalURL, filepath); err != nil {
			downloadLogger.Error(err.Error())
			os.Exit(1)
		}

		close(stopChan)
		close(cancelChannel)
		downloadLogger.Debug("closed stop and cancel channels")

		downloadLogger.Info("Downloading file done")
		downloadLogger.Debug(fmt.Sprintf("took %s", time.Since(startTime).String()))
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringP("config", "c", "", "Sets the path of our env file if wanted")
	downloadCmd.Flags().StringP("logs", "l", "./logs", "Sets the path for our logs file")
	downloadCmd.Flags().BoolP("debug", "d", false, "Sets the debug mode for our application")
	downloadCmd.Flags().
		StringP("filepath", "f", "", "Sets a custom filepath for the downloaded file")
}
