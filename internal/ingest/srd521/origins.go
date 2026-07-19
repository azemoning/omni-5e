package srd521

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/ingest/shared"
	"github.com/yuin/goldmark/ast"
)

// ParseSpecies parses species from character-origins.md.
// Species are at heading level 4 under "### Species Descriptions".
func (p *Parser) ParseSpecies(ctx context.Context, sourceDir string) ([]domain.Species, error) {
	path := filepath.Join(sourceDir, "character-origins.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var species []domain.Species
	var current *domain.Species
	inSpeciesSection := false

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}
			lower := strings.ToLower(name)

			if strings.Contains(lower, "species descriptions") {
				inSpeciesSection = true
				continue
			}
			// Stop at next major section
			if n.Level <= 2 && inSpeciesSection {
				inSpeciesSection = false
			}

			// Species entries are at level 4 within the species section
			if n.Level == 4 && inSpeciesSection {
				if current != nil {
					species = append(species, *current)
				}
				current = &domain.Species{
					BaseEntity: domain.BaseEntity{
						Name:       name,
						Slug:       shared.Slugify(name),
						SRDVersion: p.Version(),
					},
				}
			}

		case *ast.Paragraph:
			if current == nil || !inSpeciesSection {
				continue
			}
			text := shared.ExtractText(n, src)

			// Parse structured fields - may be combined in one paragraph
			if strings.Contains(text, "Creature Type:") || strings.Contains(text, "Size:") || strings.Contains(text, "Speed:") {
				parseSpeciesFields(current, text)
			} else {
				if current.Description == "" {
					current.Description = text
				} else {
					current.Description += "\n\n" + text
				}
			}
		}
	}

	if current != nil {
		species = append(species, *current)
	}

	return species, nil
}

// ParseBackgrounds parses backgrounds from character-origins.md.
// Backgrounds are at heading level 4 under "### Background Descriptions".
func (p *Parser) ParseBackgrounds(ctx context.Context, sourceDir string) ([]domain.Background, error) {
	path := filepath.Join(sourceDir, "character-origins.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var backgrounds []domain.Background
	var current *domain.Background
	inBackgroundsSection := false

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}
			lower := strings.ToLower(name)

			if strings.Contains(lower, "background descriptions") {
				inBackgroundsSection = true
				continue
			}
			// Stop at next major section (species)
			if strings.Contains(lower, "character species") || strings.Contains(lower, "species descriptions") {
				inBackgroundsSection = false
			}

			// Background entries are at level 4 within the backgrounds section
			if n.Level == 4 && inBackgroundsSection {
				if current != nil {
					backgrounds = append(backgrounds, *current)
				}
				current = &domain.Background{
					BaseEntity: domain.BaseEntity{
						Name:       name,
						Slug:       shared.Slugify(name),
						SRDVersion: p.Version(),
					},
				}
			}

		case *ast.Paragraph:
			if current == nil || !inBackgroundsSection {
				continue
			}
			text := shared.ExtractText(n, src)

			// Parse structured fields - may be combined in one paragraph
			if strings.Contains(text, "Ability Scores:") || strings.Contains(text, "Feat:") ||
				strings.Contains(text, "Skill Proficiencies:") || strings.Contains(text, "Equipment:") {
				parseBackgroundFields(current, text)
			} else {
				if current.Description == "" {
					current.Description = text
				} else {
					current.Description += "\n\n" + text
				}
			}
		}
	}

	if current != nil {
		backgrounds = append(backgrounds, *current)
	}

	return backgrounds, nil
}

// parseSpeciesFields parses combined species metadata fields.
func parseSpeciesFields(s *domain.Species, text string) {
	fields := []string{"Creature Type:", "Size:", "Speed:"}

	type fieldPos struct {
		name  string
		start int
	}
	var positions []fieldPos

	for _, field := range fields {
		idx := strings.Index(text, field)
		if idx >= 0 {
			positions = append(positions, fieldPos{name: field, start: idx})
		}
	}

	// Sort by position
	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			if positions[j].start < positions[i].start {
				positions[i], positions[j] = positions[j], positions[i]
			}
		}
	}

	for i, pos := range positions {
		valueStart := pos.start + len(pos.name)
		valueEnd := len(text)
		if i+1 < len(positions) {
			valueEnd = positions[i+1].start
		}
		value := strings.TrimSpace(text[valueStart:valueEnd])

		switch {
		case strings.Contains(pos.name, "Size"):
			s.Size = value
		case strings.Contains(pos.name, "Speed"):
			for _, c := range value {
				if c >= '0' && c <= '9' {
					s.Speed = s.Speed*10 + int(c-'0')
				}
			}
		}
	}
}

// parseBackgroundFields parses combined background metadata fields.
func parseBackgroundFields(b *domain.Background, text string) {
	fields := []string{"Ability Scores:", "Feat:", "Skill Proficiencies:", "Tool Proficiency:", "Equipment:"}

	type fieldPos struct {
		name  string
		start int
	}
	var positions []fieldPos

	for _, field := range fields {
		idx := strings.Index(text, field)
		if idx >= 0 {
			positions = append(positions, fieldPos{name: field, start: idx})
		}
	}

	// Sort by position
	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			if positions[j].start < positions[i].start {
				positions[i], positions[j] = positions[j], positions[i]
			}
		}
	}

	for i, pos := range positions {
		valueStart := pos.start + len(pos.name)
		valueEnd := len(text)
		if i+1 < len(positions) {
			valueEnd = positions[i+1].start
		}
		value := strings.TrimSpace(text[valueStart:valueEnd])

		switch {
		case strings.Contains(pos.name, "Ability Scores"):
			b.AbilityScoreOptions = map[string]any{"choices": value}
		case strings.Contains(pos.name, "Feat"):
			b.GrantedFeatSlug = shared.Slugify(value)
		case strings.Contains(pos.name, "Skill Proficiencies"):
			b.SkillProficiencies = []string{value}
		case strings.Contains(pos.name, "Equipment"):
			b.Equipment = []string{value}
		}
	}
}
