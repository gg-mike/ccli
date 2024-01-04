package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gg-mike/ccli/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	logger  log.Logger
	appName = "ccli"

	rootCmd = &cobra.Command{
		Use:   appName,
		Short: "Command line tool for CI/CD workflow",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger = log.NewLogger(cmd.Use, "main", viper.GetString("log.level"), viper.GetString("log.dir"))
		},
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("log.level", "info", "log filtering level")
	rootCmd.PersistentFlags().String("log.dir", "", "log save location")
	rootCmd.PersistentFlags().StringP("config", "f", "", "location of config file")
}

func initConfig() {
	viper.SetEnvPrefix(fmt.Sprintf("%s_", strings.ToUpper(appName)))
	cfgFlag, _ := rootCmd.PersistentFlags().GetString("config")
	if cfgFlag != "" {
		viper.SetConfigFile(cfgFlag)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yml")
		viper.AddConfigPath(fmt.Sprintf("/etc/%s", appName))
		viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", appName))
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	postInitCommands(rootCmd.Commands())
}

func postInitCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		presetRequiredFlags(cmd)
		if cmd.HasSubCommands() {
			postInitCommands(cmd.Commands())
		}
	}
}

func presetRequiredFlags(cmd *cobra.Command) {
	viper.BindPFlags(cmd.Flags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "config" {
			return
		}
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}
