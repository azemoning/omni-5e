package srd521

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/ingest/shared"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

// ParseMonsters parses monsters from monsters-A-Z.md and animals from animals.md.
func (p *Parser) ParseMonsters(ctx context.Context, sourceDir string) ([]domain.Monster, error) {
	var all []domain.Monster

	monsters, err := p.parseMonsterFile(filepath.Join(sourceDir, "monsters-A-Z.md"), "monster")
	if err != nil {
		return nil, err
	}
	all = append(all, monsters...)

	animals, err := p.parseMonsterFile(filepath.Join(sourceDir, "animals.md"), "animal")
	if err != nil {
		return nil, err
	}
	all = append(all, animals...)

	return all, nil
}

func (p *Parser) parseMonsterFile(path, category string) ([]domain.Monster, error) {
	node, src, err := shared.ReadMarkdownFile(path)
	if err != nil {
		return nil, err
	}

	var monsters []domain.Monster
	var current *domain.Monster
	var parseState string // "", "stats", "traits", "actions", "bonus_actions", "reactions", "legendary"
	var currentBlock *domain.NamedBlock

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Heading:
			name := shared.HeadingText(n, src)
			if name == "" {
				continue
			}

			// Skip known section headers that are not monster names
			lowerName := strings.ToLower(name)
			isSectionHeader := lowerName == "traits" || lowerName == "actions" ||
				lowerName == "bonus actions" || lowerName == "reactions" ||
				lowerName == "legendary actions"
			if isSectionHeader {
				if current != nil {
					if strings.Contains(lowerName, "trait") {
						parseState = "traits"
					} else if strings.Contains(lowerName, "bonus") {
						parseState = "bonus_actions"
					} else if strings.Contains(lowerName, "reaction") {
						parseState = "reactions"
					} else if strings.Contains(lowerName, "legendary") {
						parseState = "legendary"
					} else if strings.Contains(lowerName, "action") {
						parseState = "actions"
					}
					currentBlock = nil
				}
				continue
			}

			// Level 2 headings are category headers (A, B, C...)
			// Level 3 headings are monster names
			if n.Level == 3 {
				if current != nil {
					finalizeMonster(current)
					monsters = append(monsters, *current)
				}
				current = &domain.Monster{
					BaseEntity: domain.BaseEntity{
						Name:       name,
						Slug:       shared.Slugify(name),
						SRDVersion: p.Version(),
					},
					Category: category,
					Speed:    make(map[string]int),
				}
				parseState = ""
				currentBlock = nil
			}

		case *ast.Paragraph:
			if current == nil {
				continue
			}
			text := shared.ExtractText(n, src)
			if text == "" {
				continue
			}

			// First paragraph is size/type/alignment: _Large Aberration, Lawful Evil_
			if current.Size == "" && (strings.Contains(text, ",") || isSizeWord(strings.Fields(text)[0])) {
				parseSizeTypeAlignment(current, text)
				continue
			}

			// Parse stat block fields - paragraph must start with AC
			if containsStatField(text) {
				parseMonsterStatFields(current, text)
				continue
			}

			// Parse CR from Skills/Senses/Languages/CR paragraph
			if current.CR == 0 && containsCRField(text) {
				parseCRFromParagraph(current, text)
				continue
			}

			// Parse trait/action blocks
			if parseState != "" {
				// Check for bold block name: **_Name._** or **Name.**
				blockName, blockDesc := parseNamedBlock(text)
				if blockName != "" {
					currentBlock = &domain.NamedBlock{Name: blockName, Description: blockDesc}
					addBlockToMonster(current, parseState, currentBlock)
				} else if currentBlock != nil {
					// Continuation of previous block
					currentBlock.Description += " " + text
				}
			}

		case *extast.Table:
			if current == nil {
				continue
			}
			// Ability score table
			rows := shared.ExtractTableRows(n, src)
			parseAbilityScoreTable(current, rows)
		}
	}

	if current != nil {
		finalizeMonster(current)
		monsters = append(monsters, *current)
	}

	return monsters, nil
}

func isSizeWord(word string) bool {
	sizes := map[string]bool{
		"tiny": true, "small": true, "medium": true,
		"large": true, "huge": true, "gargantuan": true,
	}
	return sizes[strings.ToLower(word)]
}

// containsStatField checks if text contains monster stat fields.
// Must start with AC/Armor Class to be a stat block paragraph.
func containsStatField(text string) bool {
	trimmed := strings.TrimSpace(text)
	return strings.HasPrefix(trimmed, "AC ") || strings.HasPrefix(trimmed, "Armor Class ")
}

