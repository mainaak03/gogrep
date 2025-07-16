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
	"regexp"
	"runtime/pprof"
	"strconv"
	"sync"

	"github.com/BurntSushi/rure-go"

	"github.com/spf13/cobra"
)

type WorkerConfig struct {
	isLiteral           bool
	enableLineNumber    bool
	enableGoRegexEngine bool
	patternBytes        []byte
	rureRegex           *rure.Regex
	goRegex             *regexp.Regexp
}

type IoConfig struct {
	filenameChannel chan string
	outputChannel   chan []byte
}

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
		if err_pattern != nil {
			log.Fatalf("Error while parsing pattern: %v", err_pattern.Error())
		}
		filename, err_filename := cmd.Flags().GetString("filename")
		if err_filename != nil {
			log.Fatalf("Error while parsing filename: %v", err_filename.Error())
		}
		enableLineNumber, err_enableLineNumber := cmd.Flags().GetBool("line-number")
		if err_enableLineNumber != nil {
			log.Fatalf("Error: %v", err_filename.Error())
		}
		enableGoRegexEngine, _ := cmd.Flags().GetBool("go-regex")
		match(&pattern, &filename, &enableLineNumber, &enableGoRegexEngine)
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
	rootCmd.Flags().Bool("go-regex", false, "Use Go Regex Engine instead of Rust Regex Engine")
	rootCmd.MarkFlagRequired("pattern")
	rootCmd.MarkFlagRequired("filename")
}

func match(pattern *string, path *string, enableLineNumber *bool, enableGoRegexEngine *bool) {

	// If the pattern is just a string literal, we will skip regex matching
	isLiteral := !rure.MustCompile(`[.*+?^$()\[\]{}|\\]`).IsMatch(*pattern)

	var re *rure.Regex
	var reGo *regexp.Regexp

	if !isLiteral {
		if *enableGoRegexEngine {
			reGo = regexp.MustCompile(*pattern)
		} else {
			re = rure.MustCompile(*pattern)
		}
	}

	patternBytes := []byte(*pattern)

	files, err := filepath.Glob(*path)
	if err != nil || len(files) == 0 {
		log.Fatalf("Error while listing files: %v", err.Error())
	}

	var matchWg sync.WaitGroup
	var writerWg sync.WaitGroup
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	filenameChannel := make(chan string, 1024 * 1024)
	outputChannel := make(chan []byte, 4096 * 1024)

	workerConfig := &WorkerConfig{
		isLiteral:           isLiteral,
		enableGoRegexEngine: *enableGoRegexEngine,
		enableLineNumber:    *enableLineNumber,
		patternBytes:        patternBytes,
	}
	if *enableGoRegexEngine {
		workerConfig.goRegex = reGo
	} else {
		workerConfig.rureRegex = re
	}

	ioConfig := &IoConfig{
		filenameChannel: filenameChannel,
		outputChannel:   outputChannel,
	}

	writerWg.Add(1)
	go func (outputChannel chan []byte) {
		defer writerWg.Done()
		for line := range outputChannel {
			writer.Write(line)
		}
	}(outputChannel)

	// Setting num-workers to 8 for now, more doesnt seem to provide more performnce
	// as they are green threads and are scheduled on fixed number of OS threads
	for range 8 {
		matchWg.Add(1)
		go matchTextWorker(workerConfig, ioConfig, &matchWg)
	}

	for _, filename := range files {
		filenameChannel <- filename
	}
	close(filenameChannel)
	matchWg.Wait()
	close(outputChannel)
	writerWg.Wait()
}

func matchTextWorker(workerConfig *WorkerConfig, ioConfig *IoConfig, wg *sync.WaitGroup) {
	
	defer wg.Done()

	for filename := range ioConfig.filenameChannel {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("Error while opening file: %v", err.Error())
		}

		bufferedScanner := bufio.NewScanner(file)
		const bufSize = 1024 * 1024
		buf := make([]byte, bufSize)
		bufferedScanner.Buffer(buf, bufSize)

		lineNumber := 1
		for bufferedScanner.Scan() {
			var matched = false
			if workerConfig.isLiteral {
				matched = bytes.Contains(bufferedScanner.Bytes(), workerConfig.patternBytes)
			} else if workerConfig.enableGoRegexEngine {
				matched = workerConfig.goRegex.Match(bufferedScanner.Bytes())
			} else {
				matched = workerConfig.rureRegex.IsMatchBytes(bufferedScanner.Bytes())
			}

			if matched {
				// Creating a buffer because fmt.Sprintf is (kinda) slow
				buf := make([]byte, 0, len(filename)+len(bufferedScanner.Bytes())+32)
				buf = append(buf, '[')
				buf = append(buf, filename...)
				if workerConfig.enableLineNumber {
					buf = append(buf, "]-["...)
					buf = strconv.AppendInt(buf, int64(lineNumber), 10)
				}
				buf = append(buf, "]: "...)
				buf = append(buf, bufferedScanner.Bytes()...)
				buf = append(buf, '\n')
				ioConfig.outputChannel <- buf
			}
			lineNumber++
		}
		file.Close()
	}
}
