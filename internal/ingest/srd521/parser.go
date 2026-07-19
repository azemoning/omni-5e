package srd521

import (
	"github.com/azemoning/omni-5e/internal/ingest"
)

// Parser implements ContentParser for SRD 5.2.1.
type Parser struct{}

func init() {
	ingest.Register(&Parser{})
}

func (p *Parser) Version() string {
	return "5.2.1"
}
