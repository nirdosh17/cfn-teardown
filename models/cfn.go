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
package models

type StackDetails struct {
	StackName             string
	Status                string
	StackStatusReason     string // useful for failed cases
	DeleteStartedAt       string
	DeleteCompletedAt     string // must be fetched from describe status command. Wait time should not be considered
	DeletionTimeInMinutes string // total minutes taken to delete the stack
	DeleteAttempt         int16
	Exports               []string
	ActiveImporterStacks  map[string]struct{} // active(not deleted) stacks which are importing exports from THIS stack
	CFNConsoleLink        string
}

// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-describing-stacks.html
// Stack status and eligibility for deletion
var CREATE_IN_PROGRESS string = "CREATE_IN_PROGRESS"                                                     // Wait
var CREATE_FAILED string = "CREATE_FAILED"                                                               // Eligible for deletion
var CREATE_COMPLETE string = "CREATE_COMPLETE"                                                           // Eligible for deletion
var ROLLBACK_IN_PROGRESS string = "ROLLBACK_IN_PROGRESS"                                                 // Wait
var ROLLBACK_FAILED string = "ROLLBACK_FAILED"                                                           // Eligible for deletion
var ROLLBACK_COMPLETE string = "ROLLBACK_COMPLETE"                                                       // Eligible for deletion
var DELETE_IN_PROGRESS string = "DELETE_IN_PROGRESS"                                                     // Wait
var DELETE_FAILED string = "DELETE_FAILED"                                                               // Cannot be deleted. Manual Intervention Required. Post Message in RC.
var DELETE_COMPLETE string = "DELETE_COMPLETE"                                                           // Skip
var UPDATE_IN_PROGRESS string = "UPDATE_IN_PROGRESS"                                                     // Wait
var UPDATE_COMPLETE_CLEANUP_IN_PROGRESS string = "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS"                   // Wait
var UPDATE_COMPLETE string = "UPDATE_COMPLETE"                                                           // Eligible for deletion
var UPDATE_ROLLBACK_IN_PROGRESS string = "UPDATE_ROLLBACK_IN_PROGRESS"                                   // Wait
var UPDATE_ROLLBACK_FAILED string = "UPDATE_ROLLBACK_FAILED"                                             // Cannot be deleted. Manual Intervention Required. Post Message in RC.
var UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS string = "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS" // Wait
var UPDATE_ROLLBACK_COMPLETE string = "UPDATE_ROLLBACK_COMPLETE"                                         // Eligible for deletion
var REVIEW_IN_PROGRESS string = "REVIEW_IN_PROGRESS"                                                     // Wait
var IMPORT_IN_PROGRESS string = "IMPORT_IN_PROGRESS"                                                     // Wait
var IMPORT_COMPLETE string = "IMPORT_COMPLETE"                                                           // Wait
var IMPORT_ROLLBACK_IN_PROGRESS string = "IMPORT_ROLLBACK_IN_PROGRESS"                                   // Wait
var IMPORT_ROLLBACK_FAILED string = "IMPORT_ROLLBACK_FAILED"                                             // Cannot be deleted. Manual Intervention Required. Post Message in RC.
var IMPORT_ROLLBACK_COMPLETE string = "IMPORT_ROLLBACK_COMPLETE"                                         // Wait

// all statuses except DELETE_COMPLETE
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
