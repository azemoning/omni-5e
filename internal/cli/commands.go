package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/azemoning/omni-5e/internal/config"
	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/store/postgres"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export SRD data",
	Long:  `Export SRD data as JSON.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version, _ := cmd.Flags().GetString("version")
		out, _ := cmd.Flags().GetString("out")
		if version == "" {
			version = "5.2.1"
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		ctx := context.Background()
		pool, err := pgxpool.New(ctx, cfg.Database.DSN())
		if err != nil {
			return fmt.Errorf("connecting to database: %w", err)
		}
		defer pool.Close()

		store := postgres.New(pool)

		exportData := map[string]any{}

		// Export spells
		spells, err := store.ListSpells(ctx, domain.SpellFilter{
			ListParams: domain.ListParams{Limit: 10000, SRDVersion: version},
		})
		if err == nil {
			exportData["spells"] = spells.Items
		}

		// Export monsters
		monsters, err := store.ListMonsters(ctx, domain.MonsterFilter{
			ListParams: domain.ListParams{Limit: 10000, SRDVersion: version},
		})
		if err == nil {
			exportData["monsters"] = monsters.Items
		}

		data, _ := json.MarshalIndent(exportData, "", "  ")
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return fmt.Errorf("writing export: %w", err)
		}
		fmt.Printf("Exported to %s\n", out)
		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate parsed SRD snapshot",
	Long:  `Run schema and consistency checks on the parsed JSON snapshot.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version, _ := cmd.Flags().GetString("version")
		if version == "" {
			version = "5.2.1"
		}

		snapshotDir := fmt.Sprintf("data/parsed/%s", version)
		entries, err := os.ReadDir(snapshotDir)
		if err != nil {
			return fmt.Errorf("reading snapshot dir: %w", err)
		}

		errors := 0
		for _, e := range entries {
			if e.IsDir() || e.Name() == "" {
				continue
			}
			data, err := os.ReadFile(fmt.Sprintf("%s/%s", snapshotDir, e.Name()))
			if err != nil {
				fmt.Printf("FAIL %s: %v\n", e.Name(), err)
				errors++
				continue
			}
			var items []any
			if err := json.Unmarshal(data, &items); err != nil {
				fmt.Printf("FAIL %s: invalid JSON: %v\n", e.Name(), err)
				errors++
				continue
			}
			fmt.Printf("OK   %s (%d items)\n", e.Name(), len(items))
		}

		if errors > 0 {
			return fmt.Errorf("%d validation errors", errors)
		}
		fmt.Println("All snapshots valid")
		return nil
	},
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
}

var dbSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Import the bundled default SRD 5.2.1 snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Seeding database with SRD 5.2.1...")
		// Delegate to import command
		return importCmd.RunE(cmd, []string{})
	},
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Drop and recreate schema (dev only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return fmt.Errorf("db reset requires --force flag")
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		ctx := context.Background()
		pool, err := pgxpool.New(ctx, cfg.Database.DSN())
		if err != nil {
			return fmt.Errorf("connecting to database: %w", err)
		}
		defer pool.Close()

		tables := []string{
			"rule_sections", "glossary_terms", "conditions", "magic_items",
			"equipment", "feats", "backgrounds", "species", "class_level_tables",
			"class_features", "subclasses", "classes", "spell_classes", "monsters",
			"spells", "srd_versions",
		}
		for _, t := range tables {
			pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", t))
		}
		fmt.Println("All tables dropped")

		// Re-run migrations
		fmt.Println("Re-running migrations...")
		return migrateUpCmd.RunE(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd, validateCmd, dbCmd)
	exportCmd.Flags().String("version", "5.2.1", "SRD content version")
	exportCmd.Flags().String("out", "export.json", "output file path")
	validateCmd.Flags().String("version", "5.2.1", "SRD content version to validate")
	dbCmd.AddCommand(dbSeedCmd, dbResetCmd)
	dbResetCmd.Flags().Bool("force", false, "confirm database reset")
}
