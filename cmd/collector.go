package cmd

import (
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

func init() {
	RootCmd.AddCommand(collectorCmd)
}

var collectorCmd = &cobra.Command{
	Use:   "collector",
	Short: "starts the collector",
	Long:  "Will poll GitHub for changes and notify agents with updates.",
	Run: func(cmd *cobra.Command, args []string) {
		simplelog.Infof("starting up")

		for _, cv := range []string{"organizationName", "githubAccessToken", "collectorCacheTTL", "collectorHTTPTimeout", "collectorHTTPAddress"} {
			if viper.GetString(cv) == "" {
				simplelog.Errorf("please specify a config value for %s", cv)
				os.Exit(-1)
			}
		}

		cache := gskp.NewKeyCache(viper.GetString("organizationName"), viper.GetString("githubAccessToken"), time.Duration(viper.GetInt("collectorCacheTTL"))*time.Second)

		server, err := gskp.NewServer(cache)
		if err != nil {
			simplelog.Errorf("failed to create HTTP server, exiting: %v", err)
			os.Exit(-1)
		}

		shutdownComplete := make(chan bool, 1)

		// handle interrupt
		sigChannel := make(chan os.Signal, 1)
		signal.Notify(sigChannel, os.Interrupt)
		go func() {
			<-sigChannel
			simplelog.Infof("received interrupt: shutdown started, waiting for server to stop")
			server.Stop(time.Duration(viper.GetInt("collectorHTTPTimeout")) * time.Second)
			shutdownComplete <- true
		}()

		if err := server.Start(viper.GetString("collectorHTTPAddress"), time.Duration(viper.GetInt("collectorHTTPTimeout"))*time.Second); err != nil {
			simplelog.Errorf("failed to start HTTP server, exiting: %v", err)
			os.Exit(-1)
		}

		<-shutdownComplete
		simplelog.Infof("shutdown complete, exiting now")
	},
}
