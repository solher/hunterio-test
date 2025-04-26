package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-kit/log"
	"github.com/jackc/pgx/v5/pgxpool"
	okrun "github.com/oklog/run"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/peterbourgon/ff"
	"github.com/solher/hunterio-test/entities/extracteddata"
	"github.com/solher/hunterio-test/services/dataextraction"
	"github.com/solher/toolbox/api"
	_ "go.uber.org/automaxprocs"
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
	httpAddr := fs.String("http-addr", ":8080", "HTTP listen address")
	postgresHost := fs.String("postgres-host", "localhost", "The Postgres database host")
	postgresPort := fs.String("postgres-port", "5432", "The Postgres database port")
	postgresDatabase := fs.String("postgres-database", "hunterio", "The Postgres database name")
	postgresUser := fs.String("postgres-user", "hunterio", "The Postgres user")
	postgresPassword := fs.String("postgres-password", "hunterio", "The Postgres user password")
	openAISecretKey := fs.String("openai-secret-key", "", "The OpenAI secret key")
	ff.Parse(fs, args[1:], ff.WithEnvVarNoPrefix())

	// Infrastructure
	ctx := context.Background()
	g := okrun.Group{}

	// Loggers
	var logger log.Logger
	switch *environment {
	case "stage", "prod":
		logger = log.NewJSONLogger(log.NewSyncWriter(stdout))
	default:
		logger = log.NewLogfmtLogger(log.NewSyncWriter(stdout))
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	// Encoders
	jsonRenderer := api.NewJSON(logger, (*environment != "prod"))

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

	// Repositories
	extractedDataRepo := extracteddata.NewPostgresRepository(db)

	// Services
	dataExtractionService := dataextraction.NewService(logger, &openAICli, extractedDataRepo)

	// App router
	httpRouter := chi.NewRouter()
	httpRouter.Mount("/extract", dataextraction.NewHTTPHandler(dataExtractionService, jsonRenderer))

	logger.Log("msg", fmt.Sprintf("listening on %s (HTTP)", *httpAddr))
	httpServer := &http.Server{Addr: *httpAddr, Handler: httpRouter}
	g.Add(func() error { return httpServer.ListenAndServe() }, func(error) { httpServer.Shutdown(ctx) })

	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		s := <-c
		logger.Log("signal", s.String(), "msg", "gracefully shutting down")
		return nil
	}, func(error) {})

	if err := g.Run(); err != nil {
		return err
	}

	return nil
}
