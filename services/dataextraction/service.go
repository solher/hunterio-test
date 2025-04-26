package dataextraction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-kit/log"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/solher/hunterio-test/entities/companies"
	"github.com/solher/hunterio-test/entities/extracteddata"
	"github.com/solher/hunterio-test/entities/people"
)

// Service represents the data extraction service interface.
type Service interface {
	ExtractAndPersistFromURL(ctx context.Context, url string) (*extracteddata.ExtractedData, error)
	GetExtractedDataHistory(ctx context.Context, url string, from time.Time, to time.Time, limit int, offset int) ([]extracteddata.ExtractedData, error)
}

// NewService returns a new instance of the data extraction service.
func NewService(
	l log.Logger,
	openAICli *openai.Client,
	extractedDataRepo extracteddata.Repository,
) Service {
	return &service{
		l:                 l,
		httpCli:           &http.Client{},
		openAICli:         openAICli,
		extractedDataRepo: extractedDataRepo,
	}
}

type service struct {
	l                 log.Logger
	httpCli           *http.Client
	openAICli         *openai.Client
	extractedDataRepo extracteddata.Repository
}

const (
	cacheFreshness = 1 * time.Hour
)

var (
	ErrServiceUnavailable = errors.New("service unavailable, try again later")
	ErrPageNotFound       = errors.New("page not found")
)

// ExtractAndPersistFromURL fetches a page from a URL, extracts data from it, and persists it to the database.
func (s *service) ExtractAndPersistFromURL(ctx context.Context, url string) (*extracteddata.ExtractedData, error) {
	// First, we check if the data is already in the database for this URL.
	extractedData, err := s.extractedDataRepo.GetLastByURL(ctx, url)
	if err != nil && err != extracteddata.ErrNotFound {
		return nil, err
	}

	// If the data is fresh, we return it. Otherwise, we refetch.
	if extractedData != nil && extractedData.CreatedAt.After(time.Now().Add(-cacheFreshness)) {
		return extractedData, nil
	}

	// If the data is not in the database, we fetch it from the URL and extract the data using the OpenAI API.
	strData, err := s.fetchStringDataFromURL(ctx, url)
	if err != nil {
		return nil, err
	}
	openAIData, err := s.extractDataFromString(ctx, strData)
	if err != nil {
		return nil, err
	}

	// Then, we persist it to the database.
	extractedData, err = s.persistExtractedData(ctx, url, openAIData)
	if err != nil {
		return nil, err
	}
	return extractedData, nil
}

// fetchStringDataFromURL fetches a page from a URL and returns the content as a string.
func (s *service) fetchStringDataFromURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := s.httpCli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// All good, continue
	case http.StatusNotFound:
		return "", ErrPageNotFound
	default:
		return "", ErrServiceUnavailable
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// generateSchema generates a JSON schema for a given struct.
func generateSchema[T any]() interface{} {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

type openAIExtractedData struct {
	Companies []companies.Company `json:"companies"`
	People    []people.Person     `json:"people"`
}

// Generate the JSON schema at initialization time
var ExtractedDataSchema = generateSchema[openAIExtractedData]()

// extractDataFromString extracts data from a string using the OpenAI API.
func (s *service) extractDataFromString(ctx context.Context, data string) (*openAIExtractedData, error) {
	// Define the prompt
	prompt := `
You're looking for B2B data to help with lead generation for a CRM tool. Extract companies and people from the following webpage content.
Be extra careful when extracting data and prefer to discard info if you have any doubt that it's matching the expected format.

Webpage:
` + data

	// Define the response format
	resFormat := openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
			JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        "extracted_companies_people",
				Description: openai.String("Extracted companies and people from a webpage"),
				Schema:      ExtractedDataSchema,
				Strict:      openai.Bool(true),
			},
		},
	}

	chat, err := s.openAICli.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		ResponseFormat: resFormat,
		Model:          openai.ChatModelGPT4o2024_08_06,
		Temperature:    param.NewOpt(0.0), // We want the output to be the most deterministic possible.
	})
	if err != nil {
		return nil, err
	}
	if len(chat.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	// extract into a well-typed struct
	extractedData := &openAIExtractedData{}
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), extractedData)
	if err != nil {
		return nil, err
	}
	return extractedData, nil
}

// persistExtractedData persists the extracted data to the database.
func (s *service) persistExtractedData(ctx context.Context, url string, data *openAIExtractedData) (*extracteddata.ExtractedData, error) {
	newData, err := s.extractedDataRepo.Insert(ctx, &extracteddata.ExtractedData{
		URL:       url,
		Companies: data.Companies,
		People:    data.People,
	})
	if err != nil {
		return nil, err
	}
	return newData, nil
}

// GetExtractedDataHistory returns the extracted data history for a given URL.
func (s *service) GetExtractedDataHistory(ctx context.Context, url string, from time.Time, to time.Time, limit int, offset int) ([]extracteddata.ExtractedData, error) {
	if from.IsZero() {
		return nil, errors.New("from cannot be zero")
	}
	if to.IsZero() {
		return nil, errors.New("to cannot be zero")
	}
	if limit == 0 || limit > 10 {
		limit = 10
	}

	extractedDataList, err := s.extractedDataRepo.Find(ctx, extracteddata.Search{
		URL:           url,
		CreatedAtFrom: from,
		CreatedAtTo:   to,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	return extractedDataList, nil
}
