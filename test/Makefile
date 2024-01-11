DOCKERCOMPOSE := docker-compose -f docker-compose.yml
DOCKERCOMPOSEAPPETHTXMANAGER := zkevm-ethtx-manager
DOCKERCOMPOSEETHTXMANAGERDB := zkevm-ethtx-manager-db
DOCKERCOMPOSENETWORK := zkevm-mock-l1-network

RUNETHTXMANAGERDB := $(DOCKERCOMPOSE) up -d $(DOCKERCOMPOSEETHTXMANAGERDB)
RUNETHTXMANAGER := $(DOCKERCOMPOSE) up -d $(DOCKERCOMPOSEAPPETHTXMANAGER)
RUNL1NETWORK := $(DOCKERCOMPOSE) up -d $(DOCKERCOMPOSENETWORK)
RUNMETRICS := $(DOCKERCOMPOSE) up -d $(DOCKERCOMPOSEMETRICS)

RUN := $(DOCKERCOMPOSE) up -d

STOPETHTXMANAGERDB := $(DOCKERCOMPOSE) stop $(DOCKERCOMPOSEETHTXMANAGERDB) && $(DOCKERCOMPOSE) rm -f $(DOCKERCOMPOSEETHTXMANAGERDB)
STOPETHTXMANAGER := $(DOCKERCOMPOSE) stop $(DOCKERCOMPOSEAPPETHTXMANAGER) && $(DOCKERCOMPOSE) rm -f $(DOCKERCOMPOSEAPPETHTXMANAGER)
STOPNETWORK := $(DOCKERCOMPOSE) stop $(DOCKERCOMPOSENETWORK) && $(DOCKERCOMPOSE) rm -f $(DOCKERCOMPOSENETWORK)
STOPMETRICS := $(DOCKERCOMPOSE) stop $(DOCKERCOMPOSEMETRICS) && $(DOCKERCOMPOSE) rm -f $(DOCKERCOMPOSEMETRICS)

STOP := $(DOCKERCOMPOSE) down --remove-orphans

.PHONY: test-full # Runs all tests
test-full-non-e2e: stop 
	$(RUNETHTXMANAGERDB)
	sleep 2
	$(RUNL1NETWORK)
	sleep 15
	docker ps -a
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -count=1 -short -race -p 1 -covermode=atomic -coverprofile=../coverage.out -timeout 70s ../...


.PHONY: run-db
run-db: ## Runs the database
	$(RUNETHTXMANAGERDB)
	
.PHONY: stop-db
stop-db: ## Stops the database
	$(STOPETHTXMANAGERDB)

.PHONY: run-ethtx-manager
run-node: ## Runs the ethtxmanager
	$(RUNETHTXMANAGER)

.PHONY: stop-ethtx-manager
stop-node: ## Stops the ethtxmanager
	$(STOPETHTXMANAGER)

.PHONY: run-network
run-network: ## Runs the l1 network
	$(RUNL1NETWORK)

.PHONY: stop-network
stop-network: ## Stops the l1 network
	$(STOPNETWORK)

.PHONY: run
run: ## Runs a full node
	$(RUNETHTXMANAGERDB)
	$(RUNL1NETWORK)
	sleep 4
	$(RUNETHTXMANAGER)

.PHONY: stop
stop: ## Stops all services
	$(STOP)

.PHONY: restart
restart: stop run ## Executes `make stop` and `make run` commands

.PHONY: run-metrics
run-metrics: ## Runs the metrics container
	$(RUNMETRICS)

.PHONY: stop-metrics
stop-metrics: ## Stops the metrics container
	$(STOPMETRICS)

.PHONY: init-network
init-network: ## Initializes the network
	go run ./scripts/init_network/main.go .

.PHONY: show-logs
show-logs:  ## Show logs for running docker 
	$(DOCKERCOMPOSE) logs

.PHONY: run-db-scripts
run-db-scripts: ## Executes scripts on the db after it has been initialized, potentially using info from the environment
	./scripts/postgres/run.sh

## Help display.
## Pulls comments from beside commands and prints a nicely formatted
## display with the commands and their usage information.
.DEFAULT_GOAL := help

.PHONY: help
help: ## Prints this help
		@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
