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

// Package utils provides cli specifics methods for interacting with AWS services
package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/nirdosh17/cfn-teardown/models"
)

// -------------- configs ---------------
const (
	STACK_DELETION_WAIT_TIME_IN_SEC int16 = 30 // STACK_DELETION_WAIT_TIME_IN_SEC is the time to wait for stacks before peforming status checks after delete requests have been sent.
	MAX_DELETE_RETRY_COUNT          int16 = 5  // MAX_DELETE_RETRY_COUNT specifies the number of times we should retry deleting a stack before giving up.
)

var (
	NUKE_START_TIME       = CurrentUTCDateTime() // NUKE_START_TIME is the start timestamp of teardown.
	NUKE_END_TIME         = CurrentUTCDateTime() // NUKE_END_TIME is the end timestamp of teardown.
	AWS_SDK_MAX_RETRY int = 5                    // AWS_SDK_MAX_RETRY is max retry count for AWS SDK.

	// stats
	TOTAL_STACK_COUNT    int     // TOTAL_STACK_COUNT is the number of stacks found to be eligible for deletion/
	DELETED_STACK_COUNT  int     // DELETED_STACK_COUNT is the number of stacks deleted so far.
	ACTIVE_STACK_COUNT   int     // ACTIVE_STACK_COUNT is the number of stacks yet to be deleted or in the process of being deleted.
	NUKE_DURATION_IN_HRS float64 // NUKE_DURATION_IN_HRS is the total run time of teardown until now.
)

