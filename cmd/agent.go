package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
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
		simplelog.Infof("Starting up [agentRecoverInterval=%d]", viper.GetInt("agentRecoverInterval"))

		rt := gskp.NewRedisTransporter(
			viper.GetString("redisHost"),
			viper.GetString("redisPassword"),
			viper.GetString("redisChannel"),
		)

		isActive := true
		timer := time.NewTimer(time.Nanosecond)

		// handle interrupt
		sigChannel := make(chan os.Signal, 1)
		signal.Notify(sigChannel, os.Interrupt)
		go func() {
			<-sigChannel
			simplelog.Infof("Shutdown started, waiting for processes to return")
			isActive = false
			timer.Reset(time.Nanosecond)
			rt.StopListening()
		}()

		getBootstrapSnippet()

		<-timer.C
		for isActive {
			if err := rt.Listen(func(message string) error {
				updateAuthorizedKeys(message)

				return nil
			}); err != nil {
				simplelog.Infof("Listen returned error: %v", err)
			}

			if isActive {
				simplelog.Infof("Waiting %d seconds before trying to establish a connection again",
					viper.GetInt("agentRecoverInterval"))

				timer.Reset(time.Duration(viper.GetInt("agentRecoverInterval")) * time.Second)
			}

			<-timer.C
		}

		simplelog.Infof("Shutdown complete, exiting now")
	},
}

func getBootstrapSnippet() {
	simplelog.Infof("Trying to get an initial version of the snippet from %s", viper.GetString("agentBootstrapURL"))

	resp, err := http.Get(viper.GetString("agentBootstrapURL"))
	if err != nil {
		simplelog.Infof("Could not reach bootstrap URL, ignoring error: %v", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		simplelog.Infof("Error occurred while trying to read the response from the boostrap URL, ignoring: %v", err)
		return
	}

	data := gskp.HTTPResponse{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		simplelog.Infof("Error occurred while trying to decode the response from the boostrap URL, ignoring: %v", err)
		return
	}

	updateAuthorizedKeys(data["authorized_keys"].(string))
}

func updateAuthorizedKeys(snippet string) {
	simplelog.Infof("Updating %s", viper.GetString("authorizedKeysPath"))

	err := gskp.AuthorizedKeys.Update(viper.GetString("authorizedKeysPath"), snippet)
	if err == gskp.ErrAuthorizedKeysNotChanged {
		simplelog.Infof("The authorized_keys snippet makes no changes to the file, ignoring")
	} else if err != nil {
		simplelog.Infof("Error occurred while trying to update '%s': %v",
			viper.GetString("authorizedKeysPath"), err)
	}
}
