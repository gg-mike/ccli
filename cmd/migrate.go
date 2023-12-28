package cmd

import (
	"github.com/gg-mike/ccli/pkg/migrate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate model schemas to database",
	Run: func(cmd *cobra.Command, args []string) {
		if validateScheduler(cmd) != nil {
			return
		}

		flags := migrate.Flags{
			DbUrl:     viper.GetString(DB_URL),
			Scheduler: viper.GetString(SCHEDULER),
		}

		handler := migrate.NewHandler(logger, &flags)

		if err := handler.Run(); err != nil {
			logger.Fatal().Err(err).Send()
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().String(DB_URL, "", "database connection URL")
	migrateCmd.MarkFlagRequired(DB_URL)

	addSchedulerFlag(migrateCmd)
}
