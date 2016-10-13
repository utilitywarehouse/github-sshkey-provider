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
	RootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "starts the agent",
	Long:  "Will listen for notifications from the collector and adjust the authorized_keys file.",
	Run: func(cmd *cobra.Command, args []string) {
		simplelog.Infof("starting up")

		for _, cv := range []string{"agentGithubTeam", "collectorBaseURL", "authorizedKeysPath"} {
			if viper.GetString(cv) == "" {
				simplelog.Errorf("please specify a config value for %s", cv)
				os.Exit(-1)
			}
		}

		// handle interrupt
		sigChannel := make(chan os.Signal, 1)
		signal.Notify(sigChannel, os.Interrupt)
		go func() {
			<-sigChannel
			simplelog.Infof("received interrupt: shutting down")
			os.Exit(0)
		}()

		client, err := gskp.NewClient(viper.GetString("collectorBaseURL"), viper.GetInt64("agentLongpollTimeoutSeconds"))
		if err != nil {
			simplelog.Errorf("could not create a client instance: %v", err)
		}

		data, err := client.GetKeys(viper.GetString("agentGithubTeam"))
		if err != nil {
			simplelog.Errorf("error while trying to bootstrap with initial keys, ignoring: %v", err)
		} else {
			updateAuthorizedKeys(data)
		}

		simplelog.Infof("starting poll for ssh key updates")
		for {
			simplelog.Debugf("starting longpoll request")
			data, err := client.PollForKeys(viper.GetString("agentGithubTeam"))
			if err == gskp.ErrClientPollTimeout {
				simplelog.Debugf("longpoll timeout, will re-start")
				continue
			} else if err != nil {
				simplelog.Errorf("error while polling for key changes, ignoring and retrying in 15 seconds: %v", err)
				time.Sleep(15 * time.Second)
			} else {
				updateAuthorizedKeys(data)
			}
		}
	},
}

func updateAuthorizedKeys(data []gskp.UserInfo) {
	simplelog.Infof("updating %s", viper.GetString("authorizedKeysPath"))

	snippet, err := gskp.AuthorizedKeys.GenerateSnippet(data)
	if err != nil {
		simplelog.Errorf("could not generate authorized_keys snippet: %v", err)
	}

	err = gskp.AuthorizedKeys.Update(viper.GetString("authorizedKeysPath"), snippet)
	if err == gskp.ErrAuthorizedKeysNotChanged {
		simplelog.Infof("the authorized_keys snippet makes no changes to the file, ignoring")
	} else if err != nil {
		simplelog.Errorf("error occurred while trying to update '%s': %v", viper.GetString("authorizedKeysPath"), err)
	}
}
