package dataextraction

import (
	"context"

	"github.com/go-kit/log"
	"github.com/openai/openai-go"
	"github.com/solher/hunterio-test/entities/companies"
	"github.com/solher/hunterio-test/entities/people"
)

// Service represents the data extraction service interface.
type Service interface {
	ExtractAndPersistFromURL(ctx context.Context, url string) (*ExtractedData, error)
}

// NewService returns a new instance of the data extraction service.
func NewService(
	l log.Logger,
	openAICli *openai.Client,
) Service {
	return &service{
		l:         l,
		openAICli: openAICli,
	}
}

type service struct {
	l         log.Logger
	openAICli *openai.Client
}

// ExtractedData represents the data extracted from a URL.
type ExtractedData struct {
	People    []people.Person
	Companies []companies.Company
}

func (s *service) ExtractAndPersistFromURL(ctx context.Context, url string) (*ExtractedData, error) {
	return nil, nil
}

func (s *service) extractDataFromURL(ctx context.Context, url string) (*ExtractedData, error) {
	return nil, nil
}

func (s *service) persistData(ctx context.Context, data *ExtractedData) error {
	return nil
}