// containsCRField checks if text contains CR/Challenge field.
func containsCRField(text string) bool {
	return strings.Contains(text, "CR ") || strings.Contains(text, "Challenge ")
}

// parseMonsterStatFields parses all stat fields from a paragraph.
func parseMonsterStatFields(m *domain.Monster, text string) {
	// Split on field markers
	fields := []string{"AC ", "Armor Class ", "HP ", "Hit Points ", "Speed ", "CR ", "Challenge ", "Skills ", "Senses ", "Languages "}

	// Find positions of all fields
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

	// Extract values between fields
	for i, pos := range positions {
		valueStart := pos.start + len(pos.name)
		valueEnd := len(text)
		if i+1 < len(positions) {
			valueEnd = positions[i+1].start
		}
		value := strings.TrimSpace(text[valueStart:valueEnd])
		// Remove <br> tags
		value = strings.ReplaceAll(value, "<br>", "")
		value = strings.TrimSpace(value)

		switch {
		case strings.HasPrefix(pos.name, "AC ") || strings.HasPrefix(pos.name, "Armor Class "):
			parseACValue(m, value)
		case strings.HasPrefix(pos.name, "HP ") || strings.HasPrefix(pos.name, "Hit Points "):
			parseHPValue(m, value)
		case strings.HasPrefix(pos.name, "Speed "):
			parseSpeedValue(m, value)
		case strings.HasPrefix(pos.name, "CR ") || strings.HasPrefix(pos.name, "Challenge "):
			parseCRValue(m, value)
		}
	}
}

func parseACValue(m *domain.Monster, value string) {
	val := 0
	for _, c := range value {
		if c >= '0' && c <= '9' {
			val = val*10 + int(c-'0')
		} else if val > 0 {
			break
		}
	}
	m.AC = domain.ACInfo{Value: val}
}

func parseHPValue(m *domain.Monster, value string) {
	// Stop at next field marker (Speed, Skills, etc.)
	for _, marker := range []string{"Speed ", "Skills ", "Senses ", "Languages ", "CR ", "Challenge "} {
		if idx := strings.Index(value, marker); idx >= 0 {
			value = value[:idx]
		}
	}
	value = strings.TrimSpace(value)

	parts := strings.SplitN(value, "(", 2)
	avgStr := strings.TrimSpace(parts[0])
	avg := 0
	for _, c := range avgStr {
		if c >= '0' && c <= '9' {
			avg = avg*10 + int(c-'0')
		}
	}
	m.HP = domain.HPInfo{Average: avg}
	if len(parts) > 1 {
		formula := strings.TrimSuffix(strings.TrimSpace(parts[1]), ")")
		m.HP.Formula = formula
	}
}

func parseSpeedValue(m *domain.Monster, value string) {
	// Stop at next field marker
	for _, marker := range []string{"Skills ", "Senses ", "Languages ", "CR ", "Challenge "} {
		if idx := strings.Index(value, marker); idx >= 0 {
			value = value[:idx]
		}
	}
	value = strings.TrimSpace(value)

	parts := strings.Split(value, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		words := strings.Fields(part)
		if len(words) >= 2 {
			// Format: "10 ft." or "Swim 40 ft."
			dist := 0
			name := "walk"
			for _, c := range words[0] {
				if c >= '0' && c <= '9' {
					dist = dist*10 + int(c-'0')
				}
			}
			if dist > 0 {
				// First word is a number, so it's "N ft." format
				m.Speed[name] = dist
			} else {
				// First word is a name like "Swim", "Fly", etc.
				name = strings.ToLower(words[0])
				for _, c := range words[1] {
					if c >= '0' && c <= '9' {
						dist = dist*10 + int(c-'0')
					}
				}
				m.Speed[name] = dist
			}
		}
	}
}

func parseCRValue(m *domain.Monster, value string) {
	// Stop at next field marker or end
	for _, marker := range []string{"Skills ", "Senses ", "Languages "} {
		if idx := strings.Index(value, marker); idx >= 0 {
			value = value[:idx]
		}
	}
	value = strings.TrimSpace(value)

	parts := strings.Fields(value)
	if len(parts) > 0 {
		crStr := parts[0]
		switch crStr {
		case "0":
			m.CR = 0
		case "1/8":
			m.CR = 0.125
		case "1/4":
			m.CR = 0.25
		case "1/2":
			m.CR = 0.5
		default:
			cr, _ := strconv.ParseFloat(crStr, 64)
			m.CR = cr
		}
	}

	// Extract XP - format: "(XP 5,900, or 7,200 in lair; PB +4)"
	if xpIdx := strings.Index(value, "XP"); xpIdx >= 0 {
		xpText := value[xpIdx+2:]
		xpText = strings.TrimSpace(xpText)
		// Extract number, handling commas
		xpStr := ""
		for _, c := range xpText {
			if c >= '0' && c <= '9' {
				xpStr += string(c)
			} else if c == ',' {
				continue // skip commas in numbers
			} else if xpStr != "" {
				break
			}
		}
		if xpStr != "" {
			xp, _ := strconv.Atoi(xpStr)
			m.XP = xp
		}
	}
}

