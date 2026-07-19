package domain

import (
	"time"

	"github.com/google/uuid"
)

// BaseEntity contains fields common to all SRD entities.
type BaseEntity struct {
	ID         uuid.UUID `json:"id"`
	Slug       string    `json:"slug"`
	SRDVersion string    `json:"srd_version"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
