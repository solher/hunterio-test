package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/go-kit/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/peterbourgon/ff"
	"github.com/solher/hunterio-test/entities/extracteddata"
	"github.com/solher/hunterio-test/services/dataextraction"
)

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("hunterio-test", flag.ExitOnError)
	environment := fs.String("environment", "develop", "The deploy environment")
	postgresHost := fs.String("postgres-host", "localhost", "The Postgres database host")
	postgresPort := fs.String("postgres-port", "5432", "The Postgres database port")
	postgresDatabase := fs.String("postgres-database", "hunterio", "The Postgres database name")
	postgresUser := fs.String("postgres-user", "hunterio", "The Postgres user")
	postgresPassword := fs.String("postgres-password", "hunterio", "The Postgres user password")
	openAISecretKey := fs.String("openai-secret-key", "", "The OpenAI secret key")
	ff.Parse(fs, args[1:], ff.WithEnvVarNoPrefix())

	// Infrastructure
	ctx := context.Background()

	// Loggers
	var logger log.Logger
	switch *environment {
	case "stage", "prod":
		logger = log.NewJSONLogger(log.NewSyncWriter(stdout))
	default:
		logger = log.NewLogfmtLogger(log.NewSyncWriter(stdout))
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	// Databases
	config, err := pgxpool.ParseConfig(fmt.Sprintf(
		"user=%s password=%s dbname=%s port=%s sslmode=disable pool_min_conns=2 pool_max_conns=2",
		*postgresUser, *postgresPassword, *postgresDatabase, *postgresPort,
	))
	if err != nil {
		return err
	}
	config.ConnConfig.Host = *postgresHost
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}
	defer db.Close()

	// OpenAI
	openAICli := openai.NewClient(
		option.WithAPIKey(*openAISecretKey),
	)
	if *openAISecretKey == "" {
		return errors.New("openai-secret-key is not set")
	}

	// Repositories
	extractedDataRepo := extracteddata.NewPostgresRepository(db)

	// Services
	dataExtractionService := dataextraction.NewService(logger, &openAICli, extractedDataRepo)

	// We read the URL from the first argument
	if len(fs.Args()) < 1 {
		return errors.New("url is required as first argument")
	}
	url := fs.Args()[0]

	// We extract the data from the URL and print it to stdout.
	extractedData, err := dataExtractionService.ExtractAndPersistFromURL(ctx, url)
	if err != nil {
		return err
	}
	prettyData, err := json.MarshalIndent(extractedData, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%s\n", prettyData)

	return nil
}
