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

// Package cmd provides interface to register and define actions for all cli commands
package cmd

import (
	"fmt"

	"github.com/gookit/color"
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
	Example: "cfn-teardown deleteStacks --STACK_PATTERN='^qa-' --AWS_PROFILE=staging --AWS_REGION=us-east-1",
	Args: func(cmd *cobra.Command, args []string) error {
		// validate your arguments here
		return validateConfigs(config)
	},
	Run: func(cmd *cobra.Command, args []string) {
		color.Red.Println("Executing command: deleteStacks")
		fmt.Println()
		if config.DryRun != "false" {
			fmt.Println("Running in dry run mode. Set dry run to 'false' to actually delete stacks.")
		}

		utils.InitiateTearDown(config)
	},
}

func init() {
	rootCmd.AddCommand(deleteStacksCmd)

	deleteStacksCmd.Flags().Int("STACK_WAIT_TIME_SECONDS", 30, "Seconds to wait after delete requests are submitted to CFN")
	viper.BindPFlag("STACK_WAIT_TIME_SECONDS", deleteStacksCmd.Flags().Lookup("STACK_WAIT_TIME_SECONDS"))

	deleteStacksCmd.Flags().String("TARGET_ACCOUNT_ID", "", "[Safety Check] Confirmes that account id from aws session and intented target aws account are the same")
	viper.BindPFlag("TARGET_ACCOUNT_ID", deleteStacksCmd.Flags().Lookup("TARGET_ACCOUNT_ID"))

	deleteStacksCmd.Flags().Int("MAX_DELETE_RETRY_COUNT", 5, "Max stack delete attempts")
	viper.BindPFlag("MAX_DELETE_RETRY_COUNT", deleteStacksCmd.Flags().Lookup("MAX_DELETE_RETRY_COUNT"))

	deleteStacksCmd.Flags().Int("ABORT_WAIT_TIME_MINUTES", 10, "[Safety Check] Minutes to wait before initiating deletion")
	viper.BindPFlag("ABORT_WAIT_TIME_MINUTES", deleteStacksCmd.Flags().Lookup("ABORT_WAIT_TIME_MINUTES"))

	deleteStacksCmd.Flags().String("SLACK_WEBHOOK_URL", "", "Send status alerts to Slack channel")
	viper.BindPFlag("SLACK_WEBHOOK_URL", deleteStacksCmd.Flags().Lookup("SLACK_WEBHOOK_URL"))

	deleteStacksCmd.Flags().String("DRY_RUN", "true", "[Safety Check] To delete stacks, it needs to be explicitely set to false")
	viper.BindPFlag("DRY_RUN", deleteStacksCmd.Flags().Lookup("DRY_RUN"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteStacksCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteStacksCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