// InitiateTearDown scans and deletes cloudformation stacks respecting the dependencies.
// A stack is eligible for deletion when it's exports has not been imported by any other stacks.
func InitiateTearDown(config models.Config) {
	cfn := CFNManager{StackPattern: config.StackPattern, TargetAccountId: config.TargetAccountId, NukeRoleARN: config.RoleARN, AWSProfile: config.AWSProfile, AWSRegion: config.AWSRegion}
	s3 := S3Manager{TargetAccountId: config.TargetAccountId, NukeRoleARN: config.RoleARN, AWSProfile: config.AWSProfile, AWSRegion: config.AWSRegion}
	notifier := NotificationManager{StackPattern: config.StackPattern, SlackWebHookURL: config.SlackWebhookURL, DryRun: config.DryRun}

	var dependencyTree = map[string]models.StackDetails{}

	// generate dependencies for matching stacks
	dt, err := prepareDependencyTree(config.StackPattern, cfn)

	if err != nil {
		UpdateNukeStats(dependencyTree)
		msg := fmt.Sprintf("Unable to prepare dependencies. Error: %v", err.Error())
		notifier.ErrorAlert(AlertMessage{Message: msg})
		color.Error.Println(msg)
		os.Exit(1)
	}
	dependencyTree = dt // need to do this for global scope
	writeToJSON(config.StackPattern, dependencyTree)

	TOTAL_STACK_COUNT = len(dependencyTree)
	UpdateNukeStats(dependencyTree)

	if ACTIVE_STACK_COUNT == 0 {
		UpdateNukeStats(dependencyTree)
		color.Yellow.Printf("\nNo matching stacks to delete! Stack count: %v\n", TOTAL_STACK_COUNT)
		notifier.SuccessAlert(AlertMessage{})
		return
	}

	fmt.Println()
	fmt.Printf("Following stacks are eligible for deletion | Stack count: %v\n", ACTIVE_STACK_COUNT)
	for stackName := range dependencyTree {
		color.Gray.Println(" -", stackName)
	}
	color.Style{color.Yellow, color.OpItalic}.Println("\nCheck 'stack_teardown_details.json' file for more details.")
	fmt.Println()

	// safety check for accidental run
	if config.DryRun != "false" {
		return
	}

	msg := fmt.Sprintf("Waiting for `%v minutes` before starting deletion. Abort if necessary.", config.AbortWaitTimeMinutes)
	notifier.StartAlert(AlertMessage{Message: msg})
	color.Red.Println(msg)
	time.Sleep(time.Duration(config.AbortWaitTimeMinutes) * time.Minute)
	color.Green.Println("\n\n---------------------------- Deletion Started -------------------------------")
	for {
		// Algorithm:
		// 1. Scan stacks who has zero importing stacks i.e. last leaf in the dependency tree
		toDelete := stacksEligibleToDelete(dependencyTree)

		// 2. Delete stacks
		//    2.1 If stack has S3 bucket resource, then delete bucket contents first
		//    2.2 Then send request to delete stack
		//    2.3 Change stack status to DELETE_IN_PROGRESS
		fmt.Println("\n-----------------------------------------------------------------------------")
		fmt.Printf("Searching stacks with no importers(dependencies): %v\n", len(toDelete))
		for _, sName := range toDelete {
			stack := dependencyTree[sName]
			bktErr := deleteBucketIfPresent(sName, cfn, s3)
			if bktErr != nil {
				stack.StackStatusReason = bktErr.Error()
				UpdateNukeStats(dependencyTree)
				msg := fmt.Sprintf("Unable to empty bucket from stack '%v'", sName)
				notifier.ErrorAlert(AlertMessage{Message: msg, FailedStack: stack})
				color.Error.Println(msg)
				os.Exit(1)
			}

			err := cfn.DeleteStack(sName)
			if err != nil {
				UpdateNukeStats(dependencyTree)
				msg = fmt.Sprintf("Unable to send delete request for stack '%v' Error: %v", sName, err)
				stack.StackStatusReason = msg
				notifier.ErrorAlert(AlertMessage{Message: msg, FailedStack: stack})
				color.Error.Println(msg)
				os.Exit(1)
			}
			stack.Status = models.DELETE_IN_PROGRESS
			stack.DeleteStartedAt = CurrentUTCDateTime()
			stack.DeleteAttempt = stack.DeleteAttempt + 1
			dependencyTree[sName] = stack
			writeToJSON(config.StackPattern, dependencyTree)
		}

		// 3. Wait for 30 seconds
		fmt.Println("\n-----------------------------------------------------------------------------")
		fmt.Printf("Waiting for %v seconds...\n", STACK_DELETION_WAIT_TIME_IN_SEC)
		time.Sleep(time.Duration(STACK_DELETION_WAIT_TIME_IN_SEC) * time.Second)

		// 4. Get list of stacks in DELETE_IN_PROGRESS and describe stack
		//     4.1. If status is still DELETE_IN_PROGRESS, skip
		// 		 4.2. If stack is not found or already deleted
		//         4.2.1 Change status to DELETE_COMPLETE
		//         4.2.2 Remove stack from importer list
		//     4.3. If stack status is not 'DELETE_IN_PROGRESS' or 'DELETE_COMPLETE'
		//         Mark this as failure. Get stack reason. Alert in the notification channel. Abort env deletion.
		dipStacks := deleteInProgressStacks(dependencyTree)
		for _, sName := range dipStacks {
			stack := dependencyTree[sName]
			// fetch latest stack details
			details, err := cfn.DescribeStack(sName)

			var dne bool
			if err != nil {
				// this error means stack has already been deleted
				dne = strings.Contains(err.Error(), "does not exist")
				// means that the error is related to SDK. in that case we would want to notify error and exit
				if !dne {
					UpdateNukeStats(dependencyTree)
					msg := fmt.Sprintf("Unable to describe stack '%v'", sName)
					stack.StackStatusReason = msg
					notifier.ErrorAlert(AlertMessage{Message: msg, FailedStack: stack})
					color.Error.Println(msg)
					os.Exit(1)
				}
			}

			var newStatus string
			// does not exist means the stack was deleted
			if dne {
				newStatus = models.DELETE_COMPLETE
			} else {
				newStatus = *details.StackStatus
			}

			if newStatus == models.DELETE_IN_PROGRESS {
				// skip now. check again later
				continue
			} else if newStatus == models.DELETE_COMPLETE {
				// update local copy
				stack.Status = newStatus
				stack.DeleteCompletedAt = CurrentUTCDateTime()
				stack.DeletionTimeInMinutes = TimeDiff(stack.DeleteStartedAt, stack.DeleteCompletedAt)

				// updating stack details to dependency tree
				dependencyTree[sName] = stack

				// removing this stack from list of importers of all stacks and updating dependency tree
				dependencyTree = updateImporterList(sName, dependencyTree)
				writeToJSON(config.StackPattern, dependencyTree)
				fmt.Printf("Stack successfully deleted: %v\n", sName)
			} else {
				if stack.DeleteAttempt >= MAX_DELETE_RETRY_COUNT {
					stack.Status = newStatus
					statusReason := *details.StackStatusReason
					stack.StackStatusReason = statusReason

					dependencyTree[sName] = stack
					writeToJSON(config.StackPattern, dependencyTree)

					UpdateNukeStats(dependencyTree)
					msg := fmt.Sprintf("Failed to delete stack `%v`. Reason: %v", sName, statusReason)
					notifier.ErrorAlert(AlertMessage{Message: msg, FailedStack: stack})
					color.Error.Println(msg)
					os.Exit(1)
				} else {
					// In some cases cloud9 stacks can't be deleted due to security group being manually attached to other resources like elastic search or redis
					// In such case it is better to wait for dependent resource's(mostly datastore or cache) stack and security group to get deleted and retry again
					newDeleteAttempt := stack.DeleteAttempt + 1
					fmt.Printf("Retrying deleting stack: %v Delete Attempt: %v/%v\n", sName, newDeleteAttempt, MAX_DELETE_RETRY_COUNT)
					err := cfn.DeleteStack(sName)
					if err != nil {
						UpdateNukeStats(dependencyTree)
						msg = fmt.Sprintf("Unable to send delete retry request for stack '%v' Error: %v", sName, err)
						stack.StackStatusReason = msg
						notifier.ErrorAlert(AlertMessage{Message: msg, FailedStack: stack})
						color.Error.Println(msg)
						os.Exit(1)
					}
					stack.Status = models.DELETE_IN_PROGRESS
					stack.DeleteStartedAt = CurrentUTCDateTime()
					stack.DeleteAttempt = newDeleteAttempt
					dependencyTree[sName] = stack
					writeToJSON(config.StackPattern, dependencyTree)
				}
			}
		}

		// 5. If all stacks have already been deleted, stop execution. Else Go to step 1
		if isEnvNuked(dependencyTree) {
			UpdateNukeStats(dependencyTree)
			color.Green.Printf("\n---------- STACK TEARDOWN SUCCESSFUL! STACKS DELETED: (%v) ----------\n\n", DELETED_STACK_COUNT)
			notifier.SuccessAlert(AlertMessage{})
			break
		}

		// 6. Check if nuke is stuck
		if isNukeStuck(dependencyTree) {
			UpdateNukeStats(dependencyTree)
			// TODO: better messaging
			msg := "No stacks are eligible for deletion. Please find and delete stacks which do not have follow given pattern: " + config.StackPattern
			notifier.StuckAlert(AlertMessage{Message: msg})
			color.Error.Println(msg)
			os.Exit(1)
			break
		}
	}
}

