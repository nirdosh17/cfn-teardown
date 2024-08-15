#!/bin/bash

set -e

LOCALSTACK_ENDPOINT="${LOCALSTACK_ENDPOINT:-'http://localhost:4566'}"

ARGS="--endpoint-url $LOCALSTACK_ENDPOINT --region us-east-1"
ENV_NAME="test"
VPC_STACK_NAME=$ENV_NAME-vpc
DDB_STACK_NAME=$ENV_NAME-dynamodb
LAMBDA_STACK_NAME=$ENV_NAME-lambda

TEMPLATES_FOLDER="test/templates"

echo "--- test stacks creation started ---"

aws cloudformation create-stack --stack-name $VPC_STACK_NAME --template-body file://$TEMPLATES_FOLDER/vpc.yaml $ARGS > /dev/null
aws cloudformation wait stack-create-complete --stack-name $VPC_STACK_NAME $ARGS
echo "vpc created!"

aws cloudformation create-stack --stack-name $DDB_STACK_NAME --template-body file://$TEMPLATES_FOLDER/dynamodb.yaml $ARGS > /dev/null
aws cloudformation wait stack-create-complete --stack-name $DDB_STACK_NAME $ARGS
echo "dynamodb created!"

aws cloudformation create-stack --stack-name $LAMBDA_STACK_NAME --template-body file://$TEMPLATES_FOLDER/lambda.yaml $ARGS > /dev/null
aws cloudformation wait stack-create-complete --stack-name $LAMBDA_STACK_NAME $ARGS
echo "lambda created!"

echo "--- test stacks created! ---"
