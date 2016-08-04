package cmd

import (
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/transporter"
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

		rt := transporter.NewRedis(
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

		<-timer.C
		for isActive {
			if err := rt.Listen(func(message string) error {
				simplelog.Infof("Updating %s", viper.GetString("authorizedKeysPath"))

				err := gskp.AuthorizedKeys.Update(viper.GetString("authorizedKeysPath"), message)
				if err != nil {
					simplelog.Infof("Error occured while trying to update '%s': %v",
						viper.GetString("authorizedKeysPath"), err)
				}

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

func updateAuthorizedKeys(rt *transporter.Redis) {

}
