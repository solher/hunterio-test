package dataextraction

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-kit/log"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
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
		httpCli:   &http.Client{},
		openAICli: openAICli,
	}
}

type service struct {
	l         log.Logger
	httpCli   *http.Client
	openAICli *openai.Client
}

// ExtractedData represents the data extracted from a URL.
type ExtractedData struct {
	People    []people.Person     `json:"people"`
	Companies []companies.Company `json:"companies"`
}

// ExtractAndPersistFromURL fetches a page from a URL, extracts data from it, and persists it to the database.
func (s *service) ExtractAndPersistFromURL(ctx context.Context, url string) (*ExtractedData, error) {
	strData, err := s.fetchStringDataFromURL(ctx, url)
	if err != nil {
		return nil, err
	}
	data, err := s.extractDataFromString(ctx, strData)
	if err != nil {
		return nil, err
	}
	if err := s.persistExtractedData(ctx, data); err != nil {
		return nil, err
	}
	return data, nil
}

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
	case http.StatusServiceUnavailable:
		return "", fmt.Errorf("service unavailable, try again later")
	case http.StatusNotFound:
		return "", fmt.Errorf("page not found")
	case http.StatusOK:
		// All good, continue
	default:
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

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

// Generate the JSON schema at initialization time
var ExtractedDataSchema = generateSchema[ExtractedData]()

func (s *service) extractDataFromString(ctx context.Context, data string) (*ExtractedData, error) {
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
	var extractedData *ExtractedData
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &extractedData)
	if err != nil {
		return nil, err
	}
	return extractedData, nil
}

func (s *service) persistExtractedData(ctx context.Context, data *ExtractedData) error {
	return nil
}
