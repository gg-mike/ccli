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
		logger.Error().Msgf("flag '%s' has invalid value [%s]", SCHEDULER, scheduler)
		cmd.Help()
		return errors.New("invalid")
	}
	mode := viper.GetString(K8S_MODE)
	if !slices.Contains([]string{"inner", "outer"}, mode) {
		logger.Error().Msgf("flag '%s' has invalid value [%s]", K8S_MODE, mode)
		cmd.Help()
		return errors.New("invalid")
	}
	return nil
}

func addSchedulerFlag(cmd *cobra.Command) {
	cmd.Flags().String(SCHEDULER, "standalone", "scheduler type (standalone or k8s)")
	cmd.Flags().String(K8S_MODE, "outer", "mode of work, in/out-cluster (inner or outer)")
	cmd.Flags().String(K8S_CONFIG, "", "k8s config filepath (else default location is used)")
	cmd.MarkFlagFilename(K8S_CONFIG)
	cmd.Flags().String(K8S_NAMESPACE, "default", "namespace were worker pods will be located")
}
