package srd521

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/ingest/shared"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

// ParseFeats parses feats from feats.md.
func (p *Parser) ParseFeats(ctx context.Context, sourceDir string) ([]domain.Feat, error) {
	path := filepath.Join(sourceDir, "feats.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var feats []domain.Feat
	var current *domain.Feat
	inFeatSection := false

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}
			lower := strings.ToLower(name)
			if strings.Contains(lower, "feat descriptions") || strings.Contains(lower, "origin feats") ||
				strings.Contains(lower, "general feats") || strings.Contains(lower, "fighting style") ||
				strings.Contains(lower, "epic boon") {
				inFeatSection = true
				continue
			}
			// Feat entries are at level 4
			if n.Level == 4 && inFeatSection {
				if current != nil {
					feats = append(feats, *current)
				}
				current = &domain.Feat{
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
			lower := strings.ToLower(text)
			trimmed := strings.Trim(text, "_*")

			// Category line: "_Origin Feat_" or "_General Feat_"
			if strings.Contains(lower, "feat") && len(trimmed) < 30 {
				current.Category = trimmed
				continue
			}
			if strings.Contains(lower, "prerequisite:") {
				parts := strings.SplitN(text, ":", 2)
				if len(parts) > 1 {
					current.Prerequisite = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(lower, "repeatable") {
				current.Repeatable = true
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
		feats = append(feats, *current)
	}

	return feats, nil
}

// ParseEquipment parses equipment from equipment.md.
func (p *Parser) ParseEquipment(ctx context.Context, sourceDir string) ([]domain.Equipment, error) {
	path := filepath.Join(sourceDir, "equipment.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var items []domain.Equipment
	var current *domain.Equipment
	currentCategory := "gear"

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}
			lower := strings.ToLower(name)
			if n.Level == 2 {
				switch {
				case strings.Contains(lower, "weapon"):
					currentCategory = "weapon"
				case strings.Contains(lower, "armor"):
					currentCategory = "armor"
				case strings.Contains(lower, "tool"):
					currentCategory = "tool"
				case strings.Contains(lower, "mount") || strings.Contains(lower, "vehicle"):
					currentCategory = "mount"
				default:
					currentCategory = "gear"
				}
			}
			if n.Level >= 2 {
				if current != nil {
					items = append(items, *current)
				}
			current = &domain.Equipment{
				BaseEntity: domain.BaseEntity{
					Name:       name,
					Slug:       shared.Slugify(name),
					SRDVersion: p.Version(),
				},
				Category:   currentCategory,
				Properties: make(map[string]any),
			}
			}

		case *extast.Table:
			if current == nil {
				continue
			}
			rows := shared.ExtractTableRows(n, src)
			for _, row := range rows {
				if len(row) >= 2 {
					current.Properties[strings.ToLower(strings.TrimSpace(row[0]))] = strings.TrimSpace(row[1])
				}
			}

		case *ast.Paragraph:
			if current == nil {
				continue
			}
			text := shared.ExtractText(n, src)
			if current.Description == "" {
				current.Description = text
			} else {
				current.Description += "\n\n" + text
			}
		}
	}

	if current != nil {
		items = append(items, *current)
	}

	return items, nil
}

// ParseMagicItems parses magic items from magic-items.md.
func (p *Parser) ParseMagicItems(ctx context.Context, sourceDir string) ([]domain.MagicItem, error) {
	path := filepath.Join(sourceDir, "magic-items.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var items []domain.MagicItem
	var current *domain.MagicItem

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}
			if n.Level >= 2 {
				if current != nil {
					items = append(items, *current)
				}
			current = &domain.MagicItem{
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
			lower := strings.ToLower(text)

			switch {
			case strings.Contains(lower, "rarity:"):
				parts := strings.SplitN(text, ":", 2)
				if len(parts) > 1 {
					current.Rarity = strings.TrimSpace(parts[1])
				}
			case strings.Contains(lower, "requires attunement"):
				current.RequiresAttunement = true
			case strings.Contains(lower, "type:"):
				parts := strings.SplitN(text, ":", 2)
				if len(parts) > 1 {
					current.Type = strings.TrimSpace(parts[1])
				}
			default:
				if current.Description == "" {
					current.Description = text
				} else {
					current.Description += "\n\n" + text
				}
			}
		}
	}

	if current != nil {
		items = append(items, *current)
	}

	return items, nil
}

// ParseRuleSections parses rule sections from various source files.
func (p *Parser) ParseRuleSections(ctx context.Context, sourceDir string) ([]domain.RuleSection, error) {
	files := []string{
		"character-creation.md",
		"playing-the-game.md",
		"gameplay-toolbox.md",
		"monsters.md",
	}

	var sections []domain.RuleSection

	for _, file := range files {
		path := filepath.Join(sourceDir, file)
		node, src, err := shared.ReadMarkdownFile(path)
		if err != nil {
			continue // skip missing files
		}

		var currentHeading string
		var currentBody strings.Builder
		var headingPath []string

		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			switch n := child.(type) {
			case *ast.Heading:
				name := shared.HeadingText(n, src)
				if name == "" {
					continue
				}
				// Save previous section
				if currentHeading != "" && currentBody.Len() > 0 {
				sections = append(sections, domain.RuleSection{
					BaseEntity: domain.BaseEntity{
						Name:       currentHeading,
						Slug:       shared.Slugify(currentHeading),
						SRDVersion: p.Version(),
					},
					SourceFile:  file,
					HeadingPath: headingPath,
					Body:        currentBody.String(),
				})
				}
				currentHeading = name
				currentBody.Reset()
				headingPath = append(headingPath[:n.Level-1], name)

			case *ast.Paragraph:
				if currentHeading == "" {
					continue
				}
				text := shared.ExtractText(n, src)
				if currentBody.Len() > 0 {
					currentBody.WriteString("\n\n")
				}
				currentBody.WriteString(text)
			}
		}

		// Save last section
		if currentHeading != "" && currentBody.Len() > 0 {
			sections = append(sections, domain.RuleSection{
				BaseEntity: domain.BaseEntity{
					Name:       currentHeading,
					Slug:       shared.Slugify(currentHeading),
					SRDVersion: p.Version(),
				},
				SourceFile:  file,
				HeadingPath: headingPath,
				Body:        currentBody.String(),
			})
		}
	}

	return sections, nil
}
