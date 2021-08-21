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

// Package models has definition of entities used in the process of teardown
package models

// StackDetails represents a cloudformation stack, it's state and dependencies.
type StackDetails struct {
	StackName             string
	Status                string
	StackStatusReason     string // useful for failed cases
	DeleteStartedAt       string
	DeleteCompletedAt     string
	DeletionTimeInMinutes string
	DeleteAttempt         int16
	Exports               []string
	ActiveImporterStacks  map[string]struct{} // active(not deleted) stacks which are importing exports from this stack
	CFNConsoleLink        string
}

// ---------- Stack statuses and their eligibility for deletion ------------
// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-describing-stacks.html

// CREATE_IN_PROGRESS stack status requires waiting before sending delete request.
var CREATE_IN_PROGRESS string = "CREATE_IN_PROGRESS"

// CREATE_FAILED stack status is eligible for deletion.
var CREATE_FAILED string = "CREATE_FAILED"

// CREATE_COMPLETE stack status is eligible for deletion.
var CREATE_COMPLETE string = "CREATE_COMPLETE"

// ROLLBACK_IN_PROGRESS stack status requires waiting before sending delete request.
var ROLLBACK_IN_PROGRESS string = "ROLLBACK_IN_PROGRESS"

// ROLLBACK_FAILED stack status is eligible for deletion.
var ROLLBACK_FAILED string = "ROLLBACK_FAILED"

// ROLLBACK_COMPLETE stack status is eligible for deletion.
var ROLLBACK_COMPLETE string = "ROLLBACK_COMPLETE"

// DELETE_IN_PROGRESS stack status requires waiting before taking any action
var DELETE_IN_PROGRESS string = "DELETE_IN_PROGRESS"

// DELETE_FAILED stack status after max delete attempts is unactionable and requires manual intervention
var DELETE_FAILED string = "DELETE_FAILED"

// DELETE_COMPLETE stack status can be skipped as the stack has already been deleted
var DELETE_COMPLETE string = "DELETE_COMPLETE"

// UPDATE_IN_PROGRESS stack status requires waiting before sending delete request.
var UPDATE_IN_PROGRESS string = "UPDATE_IN_PROGRESS"

// UPDATE_COMPLETE_CLEANUP_IN_PROGRESS stack status requires waiting before sending delete request.
var UPDATE_COMPLETE_CLEANUP_IN_PROGRESS string = "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS"

// UPDATE_COMPLETE stack status is eligible for deletion.
var UPDATE_COMPLETE string = "UPDATE_COMPLETE"

// UPDATE_ROLLBACK_IN_PROGRESS stack status requires waiting before sending delete request.
var UPDATE_ROLLBACK_IN_PROGRESS string = "UPDATE_ROLLBACK_IN_PROGRESS"

// UPDATE_ROLLBACK_FAILED stack status is eligible for deletion.
var UPDATE_ROLLBACK_FAILED string = "UPDATE_ROLLBACK_FAILED"

// UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS stack status requires waiting before sending delete request.
var UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS string = "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS"

// UPDATE_ROLLBACK_COMPLETE stack status is eligible for deletion.
var UPDATE_ROLLBACK_COMPLETE string = "UPDATE_ROLLBACK_COMPLETE"

// REVIEW_IN_PROGRESS stack status requires waiting before sending delete request.
var REVIEW_IN_PROGRESS string = "REVIEW_IN_PROGRESS"

// IMPORT_IN_PROGRESS stack status requires waiting before sending delete request.
var IMPORT_IN_PROGRESS string = "IMPORT_IN_PROGRESS"

// IMPORT_COMPLETE stack status requires waiting before sending delete request.
var IMPORT_COMPLETE string = "IMPORT_COMPLETE"

// IMPORT_ROLLBACK_IN_PROGRESS stack status requires waiting before sending delete request.
var IMPORT_ROLLBACK_IN_PROGRESS string = "IMPORT_ROLLBACK_IN_PROGRESS"

// IMPORT_ROLLBACK_FAILED stack status is eligible for deletion.
var IMPORT_ROLLBACK_FAILED string = "IMPORT_ROLLBACK_FAILED"

// IMPORT_ROLLBACK_COMPLETE stack status requires waiting before sending delete request.
var IMPORT_ROLLBACK_COMPLETE string = "IMPORT_ROLLBACK_COMPLETE"

// ActiveStatuses includes all stack statuses except DELETE_COMPLETE
var ActiveStatuses = []*string{
	&CREATE_IN_PROGRESS,
	&CREATE_FAILED,
	&CREATE_COMPLETE,
	&ROLLBACK_IN_PROGRESS,
	&ROLLBACK_FAILED,
	&ROLLBACK_COMPLETE,
	&DELETE_IN_PROGRESS,
	&DELETE_FAILED,
	&UPDATE_IN_PROGRESS,
	&UPDATE_COMPLETE_CLEANUP_IN_PROGRESS,
	&UPDATE_COMPLETE,
	&UPDATE_ROLLBACK_IN_PROGRESS,
	&UPDATE_ROLLBACK_FAILED,
	&UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS,
	&UPDATE_ROLLBACK_COMPLETE,
	&REVIEW_IN_PROGRESS,
	&IMPORT_IN_PROGRESS,
	&IMPORT_COMPLETE,
	&IMPORT_ROLLBACK_IN_PROGRESS,
	&IMPORT_ROLLBACK_FAILED,
	&IMPORT_ROLLBACK_COMPLETE,
}
