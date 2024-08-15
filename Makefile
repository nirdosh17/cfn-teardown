.DEFAULT_GOAL=help

build: ## install deps and build binary
	go mod download && \
	go build -o cfn-teardown .

run: build ## build and run binary
	./cfn-teardown

test.start: test.stop ## start integration test
	docker build --platform linux/amd64 -t cfn-teardown-test -f test/Dockerfile .
	docker compose -f test/docker-compose.yaml up --abort-on-container-exit --remove-orphans

test.stop: ## stop integration test
	docker compose -f test/docker-compose.yaml down

help:
	@sed -ne '/@sed/!s/## //p' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ": "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
