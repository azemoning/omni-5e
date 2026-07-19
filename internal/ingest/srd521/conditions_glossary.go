package srd521

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/ingest/shared"
	"github.com/yuin/goldmark/ast"
)

// ParseConditions parses conditions from rules-glossary.md.
// Conditions are heading level 4 entries with [Condition] tag.
func (p *Parser) ParseConditions(ctx context.Context, sourceDir string) ([]domain.Condition, error) {
	path := filepath.Join(sourceDir, "rules-glossary.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var conditions []domain.Condition
	var current *domain.Condition

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}
			// Conditions are at level 4 with [Condition] tag
			if n.Level == 4 && strings.Contains(strings.ToLower(name), "[condition]") {
				if current != nil {
					conditions = append(conditions, *current)
				}
				// Remove [Condition] tag from name
				cleanName := strings.ReplaceAll(name, "[Condition]", "")
				cleanName = strings.ReplaceAll(cleanName, "[condition]", "")
				cleanName = strings.TrimSpace(cleanName)

				current = &domain.Condition{
					BaseEntity: domain.BaseEntity{
						Name:       cleanName,
						Slug:       shared.Slugify(cleanName),
						SRDVersion: p.Version(),
					},
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
		conditions = append(conditions, *current)
	}

	return conditions, nil
}

// ParseGlossaryTerms parses glossary terms from rules-glossary.md.
// Terms are heading level 4 entries (excluding those with [Condition] tag).
func (p *Parser) ParseGlossaryTerms(ctx context.Context, sourceDir string) ([]domain.GlossaryTerm, error) {
	path := filepath.Join(sourceDir, "rules-glossary.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var terms []domain.GlossaryTerm
	var current *domain.GlossaryTerm

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}
			// Glossary terms are at level 4, skip conditions
			if n.Level == 4 && !strings.Contains(strings.ToLower(name), "[condition]") {
				if current != nil {
					terms = append(terms, *current)
				}
				current = &domain.GlossaryTerm{
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
			if current.Definition == "" {
				current.Definition = text
			} else {
				current.Definition += "\n\n" + text
			}
		}
	}

	if current != nil {
		terms = append(terms, *current)
	}

	return terms, nil
}