// When a stack is deleted, we can safely remove it from list of importers
// so that the parent stack is free of dependencies and becomes eligible for deletion in the next cycle.
func updateImporterList(deletedStackName string, dt map[string]models.StackDetails) map[string]models.StackDetails {
	for _, stackDetails := range dt {
		importers := stackDetails.ActiveImporterStacks
		delete(importers, deletedStackName)
		stackDetails.ActiveImporterStacks = importers
	}
	return dt
}

// In order to a stack with S3 bucket, we need to empty it first which is done by this method.
func deleteBucketIfPresent(stackName string, cfn CFNManager, s3 S3Manager) error {
	resources, _ := cfn.ListStackResources(stackName)

	var objDeleteError error
	for _, resource := range resources {
		// if a stack is in ROLLBACK_COMPLETE state. Some of the resources might not have physical resource ID
		// so checking this first. If there is no resource. No need to empty the bucket
		if resource.PhysicalResourceId != nil && resource.ResourceType != nil {
			rType := *resource.ResourceType
			rName := *resource.PhysicalResourceId
			// bucket should be empty before we delete the cfn stack, thus emptying bucket here
			if rType == "AWS::S3::Bucket" {
				objDeleteError = s3.EmptyBucket(rName)
				if objDeleteError != nil {
					msg := fmt.Sprintf("Failed to empty bucket '%v' from stack '%v'. Error: %v", rName, stackName, objDeleteError.Error())
					fmt.Println(msg)
					break
				}
			}
		}
	}
	return objDeleteError
}

