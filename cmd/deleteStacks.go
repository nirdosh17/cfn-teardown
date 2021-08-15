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
package cmd

import (
	"fmt"

	"github.com/nirdosh17/cfn-teardown/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deleteStacksCmd represents the deleteStacks command
var deleteStacksCmd = &cobra.Command{
	Use:   "deleteStacks",
	Short: "Deletes matching cloudformation stacks",
	Long: `Deletes cloudformation stacks and their dependencies whose names match with the given pattern.
The pattern must be provided otherwise it will scan all stacks in the region.
Example:
If your stacks to be deleted follow this naming convention: qa-{{component name}}
Supply stack pattern as: 'qa-'
	`,
	Example: "cfn-teardown deleteStacks --stackPattern='qa-' --awsProfile='staging' --region=us-east-1",
	Args: func(cmd *cobra.Command, args []string) error {
		// validate your arguments here
		return validateConfigs(config)
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Executing command: deleteStacks")
		fmt.Println()
		utils.InitiateTearDown(config)
	},
}

func init() {
	rootCmd.AddCommand(deleteStacksCmd)

	deleteStacksCmd.Flags().Int("stackWaitTimeSeconds", 30, "Seconds to wait after delete requests are submitted to CFN")
	viper.BindPFlag("stackWaitTimeSeconds", deleteStacksCmd.Flags().Lookup("stackWaitTimeSeconds"))

	deleteStacksCmd.Flags().String("awsAccountId", "", "[Safety Check] Validates against account id in current aws session and provided ID")
	viper.BindPFlag("awsAccountId", deleteStacksCmd.Flags().Lookup("awsAccountId"))

	deleteStacksCmd.Flags().Int("maxDeleteRetryCount", 5, "Max stack delete attempts")
	viper.BindPFlag("maxDeleteRetryCount", deleteStacksCmd.Flags().Lookup("maxDeleteRetryCount"))

	deleteStacksCmd.Flags().Int("abortWaitTimeMinutes", 10, "[Safety Check] Minutes to wait before initiating deletion")
	viper.BindPFlag("abortWaitTimeMinutes", deleteStacksCmd.Flags().Lookup("abortWaitTimeMinutes"))

	deleteStacksCmd.Flags().String("notificationWebhookURL", "", "Send status alerts to Slack channel")
	viper.BindPFlag("notificationWebhookURL", deleteStacksCmd.Flags().Lookup("notificationWebhookURL"))

	deleteStacksCmd.Flags().String("dryRun", "true", "[Safety Check] To delete stacks, it needs to be explicitely set to false")
	viper.BindPFlag("dryRun", deleteStacksCmd.Flags().Lookup("dryRun"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteStacksCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteStacksCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
