NAME=hunterio-test
.DEFAULT_GOAL = run-api

.PHONY: install-api
install-api: ## Installs the Go binary for development
	@go version
	GOGC=off go build -o $(GOPATH)/bin/$(NAME)-api -v ./cmd/api

.PHONY: run-api
run-api: install-api ## Runs the Go program for development
	source develop.env && $(NAME)-api

.PHONY: install-cli
install-cli: ## Installs the Go binary for development
	@go version
	GOGC=off go build -o $(GOPATH)/bin/$(NAME)-cli -v ./cmd/cli

.PHONY: tidy
tidy: ## Tidies the project
	go mod tidy && go mod vendor

.PHONY: db
db: ## Launches and migrates a development database
	source develop.env && NAME=$(NAME) make -C postgres db
