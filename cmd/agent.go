package cmd

import (
	"os"
	"os/signal"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/utilitywarehouse/github-sshkey-provider/pkg/simplelog"
	"github.com/utilitywarehouse/github-sshkey-provider/pkg/transport"
)

func init() {
	RootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "starts the agent",
	Long:  "Will listen for notifications from the collector and adjust the authorized_keys file.",
	Run: func(cmd *cobra.Command, args []string) {
		wg := &sync.WaitGroup{}

		rt := transport.NewRedisTransporter(viper.GetString("redisHost"), viper.GetString("redisChannel"))

		// handle interrupt
		sigChannel := make(chan os.Signal, 1)
		signal.Notify(sigChannel, os.Interrupt)
		go func() {
			<-sigChannel
			simplelog.Info("Shutdown started, waiting for goroutines to return")
			rt.StopListening()
		}()

		updateAuthorizedKeys(wg, rt)

		wg.Wait()

		simplelog.Info("Shutdown complete, exiting now")

		os.Exit(0)
	},
}

func updateAuthorizedKeys(wg *sync.WaitGroup, rt *transport.RedisTransporter) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := rt.Listen(func(message string) error {
			// XXX process message here
			return nil
		})

		if err != nil {
			simplelog.Info("Listen error: %v", err)
		}
	}()
}
