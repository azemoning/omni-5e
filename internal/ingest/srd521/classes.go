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

// ParseClasses parses classes from classes.md.
func (p *Parser) ParseClasses(ctx context.Context, sourceDir string) ([]domain.Class, error) {
	path := filepath.Join(sourceDir, "classes.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var classes []domain.Class
	var current *domain.Class
	var inDescription bool

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}
			if n.Level == 2 {
				if current != nil {
					classes = append(classes, *current)
				}
			current = &domain.Class{
				BaseEntity: domain.BaseEntity{
					Name:       name,
					Slug:       shared.Slugify(name),
					SRDVersion: p.Version(),
				},
				HitDie: 0,
			}
				inDescription = true
			} else if n.Level == 3 && current != nil {
				lower := strings.ToLower(name)
				if strings.Contains(lower, "hit die") || strings.Contains(lower, "hit points") {
					inDescription = false
				}
			}

		case *ast.Paragraph:
			if current == nil || !inDescription {
				continue
			}
			text := shared.ExtractText(n, src)
			if text == "" {
				continue
			}
			lower := strings.ToLower(text)

			switch {
			case strings.HasPrefix(lower, "hit die:"):
				val := strings.TrimSpace(strings.TrimPrefix(text, "Hit Die:"))
				if len(val) > 1 && val[0] == 'd' {
					current.HitDie = int(val[1] - '0')
				}
			case strings.Contains(lower, "primary ability"):
				current.PrimaryAbility = strings.TrimSpace(strings.TrimPrefix(text, "Primary Ability:"))
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
		classes = append(classes, *current)
	}

	return classes, nil
}

// ParseSubclasses parses subclasses from classes.md.
func (p *Parser) ParseSubclasses(ctx context.Context, sourceDir string) ([]domain.Subclass, error) {
	path := filepath.Join(sourceDir, "classes.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var subclasses []domain.Subclass
	var currentClass string
	var currentSubclass *domain.Subclass

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}
			if n.Level == 2 {
				currentClass = shared.Slugify(name)
			} else if n.Level == 3 && currentClass != "" {
				// Subclass headings typically contain the class name
				if currentSubclass != nil {
					subclasses = append(subclasses, *currentSubclass)
				}
			currentSubclass = &domain.Subclass{
				BaseEntity: domain.BaseEntity{
					Name:       name,
					Slug:       shared.Slugify(name),
					SRDVersion: p.Version(),
				},
				ClassSlug: currentClass,
			}
			}

		case *ast.Paragraph:
			if currentSubclass == nil {
				continue
			}
			text := shared.ExtractText(n, src)
			if currentSubclass.Description == "" {
				currentSubclass.Description = text
			} else {
				currentSubclass.Description += "\n\n" + text
			}
		}
	}

	if currentSubclass != nil {
		subclasses = append(subclasses, *currentSubclass)
	}

	return subclasses, nil
}

// ParseClassFeatures parses class features from classes.md.
func (p *Parser) ParseClassFeatures(ctx context.Context, sourceDir string) ([]domain.ClassFeature, error) {
	path := filepath.Join(sourceDir, "classes.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var features []domain.ClassFeature
	var currentClass string

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}
			if n.Level == 2 {
				currentClass = shared.Slugify(name)
			}
			// Feature names are typically at level 3 or 4
			if (n.Level == 3 || n.Level == 4) && currentClass != "" {
				lower := strings.ToLower(name)
				// Skip non-feature headings
				if strings.Contains(lower, "table") || strings.Contains(lower, "hit die") {
					continue
				}
			features = append(features, domain.ClassFeature{
				BaseEntity: domain.BaseEntity{
					Name:       name,
					Slug:       shared.Slugify(currentClass + "-" + name),
					SRDVersion: p.Version(),
				},
				ClassSlug: currentClass,
			})
			}

		case *ast.Paragraph:
			if len(features) == 0 || currentClass == "" {
				continue
			}
			text := shared.ExtractText(n, src)
			last := &features[len(features)-1]
			if last.Description == "" {
				last.Description = text
			} else {
				last.Description += "\n\n" + text
			}
		}
	}

	return features, nil
}

// ParseClassLevelTables parses level tables from classes.md.
func (p *Parser) ParseClassLevelTables(ctx context.Context, sourceDir string) ([]domain.ClassLevelTableRow, error) {
	path := filepath.Join(sourceDir, "classes.md")
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var rows []domain.ClassLevelTableRow
	var currentClass string

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if n.Level == 2 && name != "" {
				currentClass = shared.Slugify(name)
			}

		case *extast.Table:
			if currentClass == "" {
				continue
			}
			tableRows := shared.ExtractTableRows(n, src)
			if len(tableRows) < 2 {
				continue // need at least header + 1 data row
			}

			// First row is header
			headers := tableRows[0]
			for i := 1; i < len(tableRows); i++ {
				if len(tableRows[i]) < 2 {
					continue
				}
				row := domain.ClassLevelTableRow{
					ClassSlug:      currentClass,
					OtherColumns:   make(map[string]string),
				}
				for j, cell := range tableRows[i] {
					if j == 0 {
						// Level column
						for _, c := range cell {
							if c >= '0' && c <= '9' {
								row.Level = row.Level*10 + int(c-'0')
							}
						}
					} else if j < len(headers) {
						header := strings.ToLower(strings.TrimSpace(headers[j]))
						if strings.Contains(header, "proficiency") {
							for _, c := range cell {
								if c >= '0' && c <= '9' {
									row.ProficiencyBonus = row.ProficiencyBonus*10 + int(c-'0')
								}
							}
						} else {
							row.OtherColumns[header] = strings.TrimSpace(cell)
						}
					}
				}
				if row.Level > 0 {
					rows = append(rows, row)
				}
			}
		}
	}

	return rows, nil
}
