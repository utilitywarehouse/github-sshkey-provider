package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appName    = "GitHub SSH Key Provider"
	appVersion = "0.1"
	binaryName = "gskp"
)

func init() {
	RootCmd.AddCommand(versionCmd)

	viper.SetDefault("redisHost", ":6379")
	viper.Set("redisChannel", "gskp")
}

// RootCmd is the root command for cobra
var RootCmd = &cobra.Command{
	Use:   binaryName,
	Short: appName,
	Long:  "Manages authorized_keys files based on GitHub team membership.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints the version",
	Long:  "Prints the version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s v%s\n", appName, appVersion)
	},
}
