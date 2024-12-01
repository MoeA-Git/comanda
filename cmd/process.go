package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/kris-hansen/comanda/utils/config"
	"github.com/kris-hansen/comanda/utils/processor"
)

var processCmd = &cobra.Command{
	Use:   "process [files...]",
	Short: "Process YAML DSL configuration files",
	Long:  `Process one or more DSL configuration files and execute the specified actions.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get environment file path
		envPath := config.GetEnvPath()

		// Load environment configuration
		if verbose {
			fmt.Printf("[DEBUG] Loading environment configuration from %s\n", envPath)
		}

		envConfig, err := config.LoadEnvConfigWithPassword(envPath)
		if err != nil {
			log.Fatalf("Error loading environment configuration: %v", err)
		}

		if verbose {
			fmt.Println("[DEBUG] Environment configuration loaded successfully")
		}

		// Check if there's data on STDIN
		stat, _ := os.Stdin.Stat()
		var stdinData string
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Read from STDIN
			reader := bufio.NewReader(os.Stdin)
			var builder strings.Builder
			for {
				input, err := reader.ReadString('\n')
				if err != nil && err != io.EOF {
					log.Fatalf("Error reading from STDIN: %v", err)
				}
				builder.WriteString(input)
				if err == io.EOF {
					break
				}
			}
			stdinData = builder.String()
		}

		for _, file := range args {
			fmt.Printf("\nProcessing DSL file: %s\n", file)

			// Read YAML file
			if verbose {
				fmt.Printf("[DEBUG] Reading YAML file: %s\n", file)
			}
			yamlFile, err := os.ReadFile(file)
			if err != nil {
				log.Printf("Error reading YAML file %s: %v\n", file, err)
				continue
			}

			// First unmarshal into a map to preserve step names
			var rawConfig map[string]processor.StepConfig
			err = yaml.Unmarshal(yamlFile, &rawConfig)
			if err != nil {
				log.Printf("Error parsing YAML file %s: %v\n", file, err)
				continue
			}

			// Convert map to ordered Steps slice
			var dslConfig processor.DSLConfig
			for name, config := range rawConfig {
				dslConfig.Steps = append(dslConfig.Steps, processor.Step{
					Name:   name,
					Config: config,
				})
			}

			// Create processor
			if verbose {
				fmt.Printf("[DEBUG] Creating processor for %s\n", file)
			}
			proc := processor.NewProcessor(&dslConfig, envConfig, verbose)

			// If we have STDIN data, set it as initial output
			if stdinData != "" {
				proc.SetLastOutput(stdinData)
			}

			// Print configuration summary before processing
			fmt.Println("\nConfiguration:")
			for _, step := range dslConfig.Steps {
				fmt.Printf("\nStep: %s\n", step.Name)
				inputs := proc.NormalizeStringSlice(step.Config.Input)
				if len(inputs) > 0 && inputs[0] != "NA" {
					fmt.Printf("- Input: %v\n", inputs)
				}
				fmt.Printf("- Model: %v\n", proc.NormalizeStringSlice(step.Config.Model))
				fmt.Printf("- Action: %v\n", proc.NormalizeStringSlice(step.Config.Action))
				fmt.Printf("- Output: %v\n", proc.NormalizeStringSlice(step.Config.Output))
				nextActions := proc.NormalizeStringSlice(step.Config.NextAction)
				if len(nextActions) > 0 {
					fmt.Printf("- Next Action: %v\n", nextActions)
				}
			}
			fmt.Println()

			// Run processor
			if err := proc.Process(); err != nil {
				log.Printf("Error processing DSL file %s: %v\n", file, err)
				continue
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(processCmd)
}
