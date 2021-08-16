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
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/nirdosh17/cfn-teardown/models"
)

// config vars
var (
	cfgFile string
)

var config models.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cfn-teardown",
	Short: "Delete cloudformation stacks with matching names",
	Long: `Finds and deletes stacks whose name matches with the given pattern.

	First, It prepares the list of stacks and their dependencies.
	Then, It recursively searches for importer stacks until stacks in leaf node has no importer stacks.
	Finally, the stack in the leaf nodes are deleted concurrently.
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		// validate your arguments here
		return nil
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func validateConfigs(config models.Config) (err error) {
	emptyFlags := []string{}

	if config.StackPattern == "" {
		emptyFlags = append(emptyFlags, "STACK_PATTERN")
	}

	if config.AWSProfile == "" {
		emptyFlags = append(emptyFlags, "AWS_PROFILE")
	}

	if config.AWSRegion == "" {
		emptyFlags = append(emptyFlags, "AWS_REGION")
	}

	if len(emptyFlags) > 0 {
		err = errors.New("required flag(s) " + strings.Join(emptyFlags, ", ") + " not set")
	}

	return
}

func init() {
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().String("STACK_PATTERN", "", "Pattern to match stack name e.g. 'staging-'")
	viper.BindPFlag("STACK_PATTERN", rootCmd.PersistentFlags().Lookup("STACK_PATTERN"))

	rootCmd.PersistentFlags().String("AWS_REGION", "", "AWS Region where the stacks are present")
	viper.BindPFlag("AWS_REGION", rootCmd.PersistentFlags().Lookup("AWS_REGION"))

	rootCmd.PersistentFlags().String("AWS_PROFILE", "", "AWS Profile")
	viper.BindPFlag("AWS_PROFILE", rootCmd.PersistentFlags().Lookup("AWS_PROFILE"))

	rootCmd.PersistentFlags().String("ROLE_ARN", "", "Assume this role to scan and delete stacks if provided")
	viper.BindPFlag("ROLE_ARN", rootCmd.PersistentFlags().Lookup("ROLE_ARN"))

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cfn-teardown.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cfn-teardown.yaml"
		viper.AddConfigPath(home)
		viper.SetConfigName(".cfn-teardown.yaml")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file: ", viper.ConfigFileUsed())
		err = viper.Unmarshal(&config)
		if err != nil {
			log.Fatalf("Error parsing config, %s", err)
		}
	} else {
		log.Fatalf("Error reading config file, %s", err)
	}
}
