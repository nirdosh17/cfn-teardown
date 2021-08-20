/*
Copyright © 2021 Nirdosh Gautam

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/sts"

	. "github.com/nirdosh17/cfn-teardown/models"
)

type CFNManager struct {
	TargetAccountId string
	NukeRoleARN     string
	StackPattern    string
	AWSProfile      string
	AWSRegion       string
}

func (dm CFNManager) DescribeStack(stackName string) (*cloudformation.Stack, error) {
	cfn, err := dm.Session()
	if err != nil {
		return nil, err
	}

	resp, err := cfn.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: &stackName})
	if err != nil {
		return nil, err
	}
	return resp.Stacks[0], err
}

func (dm CFNManager) ListStackResources(stackName string) ([]*cloudformation.StackResourceSummary, error) {
	cfn, err := dm.Session()
	if err != nil {
		return nil, err
	}

	allResources := []*cloudformation.StackResourceSummary{}
	resp, err := cfn.ListStackResources(&cloudformation.ListStackResourcesInput{StackName: &stackName})
	if err != nil {
		return nil, err
	}
	allResources = append(allResources, resp.StackResourceSummaries...)

	nextToken := resp.NextToken
	for nextToken != nil {
		// sending next token for pagination
		resp, err := cfn.ListStackResources(&cloudformation.ListStackResourcesInput{StackName: &stackName, NextToken: nextToken})
		if err != nil {
			break
		}
		allResources = append(allResources, resp.StackResourceSummaries...)
		nextToken = resp.NextToken
	}

	if err != nil {
		fmt.Printf("Error listing resources of stack '%v': %v\n", stackName, err)
	}

	return resp.StackResourceSummaries, err
}

func (dm CFNManager) ListImports(exportNames []string) (map[string]struct{}, error) {
	importers := make(map[string]struct{})
	var err error
	cfn, err := dm.Session()
	if err != nil {
		return importers, err
	}

	for _, export := range exportNames {
		resp, err := cfn.ListImports(&cloudformation.ListImportsInput{ExportName: &export})
		if err != nil {
			// no imports = eligible for deletion
			if !strings.Contains(err.Error(), "is not imported by any stack") {
				return importers, err
			}
		}
		for _, stackName := range resp.Imports {
			// using map for faster access and empty struct due to its null memory consumption
			importers[*stackName] = struct{}{}
		}
	}

	return importers, err
}

// No error means, delete request sent to cloudformation
// If the stack we are trying to delete has already been deleted, returns success
func (dm CFNManager) DeleteStack(stackName string) error {
	fmt.Printf("Submitting delete request for stack: %v\n", stackName)
	cfn, err := dm.Session()
	if err != nil {
		return err
	}
	input := cloudformation.DeleteStackInput{StackName: &stackName}
	// stack delete output is an empty struct
	_, err = cfn.DeleteStack(&input)
	return err
}

func (dm CFNManager) ListEnvironmentStacks() (map[string]StackDetails, error) {
	CFNConsoleBaseURL := "https://console.aws.amazon.com/cloudformation/home?region=" + dm.AWSRegion + "#/stacks/stackinfo?stackId="

	// using stack name as key for easy traversal
	envStacks := map[string]StackDetails{}

	cfn, err := dm.Session()
	if err != nil {
		return envStacks, err
	}

	input := cloudformation.ListStacksInput{StackStatusFilter: ActiveStatuses}
	// only returns first 100 stacks. Need to use NextToken
	listStackOutput, err := cfn.ListStacks(&input)
	if err != nil {
		return envStacks, err
	}

	for _, details := range listStackOutput.StackSummaries {
		// select stacks of our concern
		stackName := *details.StackName
		if dm.RegexMatch(stackName) {
			sd := StackDetails{
				StackName:      stackName,
				Status:         *details.StackStatus,
				CFNConsoleLink: (CFNConsoleBaseURL + stackName),
			}
			envStacks[stackName] = sd
		}
	}

	if err != nil {
		fmt.Printf("Failed listing stacks with pattern: '%v', Error: '%v'\n", dm.StackPattern, err)
		return envStacks, err
	}

	nextToken := listStackOutput.NextToken
	for nextToken != nil {
		// sending next token for pagination
		input = cloudformation.ListStacksInput{NextToken: nextToken, StackStatusFilter: ActiveStatuses}
		listStackOutput, err = cfn.ListStacks(&input)
		if err != nil {
			break
		}
		for _, details := range listStackOutput.StackSummaries {
			// select stacks of our concern
			stackName := *details.StackName
			if dm.RegexMatch(stackName) {
				sd := StackDetails{
					StackName:      stackName,
					Status:         *details.StackStatus,
					CFNConsoleLink: (CFNConsoleBaseURL + stackName),
				}
				envStacks[stackName] = sd
			}
		}
		nextToken = listStackOutput.NextToken
	}

	if err != nil {
		fmt.Printf("Error listing '%v' environment stacks: %v\n", dm.StackPattern, err)
	}
	return envStacks, err
}

// { "stack-1-name": ["export-1", "export-2"], "stack-2-name": []}
func (dm CFNManager) ListEnvironmentExports() (map[string][]string, error) {
	exports := map[string][]string{}

	cfn, err := dm.Session()
	if err != nil {
		return exports, err
	}

	input := cloudformation.ListExportsInput{}
	// only returns first 100 stacks. Need to use NextToken
	listExportOutput, err := cfn.ListExports(&input)

	for _, details := range listExportOutput.Exports {
		stackArn := *details.ExportingStackId
		stackName := strings.Split(stackArn, "/")[1]
		exportName := *details.Name
		exports[stackName] = append(exports[stackName], exportName)
	}

	if err != nil {
		fmt.Printf("Error listing '%v' environment stack exports: %v\n", dm.StackPattern, err)
		return exports, err
	}

	nextToken := listExportOutput.NextToken

	for nextToken != nil {
		// sending next token for pagination
		input := cloudformation.ListExportsInput{NextToken: nextToken}
		listExportOutput, err = cfn.ListExports(&input)

		if err != nil {
			break
		}

		for _, details := range listExportOutput.Exports {
			stackArn := *details.ExportingStackId
			stackName := strings.Split(stackArn, "/")[1]
			exportName := *details.Name
			exports[stackName] = append(exports[stackName], exportName)
		}
		nextToken = listExportOutput.NextToken
	}
	return exports, err
}

func (dm CFNManager) RegexMatch(stackName string) bool {
	match, _ := regexp.MatchString(dm.StackPattern, stackName)
	return match
}

// assumes staging nuke role
func (dm CFNManager) Session() (*cloudformation.CloudFormation, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(dm.AWSRegion)},
		SharedConfigState: session.SharedConfigEnable,
		Profile:           dm.AWSProfile,
	}))

	// validation for target account id
	if dm.TargetAccountId != "" {
		aID, err := dm.AWSSessionAccountID(sess)
		if err != nil {
			fmt.Printf("Error requesting AWS caller identity: %v", err.Error())
			return nil, err
		}

		if aID != dm.TargetAccountId {
			return nil, fmt.Errorf(
				"[CFN] Target account id (%v) did not match with account id (%v) in the current AWS session",
				dm.TargetAccountId,
				aID,
			)
		}
	}

	if dm.NukeRoleARN == "" {
		// this means, we are using given aws profile
		return cloudformation.New(sess), nil
	} else {
		// Create the credentials from AssumeRoleProvider if nuke role arn is provided
		creds := stscreds.NewCredentials(sess, dm.NukeRoleARN)
		// Create service client value configured for credentials from assumed role.
		return cloudformation.New(sess, &aws.Config{Credentials: creds, MaxRetries: &AWS_SDK_MAX_RETRY}), nil
	}
}

func (dm CFNManager) AWSSessionAccountID(sess *session.Session) (acID string, err error) {
	svc := sts.New(sess)
	result, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		fmt.Printf("Error requesting AWS caller identity: %v", err.Error())
		return
	}
	acID = *result.Account
	return
}
