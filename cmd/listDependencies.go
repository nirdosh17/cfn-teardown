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
)

// listDependenciesCmd represents the listDependencies command
var listDependenciesCmd = &cobra.Command{
	Use:   "listDependencies",
	Short: "Scan stacks and list stacks and their dependencies",
	Long: `Scan stacks and list stacks and their dependencies.
A stack pattern must be provided otherwise it will scan all stacks in the region.
Example:
If your stacks to be deleted follow this naming convention: qa-{{component name}}
Supply stack pattern as: 'qa-'
	`,
	Example: "cfn-teardown listDependencies --STACK_PATTERN='qa-' --AWS_PROFILE='staging' --AWS_REGION=us-east-1",

	Args: func(cmd *cobra.Command, args []string) error {
		// validate your arguments here
		return validateConfigs(config)
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		color.Green.Println("Executing command: listDependencies")
		fmt.Println()
		// for safety
		config.DryRun = "true"
		fmt.Println("Running in dry run mode...")

		utils.InitiateTearDown(config)
	},
}

func init() {
	rootCmd.AddCommand(listDependenciesCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listDependenciesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listDependenciesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
