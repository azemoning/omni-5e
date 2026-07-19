package cli

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/azemoning/omni-5e/internal/config"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Manage database schema migrations (up, down, status).`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		db, err := sql.Open("pgx", cfg.Database.DSN())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Close()

		goose.SetBaseFS(os.DirFS("migrations"))
		if err := goose.Up(db, "."); err != nil {
			return fmt.Errorf("migrate up: %w", err)
		}
		fmt.Println("Migrations applied successfully")
		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back the last migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		db, err := sql.Open("pgx", cfg.Database.DSN())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Close()

		goose.SetBaseFS(os.DirFS("migrations"))
		if err := goose.Down(db, "."); err != nil {
			return fmt.Errorf("migrate down: %w", err)
		}
		fmt.Println("Migration rolled back successfully")
		return nil
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		db, err := sql.Open("pgx", cfg.Database.DSN())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Close()

		goose.SetBaseFS(os.DirFS("migrations"))
		if err := goose.Status(db, "."); err != nil {
			return fmt.Errorf("migrate status: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd, migrateDownCmd, migrateStatusCmd)
}
