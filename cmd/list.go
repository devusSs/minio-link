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

	"github.com/devusSs/minio-link/internal/config/environment"
	"github.com/devusSs/minio-link/internal/minio"
	"github.com/devusSs/minio-link/internal/yourls"
	"github.com/devusSs/minio-link/pkg/log"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()

		cfgPath := cmd.Flag("config").Value.String()
		logsPath := cmd.Flag("logs").Value.String()
		debug, err := cmd.Flags().GetBool("debug")
		cobra.CheckErr(err)
		limit, err := cmd.Flags().GetInt("limit")
		cobra.CheckErr(err)

		if strings.Contains(logsPath, "./") {
			exe, err := os.Executable()
			cobra.CheckErr(err)

			logsPath = filepath.Join(filepath.Dir(exe), logsPath)
		}

		listLogger := log.NewLogger().
			WithDirectory(logsPath).
			WithName("list").
			WithDebug(debug).
			WithConsoleOutput(debug)

		cfg, err := environment.Load(cfgPath)
		if err != nil {
			listLogger.Error(err.Error())
			os.Exit(1)
		}

		listLogger.Debug(fmt.Sprintf("loaded config: %v", cfg))

		if !cfg.MinioUseSSL {
			listLogger.Warn("minio not using SSL / TLS (INSECURE)")
		}

		if !strings.Contains(cfg.YourlsEndpoint, "https://") {
			listLogger.Warn("yourls not using SSL / TLS (INSECURE)")
		}

		stopChan := make(chan bool, 1)
		cancelChannel := make(chan os.Signal, 1)
		signal.Notify(cancelChannel, os.Interrupt, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			select {
			case sig := <-cancelChannel:
				listLogger.Debug(fmt.Sprintf("received sys signal: %s", sig.String()))
				cancel()
				return
			case <-stopChan:
				listLogger.Debug("received stop signal")
				return
			}
		}()

		yClient := yourls.NewClient(logsPath, debug, cfg)
		yURLs, err := yClient.GetSavedURLs(ctx, limit)
		if err != nil {
			listLogger.Error(err.Error())
			os.Exit(1)
		}

		minioClient, err := minio.NewClient(logsPath, debug, cfg)
		if err != nil {
			listLogger.Error(err.Error())
			os.Exit(1)
		}

		links := make([]string, 0, len(yURLs))
		for _, yURL := range yURLs {
			links = append(links, yURL)
		}

		err = minioClient.GetObjects(ctx, links)
		if err != nil {
			listLogger.Error(err.Error())
			os.Exit(1)
		}

		close(stopChan)
		close(cancelChannel)
		listLogger.Debug("closed stop and cancel channels")

		listLogger.Info("Listing done")
		listLogger.Debug(fmt.Sprintf("took %s", time.Since(startTime).String()))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringP("config", "c", "", "Sets the path of our env file if wanted")
	listCmd.Flags().StringP("logs", "l", "./logs", "Sets the path for our logs file")
	listCmd.Flags().BoolP("debug", "d", false, "Sets the debug mode for our application")
	listCmd.Flags().IntP("limit", "i", 20, "Sets the limit for urls to fetch")
}