// In some cases, there could be no stacks which are eligible for deletion. This can happen due to cyclic dependency. In such case, we abort nuke and notify the user for manual intervention.
func isNukeStuck(dt map[string]models.StackDetails) bool {
	if len(deleteInProgressStacks(dt)) == 0 && len(stacksEligibleToDelete(dt)) == 0 {
		return true
	}
	return false
}

// stacksEligibleToDelete selects stacks for deletion which have no dependencies
func stacksEligibleToDelete(dt map[string]models.StackDetails) []string {
	deleteReady := []string{}
	for stackName, stackDetails := range dt {
		if len(stackDetails.ActiveImporterStacks) == 0 {
			// not filtering out delete failed here as it is being handled in main.go
			if stackDetails.Status != models.DELETE_COMPLETE && stackDetails.Status != models.DELETE_IN_PROGRESS {
				deleteReady = append(deleteReady, stackName)
			}
		}
	}
	return deleteReady
}

func deleteInProgressStacks(dt map[string]models.StackDetails) []string {
	dip := []string{}
	for stackName, stackDetails := range dt {
		if stackDetails.Status == models.DELETE_IN_PROGRESS {
			dip = append(dip, stackName)
		}
	}
	return dip
}

// isEnvNuked checks if all stacks have status DELETE_COMPLETE to mark the end of teardown
func isEnvNuked(dt map[string]models.StackDetails) bool {
	nuked := true
	for _, stackDetails := range dt {
		if stackDetails.Status != models.DELETE_COMPLETE {
			nuked = false
			break
		}
	}
	return nuked
}

