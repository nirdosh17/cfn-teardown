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

	"github.com/spf13/cobra"
)

// Version is the current CLI version.
// This is overwrriten by semantic version tag while building binaries.
var Version = "development"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows cli command version",
	Long:  "Shows cli command version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version: ", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
