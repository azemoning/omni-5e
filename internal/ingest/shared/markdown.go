package shared

import (
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// ReadMarkdownFile reads a markdown file and returns the AST and source bytes.
func ReadMarkdownFile(path string) (ast.Node, []byte, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	parser := goldmark.New().Parser()
	reader := text.NewReader(src)
	node := parser.Parse(reader)

	return node, src, nil
}

// ExtractText extracts plain text from an AST node.
func ExtractText(n ast.Node, src []byte) string {
	var sb strings.Builder
	ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if t, ok := n.(*ast.Text); ok {
				sb.Write(t.Segment.Value(src))
			}
		}
		return ast.WalkContinue, nil
	})
	return sb.String()
}

// HeadingText returns the text content of a heading node.
func HeadingText(n *ast.Heading, src []byte) string {
	return ExtractText(n, src)
}

// FindHeadings finds all headings at a given level under a parent node.
func FindHeadings(parent ast.Node, src []byte, level int) []*ast.Heading {
	var headings []*ast.Heading
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		if h, ok := child.(*ast.Heading); ok && h.Level == level {
			headings = append(headings, h)
		}
	}
	return headings
}

// CollectSiblingContent collects all nodes between two siblings (exclusive).
func CollectSiblingContent(start, end ast.Node, src []byte) string {
	var sb strings.Builder
	for n := start; n != nil && n != end; n = n.NextSibling() {
		sb.WriteString(ExtractText(n, src))
		sb.WriteString("\n")
	}
	return strings.TrimSpace(sb.String())
}

// ExtractTableRows extracts rows from a table node as [][]string.
func ExtractTableRows(n *extast.Table, src []byte) [][]string {
	var rows [][]string
	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		tr, ok := row.(*extast.TableRow)
		if !ok {
			continue
		}
		var cells []string
		for cell := tr.FirstChild(); cell != nil; cell = cell.NextSibling() {
			cells = append(cells, ExtractText(cell, src))
		}
		rows = append(rows, cells)
	}
	return rows
}
