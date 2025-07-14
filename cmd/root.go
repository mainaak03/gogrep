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
	"bufio"
	"bytes"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"sync"
	"github.com/BurntSushi/rure-go"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gogrep",
	Short: "Regex pattern matching implemented in Go",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		f, err := os.Create("cpu.prof")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		// Start CPU profiling
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		
		pattern, err_pattern := cmd.Flags().GetString("pattern")
		if (err_pattern != nil) {
			log.Fatalf("Error while parsing pattern: %v", err_pattern.Error())
		}
		filename, err_filename := cmd.Flags().GetString("filename")
		if (err_filename != nil) {
			log.Fatalf("Error while parsing filename: %v", err_filename.Error())
		}
		enableLineNumber, err_enableLineNumber := cmd.Flags().GetBool("line-number")
		if (err_enableLineNumber != nil) {
			log.Fatalf("Error: %v", err_filename.Error())
		}
		match(&pattern, &filename, &enableLineNumber)
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
	rootCmd.Flags().StringP("pattern", "p", "", "Regex expression to be matched")
	rootCmd.Flags().StringP("filename", "f", "", "Filename to search")
	rootCmd.Flags().BoolP("ignore-case", "i", false, "Enable case insensitive matching")
	rootCmd.Flags().BoolP("line-number", "n", false, "Prefix matching lines with line numbers")
	rootCmd.MarkFlagRequired("pattern")
	rootCmd.MarkFlagRequired("filename")
}

func match(pattern *string, filename *string, enableLineNumber *bool) {
	
	// If the pattern is just a string literal, we will skip regex matching
	isLiteral := !rure.MustCompile(`[.*+?^$()\[\]{}|\\]`).IsMatch(*pattern)

	var re *rure.Regex
	if (!isLiteral) {
		re = rure.MustCompile(*pattern)
	}
	
	patternBytes := []byte(*pattern)
	
	files, err := filepath.Glob(*filename)
	if (err != nil || len(files) == 0) {
		log.Fatalf("Error while listing files: %v", err.Error())
	}

	var wg sync.WaitGroup
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for _, file := range(files) {
		wg.Add(1)
		go func (filename string) {
			defer wg.Done()
			matchText(&filename, &isLiteral, enableLineNumber, patternBytes, re, writer)
		}(file)
	}
	wg.Wait()
}

func matchText(filename *string, isLiteral *bool, enableLineNumber *bool, patternBytes []byte, re *rure.Regex, writer *bufio.Writer) {
	file, err := os.Open(*filename)
	if (err != nil) {
		log.Fatalf("Error while opening file: %v", err.Error())
	}
	defer file.Close()
	
	var bufferedScanner = bufio.NewScanner(file)
	const bufSize = 1024 * 1024
	buf := make([]byte, bufSize)
	bufferedScanner.Buffer(buf, bufSize)

	lineNumber := 1
	for (bufferedScanner.Scan()) {
		var matched = false
		if (*isLiteral) {
			if (bytes.Contains(bufferedScanner.Bytes(), patternBytes)) {
				matched = true
			}
		} else {
			if (re.IsMatchBytes(bufferedScanner.Bytes())) {
				matched = true
			}
		}
		if (matched) {
			if (*enableLineNumber) {
				// fmt.Fprintf(os.Stdout, "[%v]-[%v]: %s\n", lineNumber, *filename, bufferedScanner.Bytes())
				writer.WriteByte('[')
				writer.WriteString(strconv.Itoa(lineNumber))
				writer.WriteString("]-[")
				writer.WriteString(*filename)
				writer.WriteString("]: ")
				writer.Write(bufferedScanner.Bytes())
				writer.WriteByte('\n')
			} else {
				// fmt.Fprintf(os.Stdout, "[%v]: %s\n", *filename, bufferedScanner.Bytes())
				writer.WriteByte('[')
				writer.WriteString(*filename)
				writer.WriteString("]: ")
				writer.Write(bufferedScanner.Bytes())
				writer.WriteByte('\n')
			}
		}
		lineNumber++;
	}	
}