.DEFAULT_GOAL=help

build:  ## Download packages and build binary
	go mod download && \
	go build -o cfn-teardown .

run: build ## Build and run binary
	./cfn-teardown

# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
