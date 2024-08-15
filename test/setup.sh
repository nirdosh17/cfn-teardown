#!/bin/bash

set -e

echo "setting up aws configs.."
mkdir -p ~/.aws

echo "[default]
region = us-east-1" > ~/.aws/config

echo "[default]
aws_access_key_id = randomstringforlocalstack
aws_secret_access_key = randomstringforlocalstack" > ~/.aws/credentials
echo "aws configs written!"

echo "creating test config for cli..."
echo "AWS_REGION: us-east-1
AWS_PROFILE: default
STACK_PATTERN: test-
ENDPOINT_URL: http://localstack:4566
ABORT_WAIT_TIME_MINUTES: 0
STACK_WAIT_TIME_SECONDS: 2" > ~/.cfn-teardown.yaml
echo "configs written at: ~/.cfn-teardown.yaml"

echo "building binary..."
make build
echo "binary built!"

echo "creating test stacks..."
./test/create_test_stacks.sh
echo "stacks created!"
