package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactively add a new recording job",
	Long:  `Prompts for the necessary details to create a new recording job and adds it to the configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(cmd.InOrStdin())

		fmt.Print("Enter stream URL (e.g., http://radio.example.com/stream.pls): ")
		streamURL, _ := reader.ReadString('\n')

		fmt.Print("Enter start time (e.g., Wed 6:00pm): ")
		startTime, _ := reader.ReadString('\n')
        startTime = strings.TrimSpace(startTime)
        parsedStartTime, err := parseBestEffortTime(startTime)
        if err != nil {
            return fmt.Errorf("invalid start time: %w", err)
        }
        startTime = parsedStartTime.Format("Mon 3:04pm")

		fmt.Print("Enter duration (e.g., 1h30m): ")
		duration, _ := reader.ReadString('\n')

		fmt.Print("Enter output directory (e.g., /home/user/recordings): ")
		outputDir, _ := reader.ReadString('\n')

		fmt.Print("Recurring weekly? (y/n): ")
		recurringStr, _ := reader.ReadString('\n')
		recurring := strings.TrimSpace(strings.ToLower(recurringStr)) == "y"

		newJob := Job{
			StreamURL: strings.TrimSpace(streamURL),
			StartTime: startTime,
			Duration:  strings.TrimSpace(duration),
			OutputDir: strings.TrimSpace(outputDir),
			Recurring: recurring,
		}

		configDir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("could not get user config directory: %w", err)
		}
		configPath := filepath.Join(configDir, "overheard", "config.yaml")

		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return fmt.Errorf("could not create config directory: %w", err)
		}

		var jobs []Job
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("could not read config file: %w", err)
			}
			if err := yaml.Unmarshal(data, &jobs); err != nil {
				var singleJob Job
				if err := yaml.Unmarshal(data, &singleJob); err != nil {
					return fmt.Errorf("could not unmarshal config file: %w", err)
				}
				jobs = []Job{singleJob}
			}
		}

		jobs = append(jobs, newJob)

		data, err := yaml.Marshal(&jobs)
		if err != nil {
			return fmt.Errorf("could not marshal jobs to YAML: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("could not write config file: %w", err)
		}

		fmt.Printf("\nSuccessfully added new job to %s\n", configPath)

		return nil
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
}