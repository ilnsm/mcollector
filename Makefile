.DEFAULT_GOAL := test

export GOLANGCI_LINT_CACHE=${PWD}/golangci-lint/.cache
export SHELL=/bin/zsh
.PHONY: agent server test lint run_server run_agent
agent:
	@go build -o cmd/agent/agent ./cmd/agent/*.go

.PHONY: server
server:
	@go build -o cmd/server/server ./cmd/server/*.go

.PHONY: run_server
run_server: server postgres
	@./cmd/server/server

.PHONY: run_agent
run_agent: agent
	@./cmd/agent/agent -p 1 -r 1 -log debug

.PHONY: test
test: server agent
	./runTest.sh |tee test.result

.PHONY: postgres
postgres:
	@docker compose up -d

.PHONY: lint
lint: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.55.2 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint matted.json > ./golangci-lint/report.json

.PHONY:
truncate:
	@docker exec postgres psql -U mcollector -d metrics -c 'truncate table counters, gauges;'