// parseCRFromParagraph parses CR and XP from the Skills/Senses/Languages/CR paragraph.
func parseCRFromParagraph(m *domain.Monster, text string) {
	// Find CR in the text
	crIdx := strings.Index(text, "CR ")
	if crIdx < 0 {
		crIdx = strings.Index(text, "Challenge ")
		if crIdx < 0 {
			return
		}
	}

	// Extract everything after CR
	crText := text[crIdx:]
	// Remove "CR " or "Challenge " prefix
	crText = strings.TrimPrefix(crText, "CR ")
	crText = strings.TrimPrefix(crText, "Challenge ")
	crText = strings.TrimSpace(crText)

	// Parse CR value
	parts := strings.Fields(crText)
	if len(parts) > 0 {
		crStr := parts[0]
		switch crStr {
		case "0":
			m.CR = 0
		case "1/8":
			m.CR = 0.125
		case "1/4":
			m.CR = 0.25
		case "1/2":
			m.CR = 0.5
		default:
			cr, _ := strconv.ParseFloat(crStr, 64)
			m.CR = cr
		}
	}

	// Extract XP
	if xpIdx := strings.Index(crText, "XP"); xpIdx >= 0 {
		xpText := crText[xpIdx+2:]
		xpText = strings.TrimSpace(xpText)
		xpStr := ""
		for _, c := range xpText {
			if c >= '0' && c <= '9' {
				xpStr += string(c)
			} else if c == ',' {
				continue
			} else if xpStr != "" {
				break
			}
		}
		if xpStr != "" {
			xp, _ := strconv.Atoi(xpStr)
			m.XP = xp
		}
	}
}

func parseSizeTypeAlignment(m *domain.Monster, text string) {
	// Format: "_Large Aberration, Lawful Evil_"
	text = strings.Trim(text, "_*")
	parts := strings.SplitN(text, ",", 2)
	if len(parts) == 0 {
		return
	}
	firstPart := strings.TrimSpace(parts[0])
	words := strings.Fields(firstPart)
	if len(words) >= 2 {
		m.Size = words[0]
		m.Type = words[1]
	}
	if len(parts) > 1 {
		m.Alignment = strings.TrimSpace(parts[1])
	}
}

func parseACField(m *domain.Monster, text string) {
	// Format: "AC 17 Initiative +7 (17)" or "**AC** 17"
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "AC ")
	text = strings.TrimPrefix(text, "Armor Class ")
	text = strings.ReplaceAll(text, "**AC**", "")
	text = strings.ReplaceAll(text, "**Armor Class**", "")
	text = strings.TrimSpace(text)
	// Extract first number
	val := 0
	for _, c := range text {
		if c >= '0' && c <= '9' {
			val = val*10 + int(c-'0')
		} else if val > 0 {
			break
		}
	}
	m.AC = domain.ACInfo{Value: val}
}

func parseHPField(m *domain.Monster, text string) {
	// Format: "HP 150 (20d10 + 40)" or "**HP** 150 (20d10 + 40)"
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "HP ")
	text = strings.TrimPrefix(text, "Hit Points ")
	text = strings.ReplaceAll(text, "**HP**", "")
	text = strings.ReplaceAll(text, "**Hit Points**", "")
	text = strings.TrimSpace(text)

	// Extract average
	parts := strings.SplitN(text, "(", 2)
	avgStr := strings.TrimSpace(parts[0])
	avg := 0
	for _, c := range avgStr {
		if c >= '0' && c <= '9' {
			avg = avg*10 + int(c-'0')
		}
	}
	m.HP = domain.HPInfo{Average: avg}

	if len(parts) > 1 {
		formula := strings.TrimSuffix(strings.TrimSpace(parts[1]), ")")
		m.HP.Formula = formula
	}
}

