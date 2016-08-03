package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

const (
	appName       = "GitHub SSH Key Provider"
	appVersion    = "0.2"
	binaryName    = "gskp"
	confEnvPrefix = "GSKP"
)

var (
	// RootCmd is the root command for cobra
	RootCmd = &cobra.Command{
		Use:   binaryName,
		Short: appName,
		Long:  "Manages authorized_keys files based on GitHub team membership.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(viper.GetBool("debugLog"))
			cmd.Help()
		},
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "prints the version",
		Long:  "Prints the version.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s v%s\n", appName, appVersion)
		},
	}
)

func init() {
	RootCmd.AddCommand(versionCmd)
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "debug output")

	cobra.OnInitialize(initConfig, setConfigDefaults, func() {
		simplelog.DebugEnabled = viper.GetBool("debugLog")
	})
}

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./conf")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix(confEnvPrefix)
	viper.AutomaticEnv()

	viper.BindPFlag("debugLog", RootCmd.PersistentFlags().Lookup("debug"))

	viper.ReadInConfig()
}

func setConfigDefaults() {
	viper.SetDefault("redisHost", ":6379")
	viper.SetDefault("redisPassword", "")

	viper.Set("redisChannel", "gskp")
	viper.Set("redisCacheDB", "9")
}
