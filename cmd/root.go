/*
Copyright © 2025 Mainaak

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
	cgoutils "gogrep/cgo-utils"
	"os"
	"runtime/pprof"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gogrep",
	Short: "Regex pattern matching implemented in Go",
	Long:  ``,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		profile, _ := cmd.Flags().GetBool("profile")
		if profile {
			f, err := os.Create("cpu.prof")
			if err != nil {
				panic(err)
			}
			defer f.Close()
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		pattern := args[0]
		path := args[1]

		enableRecursive, _ := cmd.Flags().GetBool("recursive")
		// enableLineNumber, _ := cmd.Flags().GetBool("line-number")
		ignoreCase, _ := cmd.Flags().GetBool("ignore-case")

		if ignoreCase {
			// this is VERY hacky but I honestly don't want to model the logic in C 
			pattern = "(?i)" + pattern
		}
		
		cgoutils.WalkAndMatch(path, pattern, enableRecursive)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("ignore-case", "i", false, "Enable case insensitive matching")
	rootCmd.Flags().BoolP("line-number", "n", false, "Prefix matching lines with line numbers")
	rootCmd.Flags().BoolP("recursive", "r", false, "Search recursivley inside dirs")
	rootCmd.Flags().Bool("profile", false, "[debug] Profile go runtime")
}
