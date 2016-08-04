package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplecache"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/transporter"
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

		simplelog.Infof("Starting up [collectorPollingInterval=%d]", viper.GetInt("collectorPollingInterval"))

		// start ticking
		ticker := time.NewTicker(time.Duration(viper.GetInt("collectorPollingInterval")) * time.Second)

		// handle interrupt
		sigChannel := make(chan os.Signal, 1)
		signal.Notify(sigChannel, os.Interrupt)
		go func() {
			<-sigChannel
			simplelog.Infof("Shutdown started, waiting for goroutines to return")
			ticker.Stop()
			wg.Wait()

			simplelog.Infof("Shutdown complete, exiting now")
			os.Exit(0)
		}()

		// collection loop
		teamID := findTeamID()
		for {
			wg.Add(1)

			simplelog.Infof("Starting key collection")

			go func() {
				defer wg.Done()

				collectAndPublishKeys(teamID)
			}()

			<-ticker.C
		}
	},
}

func findTeamID() int {
	kc := gskp.NewKeyCollector(viper.GetString("githubAccessToken"))

	ti, err := kc.GetTeamID(viper.GetString("organizationName"), viper.GetString("teamName"))
	if err != nil {
		simplelog.Errorf("Error occured when trying to find the team's ID: %v", err)
		os.Exit(1)
	}

	simplelog.Infof("Found team ID: %d", ti)

	return ti
}

func collectAndPublishKeys(teamID int) {
	kc := gskp.NewKeyCollector(viper.GetString("githubAccessToken"))
	teamMembers, err := kc.GetTeamMemberInfo(teamID)
	if err != nil {
		simplelog.Infof("Key collection failed: %v", err)
		return
	}
	simplelog.Infof("Key collection completed for %d users", len(teamMembers))

	teamMembersSerialised, err := teamMembers.Marshal()
	if err != nil {
		simplelog.Infof("Failed to serialise the UserInfoList. Will not use the cache but will publish anyway.")
	} else {
		if err := simplecache.NewRedis(
			viper.GetString("redisHost"),
			viper.GetString("redisPassword"),
			viper.GetString("redisCacheDB"),
		).Set(fmt.Sprintf("userinfolist_%d", teamID), teamMembersSerialised); err != nil {
			if err == simplecache.ErrValueHasNotChanged {
				simplelog.Infof("Team members have not changed since last time, will not publish anything.")

				return
			}

			simplelog.Infof("Ignoring error that occured while trying to write in the cache: %v", err)
		}
	}

	authorizedKeysSnippet, err := gskp.AuthorizedKeys.GenerateSnippet(teamMembers)
	if err != nil {
		simplelog.Infof("Template generation failed: %v", err)
		return
	}

	rt := transporter.NewRedis(
		viper.GetString("redisHost"),
		viper.GetString("redisPassword"),
		viper.GetString("redisChannel"),
	)

	err = rt.Publish(authorizedKeysSnippet)
	if err != nil {
		simplelog.Infof("Could not publish to redis: %v", err)
	}
}
