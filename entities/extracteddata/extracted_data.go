package extracteddata

import (
	"context"
	"time"

	"github.com/solher/hunterio-test/entities/companies"
	"github.com/solher/hunterio-test/entities/people"
)

// ExtractedData represents an extraction run.
type ExtractedData struct {
	ID        uint64              `json:"id" db:"id"`
	URL       string              `json:"url" db:"url"`
	People    []people.Person     `json:"people" db:"people"`
	Companies []companies.Company `json:"companies" db:"companies"`
	CreatedAt time.Time           `json:"created_at" db:"created_at"`
}

// Repository provides access to an ExtractedData store.
type Repository interface {
	Insert(ctx context.Context, extractedData *ExtractedData) (*ExtractedData, error)
	Find(ctx context.Context, search Search) ([]ExtractedData, error)
	GetLastByURL(ctx context.Context, url string) (*ExtractedData, error)
}
