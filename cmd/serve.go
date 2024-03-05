package cmd

import (
	"github.com/gg-mike/ccli/pkg/engine/k8s"
	"github.com/gg-mike/ccli/pkg/serve"
	"github.com/gg-mike/ccli/pkg/vault"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "HTTP CI/CD server",
	Run: func(cmd *cobra.Command, args []string) {
		if validateScheduler(cmd) != nil {
			return
		}

		flags := serve.Flags{
			Address: viper.GetString(ADDRESS),
			DbUrl:   viper.GetString(DB_URL),
			Vault: vault.Config{
				Url:   viper.GetString(VAULT_URL),
				Token: viper.GetString(VAULT_TOKEN),
			},
			Scheduler: viper.GetString(SCHEDULER),
			K8s: k8s.Config{
				Mode:      viper.GetString(K8S_MODE),
				Config:    viper.GetString(K8S_CONFIG),
				Namespace: viper.GetString(K8S_NAMESPACE),
			},
		}

		handler := serve.NewHandler(logger, &flags)

		handler.Run()

		handler.Shutdown()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().String(ADDRESS, ":8080", "listen host:port for HTTP server")

	serveCmd.Flags().String(DB_URL, "", "database connection URL")
	serveCmd.MarkFlagRequired(DB_URL)

	serveCmd.Flags().String(VAULT_URL, "", "vault connection URL")
	serveCmd.MarkFlagRequired(VAULT_URL)

	addSchedulerFlag(serveCmd)
}
