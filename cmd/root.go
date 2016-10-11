package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

const (
	appName       = "GitHub SSH Key Provider"
	appVersion    = "0.4"
	binaryName    = "gskp"
	confEnvPrefix = "GSKP"
)

var (
	// RootCmd is the root command for cobra.
	RootCmd = &cobra.Command{
		Use:   binaryName,
		Short: appName,
		Long:  "Manages authorized_keys files based on GitHub team membership.",
		Run: func(cmd *cobra.Command, args []string) {
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
	environmentName := os.Getenv("UW_ENVIRONMENT")
	if environmentName == "" {
		environmentName = "default"
	}

	viper.SetConfigName(environmentName)
	viper.AddConfigPath("./conf")

	viper.SetEnvPrefix(confEnvPrefix)
	viper.AutomaticEnv()

	viper.BindPFlag("debugLog", RootCmd.PersistentFlags().Lookup("debug"))

	viper.ReadInConfig()
}

func setConfigDefaults() {
	viper.SetDefault("collectorHTTPTimeout", 10)
	viper.SetDefault("collectorHTTPAddress", ":3000")
	viper.SetDefault("collectorCacheTTL", 300)

	viper.SetDefault("collectorRootURL", "http://localhost:3000/")
	viper.SetDefault("authorizedKeysPath", "authorized_keys")
}
