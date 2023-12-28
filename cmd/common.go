package cmd

import (
	"errors"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func validateScheduler(cmd *cobra.Command) error {
	scheduler := viper.GetString(SCHEDULER)
	if !slices.Contains([]string{"standalone", "k8s"}, scheduler) {
		logger.Error().Msgf(`flag "scheduler" has invalid value [%s]`, scheduler)
		cmd.Help()
		return errors.New("invalid")
	}
	return nil
}

func addSchedulerFlag(cmd *cobra.Command) {
	cmd.Flags().String(SCHEDULER, "standalone", "scheduler type (standalone or k8s)")
}
