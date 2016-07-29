package cmd

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/utilitywarehouse/github-sshkey-provider/pkg/authorizedkeys"
	"github.com/utilitywarehouse/github-sshkey-provider/pkg/collector"
	"github.com/utilitywarehouse/github-sshkey-provider/pkg/simplelog"
	"github.com/utilitywarehouse/github-sshkey-provider/pkg/transport"
)

func init() {
	RootCmd.AddCommand(collectorCmd)
}

var collectorCmd = &cobra.Command{
	Use:   "collector",
	Short: "starts the collector",
	Long:  "Will poll GitHub for changes and notify agents with updates.",
	Run: func(cmd *cobra.Command, args []string) {
		wg := &sync.WaitGroup{}

		// start ticking
		ticker := time.NewTicker(time.Second * 60)
		tickChannel := ticker.C

		// handle interrupt
		sigChannel := make(chan os.Signal, 1)
		signal.Notify(sigChannel, os.Interrupt)
		go func() {
			<-sigChannel
			simplelog.Info("Shutdown started, waiting for goroutines to return")
			ticker.Stop()
			wg.Wait()

			simplelog.Info("Shutdown complete, exiting now")
			os.Exit(0)
		}()

		collectAndPublishKeys(wg)

		for {
			<-tickChannel
			collectAndPublishKeys(wg)
		}
	},
}

func collectAndPublishKeys(wg *sync.WaitGroup) {
	wg.Add(1)

	simplelog.Info("Starting key collection")

	go func() {
		defer wg.Done()

		kc := collector.NewKeyCollector(viper.GetString("githubAccessToken"))
		teamMembers, err := kc.GetTeamMemberInfo(viper.GetString("organizationName"), viper.GetString("teamName"))
		if err != nil {
			simplelog.Info("Key collection failed: %v", err)
			return
		}

		authorizedKeysSnippet, err := authorizedkeys.GenerateSnippet(teamMembers)
		if err != nil {
			simplelog.Info("Template generation failed: %v", err)
			return
		}

		rt := transport.NewRedisTransporter(viper.GetString("redisHost"), viper.GetString("redisChannel"))
		err = rt.Publish(authorizedKeysSnippet)
		if err != nil {
			simplelog.Info("Could not publish to redis: %v", err)
		}
	}()
}
