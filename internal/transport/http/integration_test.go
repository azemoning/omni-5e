package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/azemoning/omni-5e/internal/config"
	"github.com/azemoning/omni-5e/internal/service"
	"github.com/azemoning/omni-5e/internal/store/postgres"
	httpserver "github.com/azemoning/omni-5e/internal/transport/http"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		Env:          map[string]string{"POSTGRES_USER": "test", "POSTGRES_PASSWORD": "test", "POSTGRES_DB": "test"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("host=%s port=%s user=test password=test dbname=test sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	migrationSQL := `
	CREATE EXTENSION IF NOT EXISTS pgcrypto;
	CREATE TABLE srd_versions (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), version TEXT UNIQUE NOT NULL, release_date DATE, source_url TEXT NOT NULL, license TEXT NOT NULL DEFAULT 'CC-BY-4.0', is_default BOOLEAN NOT NULL DEFAULT false, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
	CREATE TABLE spells (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, level SMALLINT NOT NULL, school TEXT NOT NULL, casting_time TEXT, range TEXT, duration TEXT, concentration BOOLEAN NOT NULL DEFAULT false, ritual BOOLEAN NOT NULL DEFAULT false, components JSONB NOT NULL DEFAULT '{}', description TEXT NOT NULL, at_higher_levels TEXT, search TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', name || ' ' || description)) STORED, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE spell_classes (spell_id UUID REFERENCES spells(id) ON DELETE CASCADE, class_slug TEXT NOT NULL, PRIMARY KEY (spell_id, class_slug));
	CREATE TABLE monsters (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, size TEXT, type TEXT, alignment TEXT, ac_value SMALLINT, ac_source TEXT, hp_avg INTEGER, hp_formula TEXT, cr NUMERIC(4,2), xp INTEGER, category TEXT NOT NULL DEFAULT 'monster', speed JSONB, ability_scores JSONB, saving_throws JSONB, skills JSONB, damage_resistances JSONB, damage_immunities JSONB, damage_vulnerabilities JSONB, condition_immunities JSONB, senses JSONB, languages JSONB, environment JSONB, traits JSONB, actions JSONB, bonus_actions JSONB, reactions JSONB, legendary_actions JSONB, search TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', name)) STORED, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE classes (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, hit_die SMALLINT NOT NULL, primary_ability TEXT, saving_throw_proficiencies JSONB, armor_proficiencies JSONB, weapon_proficiencies JSONB, description TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE subclasses (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, class_slug TEXT NOT NULL, description TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE class_features (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, class_slug TEXT NOT NULL, subclass_slug TEXT, level SMALLINT NOT NULL, description TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE class_level_tables (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), class_slug TEXT NOT NULL, level SMALLINT NOT NULL, proficiency_bonus SMALLINT NOT NULL, features_unlocked JSONB, other_columns JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, class_slug, level));
	CREATE TABLE species (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, size TEXT, speed SMALLINT, traits JSONB, description TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE backgrounds (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, ability_score_options JSONB, skill_proficiencies JSONB, granted_feat_slug TEXT, equipment JSONB, description TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE feats (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, category TEXT, prerequisite TEXT, description TEXT NOT NULL, repeatable BOOLEAN NOT NULL DEFAULT false, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE equipment (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, category TEXT NOT NULL, cost TEXT, weight NUMERIC(8,2), properties JSONB, description TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE magic_items (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, rarity TEXT NOT NULL, requires_attunement BOOLEAN NOT NULL DEFAULT false, type TEXT, description TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE conditions (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, description TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE glossary_terms (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, category TEXT, definition TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	CREATE TABLE rule_sections (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), srd_version_id UUID NOT NULL REFERENCES srd_versions(id), slug TEXT NOT NULL, name TEXT NOT NULL, source_file TEXT, heading_path JSONB, body TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE (srd_version_id, slug));
	`
	_, err = pool.Exec(ctx, migrationSQL)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
		container.Terminate(ctx)
	}
	return pool, cleanup
}

func seedTestData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	srdID := uuid.New()
	_, err := pool.Exec(ctx,
		`INSERT INTO srd_versions (id, version, source_url, license, is_default) VALUES ($1, '5.2.1', 'https://example.com', 'CC-BY-4.0', true)`,
		srdID)
	require.NoError(t, err)

	for _, s := range []struct {
		slug, name, school string
		level              int
	}{
		{"fireball", "Fireball", "evocation", 3},
		{"wish", "Wish", "conjuration", 9},
		{"healing-word", "Healing Word", "abjuration", 1},
	} {
		_, err := pool.Exec(ctx,
			`INSERT INTO spells (id, srd_version_id, slug, name, level, school, casting_time, range, duration, components, description, at_higher_levels) VALUES ($1, $2, $3, $4, $5, $6, 'Action', 'Self', 'Instantaneous', '{}', 'Test spell', '')`,
			uuid.New(), srdID, s.slug, s.name, s.level, s.school)
		require.NoError(t, err)
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO monsters (id, srd_version_id, slug, name, size, type, alignment, ac_value, hp_avg, cr, xp, category) VALUES ($1, $2, 'beholder', 'Beholder', 'Large', 'Aberration', 'Lawful Evil', 18, 180, 13, 10000, 'monster')`,
		uuid.New(), srdID)
	require.NoError(t, err)
}

// getFreePort returns an available port.
func getFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

func TestIntegrationSpellEndpoints(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	seedTestData(t, pool)

	store := postgres.New(pool)
	svc := service.New(store, store, store, store, store, store, store, store, store, store, store, store, store, store)

	log := zerolog.Nop()
	port := getFreePort(t)
	cfg := &config.Config{
		Server: config.ServerConfig{Host: "127.0.0.1", Port: port},
	}
	srv := httpserver.NewServer(cfg, log, svc)

	go srv.Start()
	time.Sleep(200 * time.Millisecond)
	defer srv.Shutdown()

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	resp, err := http.Get(baseURL + "/api/v1/spells")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Logf("unexpected status %d, body: %s", resp.StatusCode, string(body))
	}
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]any
	json.Unmarshal(body, &result)
	assert.NotNil(t, result["data"])
	assert.NotNil(t, result["meta"])

	resp, err = http.Get(baseURL + "/api/v1/spells/fireball")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Logf("fireball status %d, body: %s", resp.StatusCode, string(body))
	}
	assert.Equal(t, 200, resp.StatusCode)

	json.Unmarshal(body, &result)
	data := result["data"].(map[string]any)
	assert.Equal(t, "Fireball", data["name"])
	assert.Equal(t, float64(3), data["level"])

	resp, err = http.Get(baseURL + "/api/v1/spells/nonexistent")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 404, resp.StatusCode)

	resp, err = http.Get(baseURL + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestIntegrationLicenseEndpoint(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	store := postgres.New(pool)
	svc := service.New(store, store, store, store, store, store, store, store, store, store, store, store, store, store)

	log := zerolog.Nop()
	port := getFreePort(t)
	cfg := &config.Config{Server: config.ServerConfig{Host: "127.0.0.1", Port: port}}
	srv := httpserver.NewServer(cfg, log, svc)
	go srv.Start()
	time.Sleep(200 * time.Millisecond)
	defer srv.Shutdown()

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	resp, err := http.Get(baseURL + "/api/v1/license")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(body, &result)
	assert.Equal(t, "CC-BY-4.0", result["license"])
	assert.Contains(t, result["text"].(string), "Wizards of the Coast")
}
