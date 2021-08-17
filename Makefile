build:
	go mod download && \
	go build -o cfn-teardown .

run: build
	./cfn-teardown
