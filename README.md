# hunterio-test

## Installation

First, create a `develop.env` file at the root of the project with the following variables, that will be used both for the database setup and the API:

```bash
export POSTGRES_PORT=6432
export OPENAI_SECRET_KEY=<OPENAI_SECRET>
```

Then, run the following command to spin up a local Postgres database and migrate the schema:

```bash
make db
```


That's it. No need to install the dependencies, everything is vendored.

## Running the API

```bash
make run-api
```

## Running the CLI

First, install the CLI binary:

```bash
make install-cli
```

Then, run the CLI with the following command:

```bash
hunterio-test-cli --postgres-port=6432 --openai-secret-key=<OPENAI_SECRET> https://hunter.io/about
```

## Decisions

### Database

I chose to use a Postgres database for this project. Mainly because it's simple and boring and I had all the setup ready to go :)

### Caching Strategy

Currently the extraction result is cached per URL, expiring after 1 hour. There's no stale refresh or background refresh mechanism, but it would be easy to add.

### API / CLI Separation

I chose to separate the API and the CLI into two separate main.go files. The reason is that both use slightly different environment variables and dependencies. Trying to extract some common code and share some "setup code" would be doable, but not really bringing much value at this point.

### Unit Testing

The `extractDataFromString` function is ready to be tested but since it's a function that depends on the OpenAI API and is going to be extremely slow to run in a CI, we may want to not just unit test it, but rather run it in a different separated pipeline.

## Next Steps

### Polish The Extraction

The current extraction is very basic and I did little prompt engineering. It would need more battle testing, possibly some post-processing to cleanup the data, why not some custom scraping logic for some websites or even some ld+json extraction.

### Data Modelization

Currently, the data is stored per url and per run in the database. A next step would be to modelize the data per entity (company, person, etc.) instead and aggregate the data accordingly.
