#!/bin/bash

set -e

sh -c /app/test/setup.sh

LOCALSTACK_ENDPOINT="${LOCALSTACK_ENDPOINT:-'http://localhost:4566'}"

ARGS="--endpoint-url $LOCALSTACK_ENDPOINT --region us-east-1"

expected_stacks="test-dynamodb,test-lambda,test-vpc"
stacks=$(aws cloudformation list-stacks $ARGS | jq -r '.StackSummaries | sort_by(.StackName) | map(.StackName) | join(",")')
if [ "$stacks" != "$expected_stacks" ]; then
  echo "expected stacks:"
  echo $expected_stacks
  echo "got:"
  echo $stacks
  exit 1
fi

function expect_vpc_count() {
  c=$(aws $ARGS ec2 describe-vpcs --filters "Name=tag:cfnteardown,Values=True" --query 'Vpcs | length(@)')
  if [ "$c" != "$1" ]; then
    echo "expected vpcs: $1, got: $c"
    exit 1
  fi
}

# verify cfn stack count by status
# e.g. expect_stack_count 'CREATE_COMPLETE' 3
function expect_stack_count() {
  c=$(aws --endpoint-url $LOCALSTACK_ENDPOINT --region us-east-1 cloudformation list-stacks --query "StackSummaries[?StackStatus=='$1'].StackName | length(@)")
  if [ "$c" != "$2" ]; then
    echo "expected stacks: $2, got: $c"
    exit 1
  fi
}

function expect_subnets_count() {
  c=$(aws $ARGS ec2 describe-subnets --filters "Name=tag:cfnteardown,Values=True" --query 'Subnets | length(@)')
  if [ "$c" != "$1" ]; then
    echo "expected subnets: $1, got: $c"
    exit 1
  fi
}

function expect_routetable_count() {
  c=$(aws $ARGS ec2 describe-route-tables --filters "Name=tag:cfnteardown,Values=True" --query 'RouteTables | length(@)')
  if [ "$c" != "$1" ]; then
    echo "expected route tables: $1, got: $c"
    exit 1
  fi
}

function expect_dynamodb_count() {
  c=$(aws $ARGS dynamodb list-tables --query 'TableNames | length(@)')
  if [ "$c" != "$1" ]; then
    echo "expected dynamodb tables: $1, got: $c"
    exit 1
  fi
}

function expect_lambda_count() {
  c=$(aws $ARGS lambda list-functions --query 'Functions | length(@)')
  if [ "$c" != "$1" ]; then
    echo "expected dynamodb tables: $1, got: $c"
    exit 1
  fi
}

echo "verifying resource count before deleting stacks..."
expect_stack_count CREATE_COMPLETE 3
expect_stack_count DELETE_COMPLETE 0
expect_vpc_count 1
expect_subnets_count 2
expect_routetable_count 2
expect_dynamodb_count 1
expect_lambda_count 1
echo "all good!"

./cfn-teardown deleteStacks --DRY_RUN=false

echo "verifying resource count after running cfn-teardown cli..."
expect_stack_count CREATE_COMPLETE 0
expect_stack_count DELETE_COMPLETE 3
expect_vpc_count 0
expect_subnets_count 0
expect_routetable_count 0
expect_dynamodb_count 0
expect_lambda_count 0
echo "all tests passed!"
