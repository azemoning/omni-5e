package srd521

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/ingest/shared"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

var (
	levelSchoolRe = regexp.MustCompile(`(?i)(Level\s+(\d+)\s+)?(\w+)\s+(Cantrip|(\w+))`)
	cantripRe     = regexp.MustCompile(`(?i)(\w+)\s+Cantrip`)
)

// ParseSpells parses spells from spells.md.
func (p *Parser) ParseSpells(ctx context.Context, sourceDir string) ([]domain.Spell, error) {
	path := filepath.Join(sourceDir, "spells.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var spells []domain.Spell
	var current *domain.Spell
	inSpellSection := false

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}
			// "Spell Descriptions" marks the start of actual spell entries
			if strings.Contains(strings.ToLower(name), "spell descriptions") {
				inSpellSection = true
				continue
			}
			// Skip section headers within spells (for summoned creatures)
			lowerName := strings.ToLower(name)
			if lowerName == "traits" || lowerName == "actions" || lowerName == "bonus actions" ||
				lowerName == "reactions" || lowerName == "legendary actions" {
				continue
			}
			// Spell entries are at heading level 4
			if n.Level == 4 && inSpellSection {
				if current != nil {
					spells = append(spells, *current)
				}
				current = &domain.Spell{
					BaseEntity: domain.BaseEntity{
						Name:       name,
						Slug:       shared.Slugify(name),
						SRDVersion: p.Version(),
					},
				}
			}

		case *ast.Paragraph:
			if current == nil {
				continue
			}
			text := shared.ExtractText(n, src)
			if text == "" {
				continue
			}

			// First paragraph after name is italic metadata: _Level 2 Evocation (Wizard)_
			if current.School == "" && (strings.Contains(text, "Cantrip") || strings.Contains(text, "Level")) {
				parseSpellLevelSchool(current, text)
				continue
			}

			// Parse metadata fields - they may be combined in one paragraph
			// Split on bold markers to handle "**Casting Time:** Action **Range:** 150 feet"
			if containsBoldField(text) {
				parseSpellMetadata(current, text)
				continue
			}

			// Description text
			if current.Description == "" {
				current.Description = text
			} else {
				current.Description += "\n\n" + text
			}

		case *extast.Table:
			if current == nil {
				continue
			}
			rows := shared.ExtractTableRows(n, src)
			for _, row := range rows {
				current.Description += "\n| " + strings.Join(row, " | ") + " |"
			}
		}
	}

	if current != nil {
		spells = append(spells, *current)
	}

	return spells, nil
}

// containsBoldField checks if text contains any metadata fields.
func containsBoldField(text string) bool {
	return strings.Contains(text, "Casting Time:") ||
		strings.Contains(text, "Range:") ||
		strings.Contains(text, "Components:") ||
		strings.Contains(text, "Component:") ||
		strings.Contains(text, "Duration:")
}

// parseSpellMetadata parses all metadata fields from a paragraph.
func parseSpellMetadata(spell *domain.Spell, text string) {
	// Split on field markers (bold markers may be stripped)
	// Check both "Components:" and "Component:" since goldmark may merge lines
	// e.g. "90 feetComponents: V, S" or "90 feetComponent: V, S"
	fields := []string{"Casting Time:", "Range:", "Components:", "Component:", "Duration:"}

	// Find positions of all fields, deduplicating overlapping matches
	type fieldPos struct {
		name  string
		start int
	}
	var positions []fieldPos

	for _, field := range fields {
		idx := strings.Index(text, field)
		if idx >= 0 {
			// Skip if another field already claimed this position
			dupe := false
			for _, p := range positions {
				if p.start == idx {
					dupe = true
					break
				}
			}
			if !dupe {
				positions = append(positions, fieldPos{name: field, start: idx})
			}
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

	// Extract values between fields
	for i, pos := range positions {
		valueStart := pos.start + len(pos.name)
		valueEnd := len(text)
		if i+1 < len(positions) {
			valueEnd = positions[i+1].start
		}
		value := strings.TrimSpace(text[valueStart:valueEnd])

		switch {
		case strings.Contains(pos.name, "Casting Time"):
			spell.CastingTime = value
		case strings.Contains(pos.name, "Range"):
			spell.Range = value
		case strings.Contains(pos.name, "Component"):
			parseComponents(spell, value)
		case strings.Contains(pos.name, "Duration"):
			spell.Duration = value
			if strings.Contains(strings.ToLower(value), "concentration") {
				spell.Concentration = true
			}
			if strings.Contains(strings.ToLower(value), "ritual") || strings.Contains(strings.ToLower(spell.CastingTime), "ritual") {
				spell.Ritual = true
			}
		}
	}
}

// parseSpellLevelSchool extracts level and school from the italic metadata line.
// Format: "_Level 2 Evocation (Wizard)_" or "_Evocation Cantrip (Sorcerer, Wizard)_"
func parseSpellLevelSchool(spell *domain.Spell, text string) {
	// Remove italic markers
	text = strings.Trim(text, "_*")
	lower := strings.ToLower(text)

	// Check for cantrip
	if strings.Contains(lower, "cantrip") {
		spell.Level = 0
		// Extract school before "Cantrip"
		parts := strings.SplitN(text, "Cantrip", 2)
		if len(parts) > 0 {
			school := strings.TrimSpace(parts[0])
			spell.School = strings.ToLower(school)
		}
		return
	}

	// Check for "Level N School"
	if strings.HasPrefix(lower, "level") {
		parts := strings.Fields(text)
		if len(parts) >= 3 {
			// parts[0] = "Level", parts[1] = number, parts[2] = school
			level := 0
			for _, c := range parts[1] {
				if c >= '0' && c <= '9' {
					level = level*10 + int(c-'0')
				}
			}
			spell.Level = level
			spell.School = strings.ToLower(parts[2])
		}
	}
}

// parseComponents extracts V/S/M from component string.
func parseComponents(spell *domain.Spell, val string) {
	val = strings.TrimSpace(val)
	parts := strings.Split(val, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		partLower := strings.ToLower(part)
		switch {
		case strings.HasPrefix(partLower, "v"):
			spell.Components.Verbal = true
		case strings.HasPrefix(partLower, "s"):
			spell.Components.Somatic = true
		case strings.HasPrefix(partLower, "m"):
			spell.Components.Material = true
			// Extract material detail in parentheses
			if idx := strings.Index(part, "("); idx >= 0 {
				end := strings.LastIndex(part, ")")
				if end > idx {
					spell.Components.MaterialDetail = strings.TrimSpace(part[idx+1 : end])
				}
			}
		}
	}
}
