package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/azemoning/omni-5e/internal/config"
	"github.com/azemoning/omni-5e/internal/service"
	"github.com/azemoning/omni-5e/internal/store/postgres"
	httpserver "github.com/azemoning/omni-5e/internal/transport/http"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		logLevel, _ := zerolog.ParseLevel(viper.GetString("log.level"))
		zerolog.SetGlobalLevel(logLevel)
		log := zerolog.New(os.Stdout).With().Timestamp().Str("component", "http").Logger()

		// Connect to database
		ctx := context.Background()
		pool, err := pgxpool.New(ctx, cfg.Database.DSN())
		if err != nil {
			return fmt.Errorf("connecting to database: %w", err)
		}
		defer pool.Close()

		// Create stores
		store := postgres.New(pool)

		// Create service
		svc := service.New(
			store, store, store, store, store, store,
			store, store, store, store, store, store,
			store, store,
		)

		srv := httpserver.NewServer(cfg, log, svc)
		return srv.Start()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Int("port", 8080, "HTTP server port")
	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
}
