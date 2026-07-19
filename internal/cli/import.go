package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/azemoning/omni-5e/internal/config"
	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/ingest"
	_ "github.com/azemoning/omni-5e/internal/ingest/srd521" // register srd521 parser
	"github.com/azemoning/omni-5e/internal/store/postgres"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import SRD content into the database",
	Long:  `Parse SRD markdown, snapshot to JSON, and load into PostgreSQL.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		source, _ := cmd.Flags().GetString("source")
		version, _ := cmd.Flags().GetString("version")
		if source == "" || version == "" {
			return fmt.Errorf("--source and --version are required")
		}

		log := zerolog.New(os.Stdout).With().Timestamp().Str("component", "import").Logger()

		// Get parser for this version
		parser := ingest.Get(version)
		if parser == nil {
			return fmt.Errorf("no parser registered for SRD version %s", version)
		}

		ctx := context.Background()

		// Connect to database
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		pool, err := pgxpool.New(ctx, cfg.Database.DSN())
		if err != nil {
			return fmt.Errorf("connecting to database: %w", err)
		}
		defer pool.Close()

		store := postgres.New(pool)

		// Ensure SRD version exists
		log.Info().Str("version", version).Msg("ensuring SRD version exists")
		srdVersion := &domain.SRDVersion{
			ID:        uuid.New(),
			Version:   version,
			SourceURL: "https://github.com/downfallx/dnd-5e-srd-markdown",
			License:   "CC-BY-4.0",
			IsDefault: true,
		}
		if err := store.UpsertSRDVersion(ctx, srdVersion); err != nil {
			return fmt.Errorf("upserting SRD version: %w", err)
		}

		// Parse and load each entity type
		entities := []struct {
			name    string
			parseFn func() (any, error)
			loadFn  func(any) error
		}{
			{"spells", func() (any, error) { return parser.ParseSpells(ctx, source) }, func(d any) error {
				spells := d.([]domain.Spell)
				for i := range spells {
					if spells[i].ID == uuid.Nil {
						spells[i].ID = uuid.New()
					}
				}
				return store.UpsertSpells(ctx, spells)
			}},
			{"monsters", func() (any, error) { return parser.ParseMonsters(ctx, source) }, func(d any) error {
				monsters := d.([]domain.Monster)
				for i := range monsters {
					if monsters[i].ID == uuid.Nil {
						monsters[i].ID = uuid.New()
					}
				}
				return store.UpsertMonsters(ctx, monsters)
			}},
			{"classes", func() (any, error) { return parser.ParseClasses(ctx, source) }, func(d any) error {
				for _, c := range d.([]domain.Class) {
					cl := c
					if cl.ID == uuid.Nil {
						cl.ID = uuid.New()
					}
					if err := store.UpsertClass(ctx, &cl); err != nil {
						return err
					}
				}
				return nil
			}},
			{"species", func() (any, error) { return parser.ParseSpecies(ctx, source) }, func(d any) error {
				for _, s := range d.([]domain.Species) {
					sp := s
					if sp.ID == uuid.Nil {
						sp.ID = uuid.New()
					}
					if err := store.UpsertSpecies(ctx, &sp); err != nil {
						return err
					}
				}
				return nil
			}},
			{"backgrounds", func() (any, error) { return parser.ParseBackgrounds(ctx, source) }, func(d any) error {
				for _, b := range d.([]domain.Background) {
					bg := b
					if bg.ID == uuid.Nil {
						bg.ID = uuid.New()
					}
					if err := store.UpsertBackground(ctx, &bg); err != nil {
						return err
					}
				}
				return nil
			}},
			{"feats", func() (any, error) { return parser.ParseFeats(ctx, source) }, func(d any) error {
				for _, f := range d.([]domain.Feat) {
					feat := f
					if feat.ID == uuid.Nil {
						feat.ID = uuid.New()
					}
					if err := store.UpsertFeat(ctx, &feat); err != nil {
						return err
					}
				}
				return nil
			}},
			{"equipment", func() (any, error) { return parser.ParseEquipment(ctx, source) }, func(d any) error {
				for _, e := range d.([]domain.Equipment) {
					eq := e
					if eq.ID == uuid.Nil {
						eq.ID = uuid.New()
					}
					if err := store.UpsertEquipment(ctx, &eq); err != nil {
						return err
					}
				}
				return nil
			}},
			{"magic_items", func() (any, error) { return parser.ParseMagicItems(ctx, source) }, func(d any) error {
				for _, m := range d.([]domain.MagicItem) {
					mi := m
					if mi.ID == uuid.Nil {
						mi.ID = uuid.New()
					}
					if err := store.UpsertMagicItem(ctx, &mi); err != nil {
						return err
					}
				}
				return nil
			}},
			{"conditions", func() (any, error) { return parser.ParseConditions(ctx, source) }, func(d any) error {
				for _, c := range d.([]domain.Condition) {
					cond := c
					if cond.ID == uuid.Nil {
						cond.ID = uuid.New()
					}
					if err := store.UpsertCondition(ctx, &cond); err != nil {
						return err
					}
				}
				return nil
			}},
			{"glossary_terms", func() (any, error) { return parser.ParseGlossaryTerms(ctx, source) }, func(d any) error {
				for _, g := range d.([]domain.GlossaryTerm) {
					term := g
					if term.ID == uuid.Nil {
						term.ID = uuid.New()
					}
					if err := store.UpsertGlossaryTerm(ctx, &term); err != nil {
						return err
					}
				}
				return nil
			}},
			{"rule_sections", func() (any, error) { return parser.ParseRuleSections(ctx, source) }, func(d any) error {
				for _, r := range d.([]domain.RuleSection) {
					rs := r
					if rs.ID == uuid.Nil {
						rs.ID = uuid.New()
					}
					if err := store.UpsertRuleSection(ctx, &rs); err != nil {
						return err
					}
				}
				return nil
			}},
		}

		// Create snapshot directory
		snapshotDir := filepath.Join("data", "parsed", version)
		os.MkdirAll(snapshotDir, 0o755)

		for _, e := range entities {
			log.Info().Str("entity", e.name).Msg("parsing")
			data, err := e.parseFn()
			if err != nil {
				log.Warn().Err(err).Str("entity", e.name).Msg("parse failed (may be missing file)")
				continue
			}

			// Snapshot to JSON
			snapshotPath := filepath.Join(snapshotDir, e.name+".json")
			jsonData, _ := json.MarshalIndent(data, "", "  ")
			os.WriteFile(snapshotPath, jsonData, 0o644)
			log.Info().Str("path", snapshotPath).Msg("wrote snapshot")

			// Load into database
			log.Info().Str("entity", e.name).Msg("loading into database")
			if err := e.loadFn(data); err != nil {
				log.Error().Err(err).Str("entity", e.name).Msg("load failed")
				continue
			}
			log.Info().Str("entity", e.name).Msg("loaded successfully")
		}

		log.Info().Msg("import complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().String("source", "", "source directory containing SRD markdown files")
	importCmd.Flags().String("version", "", "SRD content version (e.g. 5.2.1)")
	importCmd.MarkFlagRequired("source")
	importCmd.MarkFlagRequired("version")
}
