package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplecache"
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
		simplelog.Infof("Starting up [collectorPollingInterval=%d]", viper.GetInt("collectorPollingInterval"))

		wg := &sync.WaitGroup{}

		// start ticking
		ticker := time.NewTicker(time.Duration(viper.GetInt("collectorPollingInterval")) * time.Second)

		// create http server
		httpServer := gskp.NewHTTPServer()

		// handle interrupt
		sigChannel := make(chan os.Signal, 1)
		signal.Notify(sigChannel, os.Interrupt)
		go func() {
			<-sigChannel
			simplelog.Infof("Shutdown started, waiting for goroutines to return")
			ticker.Stop()
			httpServer.StopListening(viper.GetInt("collectorHTTPTimeout"))
			wg.Wait()

			simplelog.Infof("Shutdown complete, exiting now")
			os.Exit(0)
		}()

		// start http server
		wg.Add(1)
		go func() {
			defer wg.Done()

			httpServer.Listen(viper.GetString("collectorHTTPAddress"), viper.GetInt("collectorHTTPTimeout"))
		}()

		// collection loop
		teamID := findTeamID()
		setupAuthorizedKeysHTTPEndpoint(httpServer, teamID)

		for {
			wg.Add(1)

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
		simplelog.Errorf("Error occurred when trying to find the team's ID: %v", err)
		os.Exit(1)
	}

	simplelog.Infof("Found team ID: %d", ti)

	return ti
}

func collectAndPublishKeys(teamID int) {
	simplelog.Infof("Starting key collection")

	kc := gskp.NewKeyCollector(viper.GetString("githubAccessToken"))
	teamMembers, err := kc.GetTeamMemberInfo(teamID)
	if err != nil {
		simplelog.Infof("Key collection failed: %v", err)
		return
	}
	simplelog.Infof("Key collection completed for %d users", len(teamMembers))

	teamMembersSerialised, err := teamMembers.Marshal()
	if err != nil {
		simplelog.Infof("Failed to serialise the UserInfoList. Will not use the cache but will try to publish anyway.")
	} else {
		if err := simplecache.NewRedis(
			viper.GetString("redisHost"),
			viper.GetString("redisPassword"),
			viper.GetString("redisCacheDB"),
		).Set(fmt.Sprintf("userinfolist_%d", teamID), teamMembersSerialised); err != nil {
			if err == simplecache.ErrValueHasNotChanged {
				simplelog.Infof("Team members have not changed since last time, but will publish anyway.")
			} else {
				simplelog.Errorf("Ignoring error that occurred while trying to write in the cache: %v", err)
			}
		}
	}

	authorizedKeysSnippet, err := gskp.AuthorizedKeys.GenerateSnippet(teamMembers)
	if err != nil {
		simplelog.Infof("Template generation failed, will not publish: %v", err)
		return
	}

	rt := gskp.NewRedisTransporter(
		viper.GetString("redisHost"),
		viper.GetString("redisPassword"),
		viper.GetString("redisChannel"),
	)

	err = rt.Publish(authorizedKeysSnippet)
	if err != nil {
		simplelog.Infof("Could not publish to redis: %v", err)
	}
}

func setupAuthorizedKeysHTTPEndpoint(httpServer *gskp.HTTPServer, teamID int) {
	httpServer.HandleGet("/authorized_keys", func(w http.ResponseWriter, r *http.Request) {
		value, err := simplecache.NewRedis(
			viper.GetString("redisHost"),
			viper.GetString("redisPassword"),
			viper.GetString("redisCacheDB")).Get(fmt.Sprintf("userinfolist_%d", teamID))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, gskp.HTTPResponse{"error": "unexpected error occurred"}.Marshal())
		}

		teamMembers := gskp.UserInfoList{}
		teamMembers.Unmarshal(value)

		authorizedKeysSnippet, err := gskp.AuthorizedKeys.GenerateSnippet(teamMembers)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, gskp.HTTPResponse{"error": "unexpected error occurred"}.Marshal())
		}

		fmt.Fprintf(w, gskp.HTTPResponse{"authorized_keys": authorizedKeysSnippet}.Marshal())
	})
}