func parseSpeedField(m *domain.Monster, text string) {
	// Format: "Speed 10 ft., Swim 40 ft." or "**Speed** 10 ft., Swim 40 ft."
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "Speed ")
	text = strings.ReplaceAll(text, "**Speed**", "")
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "<br>", "")

	parts := strings.Split(text, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		words := strings.Fields(part)
		if len(words) >= 2 {
			name := strings.ToLower(words[0])
			dist := 0
			for _, c := range words[1] {
				if c >= '0' && c <= '9' {
					dist = dist*10 + int(c-'0')
				}
			}
			m.Speed[name] = dist
		} else if len(words) == 1 {
			dist := 0
			for _, c := range words[0] {
				if c >= '0' && c <= '9' {
					dist = dist*10 + int(c-'0')
				}
			}
			if dist > 0 {
				m.Speed["walk"] = dist
			}
		}
	}
}

func parseCRField(m *domain.Monster, text string) {
	// Format: "CR 10 (XP 5,900)" or "**CR** 10 (XP 5,900)"
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "CR ")
	text = strings.TrimPrefix(text, "Challenge ")
	text = strings.ReplaceAll(text, "**CR**", "")
	text = strings.ReplaceAll(text, "**Challenge**", "")
	text = strings.TrimSpace(text)

	// Extract CR value
	parts := strings.Fields(text)
	if len(parts) > 0 {
		crStr := parts[0]
		switch crStr {
		case "0":
			m.CR = 0
		case "1/8":
			m.CR = 0.125
		case "1/4":
			m.CR = 0.25
		case "1/2":
			m.CR = 0.5
		default:
			cr, _ := strconv.ParseFloat(crStr, 64)
			m.CR = cr
		}
	}

	// Extract XP
	if xpIdx := strings.Index(text, "XP"); xpIdx >= 0 {
		xpText := text[xpIdx+2:]
		xpText = strings.TrimSpace(xpText)
		// Find first number
		xpStr := ""
		for _, c := range xpText {
			if c >= '0' && c <= '9' {
				xpStr += string(c)
			} else if xpStr != "" {
				break
			}
		}
		if xpStr != "" {
			xp, _ := strconv.Atoi(xpStr)
			m.XP = xp
		}
	}
}

func parseAbilityScoreTable(m *domain.Monster, rows [][]string) {
	// The table has format: STR 21 +5 +5 DEX 9 -1 +3 CON 15 +2 +6
	//                       INT 18 +4 +8 WIS 15 +2 +6 CHA 18 +4 +4
	m.AbilityScores = domain.AbilityScores{}
	for _, row := range rows {
		for i := 0; i < len(row)-1; i += 2 {
			ability := strings.TrimSpace(row[i])
			valStr := strings.TrimSpace(row[i+1])
			val, err := strconv.Atoi(valStr)
			if err != nil {
				continue
			}
			switch strings.ToUpper(ability) {
			case "STR":
				m.AbilityScores.STR = val
			case "DEX":
				m.AbilityScores.DEX = val
			case "CON":
				m.AbilityScores.CON = val
			case "INT":
				m.AbilityScores.INT = val
			case "WIS":
				m.AbilityScores.WIS = val
			case "CHA":
				m.AbilityScores.CHA = val
			}
		}
	}
}

func parseNamedBlock(text string) (name, desc string) {
	// Format after stripping: "Amphibious. The aboleth can breathe air and water."
	// Or: "Name. Description"
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ""
	}

	// Find the first period followed by a space (or end of short name)
	// This separates "Amphibious." from "The aboleth can..."
	idx := strings.Index(text, ". ")
	if idx > 0 && idx < 60 {
		name = text[:idx]
		desc = strings.TrimSpace(text[idx+2:])
		return name, desc
	}

	// Also try with colon
	idx = strings.Index(text, ": ")
	if idx > 0 && idx < 60 {
		name = text[:idx]
		desc = strings.TrimSpace(text[idx+2:])
		return name, desc
	}

	return "", ""
}

func addBlockToMonster(m *domain.Monster, state string, block *domain.NamedBlock) {
	switch state {
	case "traits":
		m.Traits = append(m.Traits, *block)
	case "actions":
		m.Actions = append(m.Actions, *block)
	case "bonus_actions":
		m.BonusActions = append(m.BonusActions, *block)
	case "reactions":
		m.Reactions = append(m.Reactions, *block)
	case "legendary":
		m.LegendaryActions = append(m.LegendaryActions, *block)
	}
}

func finalizeMonster(m *domain.Monster) {
	// Set defaults if parsing missed something
	if m.Speed == nil {
		m.Speed = map[string]int{"walk": 30}
	}
}
