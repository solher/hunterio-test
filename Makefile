NAME=hunterio-test
.DEFAULT_GOAL = run

.PHONY: install
install: ## Installs the Go binary for development
	@go version
	GOGC=off go install -v

.PHONY: run
run: install ## Runs the Go program for development
	source develop.env && $(NAME)

.PHONY: tidy
tidy: ## Tidies the project
	go mod tidy && go mod vendor

.PHONY: db
db: ## Launches and migrates a development database
	source develop.env && NAME=$(NAME) make -C postgres db
