package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/devusSs/minio-link/internal/clip"
	"github.com/devusSs/minio-link/internal/config/environment"
	"github.com/devusSs/minio-link/internal/minio"
	"github.com/devusSs/minio-link/internal/yourls"
	"github.com/devusSs/minio-link/pkg/log"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload [file path]",
	Short: "Uploads a file to MinIO and then shortens the url via YOURLS",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()

		file := args[0]
		cfgPath := cmd.Flag("config").Value.String()
		logsPath := cmd.Flag("logs").Value.String()
		debug, err := cmd.Flags().GetBool("debug")
		cobra.CheckErr(err)
		private, err := cmd.Flags().GetBool("private")
		cobra.CheckErr(err)

		if strings.Contains(logsPath, "./") {
			exe, err := os.Executable()
			cobra.CheckErr(err)

			logsPath = filepath.Join(filepath.Dir(exe), logsPath)
		}

		uploadLogger := log.NewLogger().
			WithDirectory(logsPath).
			WithName("upload").
			WithDebug(debug).
			WithConsoleOutput(debug)

		cfg, err := environment.Load(cfgPath)
		if err != nil {
			uploadLogger.Error(err.Error())
			os.Exit(1)
		}

		uploadLogger.Debug(fmt.Sprintf("loaded config: %v", cfg))

		if !cfg.MinioUseSSL {
			uploadLogger.Warn("minio not using SSL / TLS (INSECURE)")
		}

		if !strings.Contains(cfg.YourlsEndpoint, "https://") {
			uploadLogger.Warn("yourls not using SSL / TLS (INSECURE)")
		}

		stopChan := make(chan bool, 1)
		cancelChannel := make(chan os.Signal, 1)
		signal.Notify(cancelChannel, os.Interrupt, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			select {
			case sig := <-cancelChannel:
				uploadLogger.Debug(fmt.Sprintf("received sys signal: %s", sig.String()))
				cancel()
				return
			case <-stopChan:
				uploadLogger.Debug("received stop signal")
				return
			}
		}()

		minioClient, err := minio.NewClient(logsPath, debug, cfg)
		if err != nil {
			uploadLogger.Error(err.Error())
			os.Exit(1)
		}

		minioURL, err := minioClient.UploadFile(ctx, file, !private)
		if err != nil {
			uploadLogger.Error(err.Error())
			os.Exit(1)
		}

		if err := clip.CopyToClipboard(minioURL); err != nil {
			uploadLogger.Error(err.Error())
			os.Exit(1)
		}

		uploadLogger.Debug("upload to MinIO successful")

		yourlsClient := yourls.NewClient(logsPath, debug, cfg)
		shortenedURL, err := yourlsClient.ShortenURL(ctx, minioURL)
		if err != nil {
			uploadLogger.Error(err.Error())
			os.Exit(1)
		}

		uploadLogger.Debug("shortening via YOURLS successful")

		if err := clip.CopyToClipboard(shortenedURL); err != nil {
			uploadLogger.Error(err.Error())
			os.Exit(1)
		}

		close(stopChan)
		close(cancelChannel)
		uploadLogger.Debug("closed stop and cancel channels")

		uploadLogger.Info("Uploading and shortening done")
		uploadLogger.Debug(fmt.Sprintf("took %s", time.Since(startTime).String()))

		uploadLogger.Info(
			fmt.Sprintf("Links will be valid for %s", cfg.MinioDefaultExpiry.String()),
		)
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringP("config", "c", "", "Sets the path of our env file if wanted")
	uploadCmd.Flags().StringP("logs", "l", "./logs", "Sets the path for our logs file")
	uploadCmd.Flags().BoolP("debug", "d", false, "Sets the debug mode for our application")
	uploadCmd.Flags().
		BoolP("private", "p", false, "Sets the bucket and therefor uploaded files to private")
}
