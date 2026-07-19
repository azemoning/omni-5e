package postgres

import (
	"context"
	"fmt"

	"github.com/azemoning/omni-5e/internal/domain"
)

// --- SRD Version ---

func (s *Store) GetSRDVersion(ctx context.Context, version string) (*domain.SRDVersion, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, version, release_date, source_url, license, is_default FROM srd_versions WHERE version = $1`, version)
	v := &domain.SRDVersion{}
	err := row.Scan(&v.ID, &v.Version, &v.ReleaseDate, &v.SourceURL, &v.License, &v.IsDefault)
	if err != nil {
		return nil, fmt.Errorf("srd version %s: %w", version, err)
	}
	return v, nil
}

func (s *Store) GetDefaultSRDVersion(ctx context.Context) (*domain.SRDVersion, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, version, release_date, source_url, license, is_default FROM srd_versions WHERE is_default = true LIMIT 1`)
	v := &domain.SRDVersion{}
	err := row.Scan(&v.ID, &v.Version, &v.ReleaseDate, &v.SourceURL, &v.License, &v.IsDefault)
	if err != nil {
		return nil, fmt.Errorf("default srd version: %w", err)
	}
	return v, nil
}

func (s *Store) ListSRDVersions(ctx context.Context) ([]domain.SRDVersion, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, version, release_date, source_url, license, is_default FROM srd_versions ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []domain.SRDVersion
	for rows.Next() {
		var v domain.SRDVersion
		if err := rows.Scan(&v.ID, &v.Version, &v.ReleaseDate, &v.SourceURL, &v.License, &v.IsDefault); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (s *Store) UpsertSRDVersion(ctx context.Context, v *domain.SRDVersion) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO srd_versions (id, version, release_date, source_url, license, is_default)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (version) DO UPDATE SET
		   release_date = EXCLUDED.release_date, source_url = EXCLUDED.source_url,
		   license = EXCLUDED.license, is_default = EXCLUDED.is_default, updated_at = NOW()`,
		v.ID, v.Version, v.ReleaseDate, v.SourceURL, v.License, v.IsDefault)
	return err
}
