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

// Config represents all the parameters supported by cfn-teardown
type Config struct {
	AWSProfile           string `mapstructure:"AWS_PROFILE"`
	AWSRegion            string `mapstructure:"AWS_REGION"`
	TargetAccountId      string `mapstructure:"TARGET_ACCOUNT_ID"`
	StackPattern         string `mapstructure:"STACK_PATTERN"`
	StackWaitTimeSeconds int16  `mapstructure:"STACK_WAIT_TIME_SECONDS"`
	MaxDeleteRetryCount  int16  `mapstructure:"MAX_DELETE_RETRY_COUNT"`
	AbortWaitTimeMinutes int16  `mapstructure:"ABORT_WAIT_TIME_MINUTES"`
	SlackWebhookURL      string `mapstructure:"SLACK_WEBHOOK_URL"`
	RoleARN              string `mapstructure:"ROLE_ARN"`
	DryRun               string `mapstructure:"DRY_RUN"`
}