// prepareDependencyTree generates list of stacks and their dependencies which is useful to determine the order of deletion
func prepareDependencyTree(envLabel string, cfn CFNManager) (map[string]models.StackDetails, error) {
	CFNConsoleBaseURL := "https://console.aws.amazon.com/cloudformation/home?region=" + cfn.AWSRegion + "#/stacks/stackinfo?stackId="

	fmt.Printf("-------------- Listing Stacks | Match Pattern: [%v] --------------\n", color.Gray.Render(envLabel))

	dependencyTree, err := cfn.ListEnvironmentStacks()
	totalStackCount := len(dependencyTree)

	if err != nil {
		UpdateNukeStats(dependencyTree)
		color.Error.Printf("  Failed listing stacks! Error: %v\n", err)
		return dependencyTree, err
	}

	color.Gray.Println("  Listing all exports...")
	stackExports, err := cfn.ListEnvironmentExports()
	if err != nil {
		color.Error.Printf("  Failed listing exports! Error: %v", err)
		return dependencyTree, err
	}

	color.Gray.Println("  Listing all imports...")
	stackCount := 0
	var listImportErr error
	for stackName, stack := range dependencyTree {
		// populate exports
		if _, ok := stackExports[stackName]; ok {
			if len(stackExports[stackName]) > 0 {
				stack.Exports = stackExports[stackName]
			}
		}

		// listing all importers. making single api call at a time to avoid rate limiting
		importingStacks, listImportErr := cfn.ListImports(stack.Exports)
		if listImportErr != nil {
			color.Error.Printf("  Failed listing imports! Error: %v", listImportErr)
			break
		}

		stack.ActiveImporterStacks = importingStacks
		dependencyTree[stackName] = stack
		stackCount++
		color.Gray.Println("  Listing imports | ", stackCount, "/", totalStackCount, " stacks complete")
	}

	if listImportErr != nil {
		return dependencyTree, listImportErr
	}

	// check if any stack is present in the importers list but not present in the dependency tree. If yes add it to dependency tree along with its dependent stacks
	// 		this can happen if a stackname does not begin match with given pattern i.e. not following the naming convention
	missing := getStackWithMissingDependencies(dependencyTree)
	for len(missing) != 0 {
		// TODO: better logging for this. include this in readme as well
		// fmt.Printf("Stack '%v' does not match pattern '%v' and imports from stacks selected for deletion", missing, cfn.EnvLabel)
		// fmt.Printf("Included '%v' stack in the deletion list", missing)
		for mStk := range missing {
			totalStackCount++
			sDetails, err := cfn.DescribeStack(mStk)
			if err != nil {
				dne := strings.Contains(err.Error(), "does not exist")
				if !dne {
					color.Error.Printf("  Error describing stack %v", mStk)
					break // real error.
				}
				dependencyTree[mStk] = models.StackDetails{
					StackName:      mStk,
					Status:         "DELETE_COMPLETE",
					CFNConsoleLink: (CFNConsoleBaseURL + mStk),
				}
			} else {
				// list exports
				exports := []string{}
				for _, output := range sDetails.Outputs {
					exports = append(exports, *output.ExportName)
				}

				// list imports
				importingStacks, listImportErr := cfn.ListImports(exports)
				if listImportErr != nil {
					color.Error.Println("  Failed listing imports!")
					break
				}

				dependencyTree[mStk] = models.StackDetails{
					StackName:            mStk,
					Status:               *sDetails.StackStatus,
					Exports:              exports,
					ActiveImporterStacks: importingStacks,
					CFNConsoleLink:       (CFNConsoleBaseURL + mStk),
				}
			}
		}
		missing = getStackWithMissingDependencies(dependencyTree)
	}

	return dependencyTree, listImportErr
}

// --------------------- Utility functions ---------------------------

func getStackWithMissingDependencies(dt map[string]models.StackDetails) map[string]struct{} {
	allImporterStacks := map[string]struct{}{}
	notListed := map[string]struct{}{}
	for _, details := range dt {
		ais := details.ActiveImporterStacks
		for k := range ais {
			allImporterStacks[k] = struct{}{}
		}
	}

	// select importer stacks which are not listed in dependency tree
	for sn := range allImporterStacks {
		if _, ok := dt[sn]; !ok {
			notListed[sn] = struct{}{}
		}
	}

	return notListed
}

func writeToJSON(envLabel string, data map[string]models.StackDetails) {
	file, _ := json.MarshalIndent(data, "", " ")
	_ = ioutil.WriteFile("stack_teardown_details.json", file, 0644)
}

// CurrentUTCDateTime returns current time in ISO string
func CurrentUTCDateTime() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// TimeDiff returns difference of two timestamps in minutes
func TimeDiff(startTime, endTime string) string {
	st, _ := time.Parse(time.RFC3339, startTime)
	et, _ := time.Parse(time.RFC3339, endTime)
	diff := et.Sub(st)
	return fmt.Sprintf("%.2f", diff.Minutes())
}

// UpdateNukeStats updates global variables used for capturing teardown stats
func UpdateNukeStats(dt map[string]models.StackDetails) {
	NUKE_END_TIME = CurrentUTCDateTime()
	st, _ := time.Parse(time.RFC3339, NUKE_START_TIME)
	et, _ := time.Parse(time.RFC3339, NUKE_END_TIME)
	NUKE_DURATION_IN_HRS = et.Sub(st).Hours()

	deletedStackCount := 0
	for _, stackDetails := range dt {
		if stackDetails.Status == models.DELETE_COMPLETE {
			deletedStackCount++
		}
	}
	DELETED_STACK_COUNT = deletedStackCount
	ACTIVE_STACK_COUNT = TOTAL_STACK_COUNT - DELETED_STACK_COUNT
}
