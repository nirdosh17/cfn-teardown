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

// Config represents all the parameters supported by cfn-teardown
type Config struct {
	AWSProfile             string `mapstructure:"awsProfile"`
	AWSRegion              string `mapstructure:"awsRegion"`
	AWSAccountId           string `mapstructure:"awsAccountId"`
	StackPattern           string `mapstructure:"stackPattern"`
	StackWaitTimeSeconds   int16  `mapstructure:"stackWaitTimeSeconds"`
	MaxDeleteRetryCount    int16  `mapstructure:"maxDeleteRetryCount"`
	AbortWaitTimeMinutes   int16  `mapstructure:"abortWaitTimeMinutes"`
	NotificationWebhookURL string `mapstructure:"notificationWebhookURL"`
	RoleARN                string `mapstructure:"roleARN"`
	DryRun                 string `mapstructure:"dryRun"`
}
