package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configFileMutex = &sync.Mutex{}

func init() {
	RootCmd.AddCommand(recordCmd)
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Starts the recording process based on the config file",
	Long:  `Reads the configuration file, waits for the scheduled time, and then records the audio stream.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir, _ := os.UserConfigDir()
		configPath := filepath.Join(configDir, "overheard", "config.yaml")

		yamlFile, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file at %s: %w", configPath, err)
		}

		var jobs []Job
		err = yaml.Unmarshal(yamlFile, &jobs)
		if err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}

		var wg sync.WaitGroup
		for _, job := range jobs {
			wg.Add(1)
			go func(j Job) {
				defer wg.Done()
				if err := recordJob(j); err != nil {
					fmt.Fprintf(os.Stderr, "Error recording job %s: %v\n", j.StreamURL, err)
				}
			}(job)
		}
		wg.Wait()

		return nil
	},
}

func recordJob(job Job) error {
	// Parse duration and start time
	duration, err := time.ParseDuration(job.Duration)
	if err != nil {
		return fmt.Errorf("invalid duration in config: %w", err)
	}

	startTime, err := parseBestEffortTime(job.StartTime)
	if err != nil {
		return fmt.Errorf("invalid start_time in config: %w", err)
	}

	fmt.Printf("Job loaded. Waiting until %s to record.\n", startTime.Format(time.RFC1123))
	time.Sleep(time.Until(startTime))

	// Start Recording
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	streamURL := job.StreamURL
	if strings.HasSuffix(streamURL, ".m3u") || strings.HasSuffix(streamURL, ".pls") {
		actualStreamURL, err := getStreamURLFromPlaylist(streamURL)
		if err != nil {
			return fmt.Errorf("could not get stream URL from playlist: %w", err)
		}
		streamURL = actualStreamURL
	}

	req, err := http.NewRequestWithContext(ctx, "GET", streamURL, nil)
	if err != nil {
		return fmt.Errorf("could not create http request: %w", err)
	}

	stream, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not connect to stream: %w", err)
	}
	defer stream.Body.Close()

	if stream.StatusCode != http.StatusOK {
		return fmt.Errorf("stream returned non-200 status code: %s", stream.Status)
	}

	// For now, we save as .mp3, as FLAC encoding is blocked.
	// Ensure the output directory exists.
	if err := os.MkdirAll(job.OutputDir, 0755); err != nil {
		return fmt.Errorf("could not create output directory: %w", err)
	}

	outputFilePath := filepath.Join(job.OutputDir, fmt.Sprintf("recording-%d.mp3", time.Now().Unix()))
	f, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("could not create output file: %w", err)
	}
	defer f.Close()

	fmt.Printf("Recording to %s...\n", outputFilePath)

	bytesWritten, err := io.Copy(f, stream.Body)
	if err != nil {
		fmt.Println("\nRecording stopped due to timeout or stream error.")
	}

	if bytesWritten == 0 {
		fmt.Println("Warning: 0 bytes were written.")
	} else {
		fmt.Printf("\nFinished recording. Wrote %d bytes.\n", bytesWritten)
	}

	if job.Recurring {
		configFileMutex.Lock()
		defer configFileMutex.Unlock()

		configDir, _ := os.UserConfigDir()
		configPath := filepath.Join(configDir, "overheard", "config.yaml")

		yamlFile, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file at %s: %w", configPath, err)
		}

		var jobs []Job
		err = yaml.Unmarshal(yamlFile, &jobs)
		if err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}

		for i, j := range jobs {
			if j.StreamURL == job.StreamURL && j.StartTime == job.StartTime {
				parsedStartTime, err := parseBestEffortTime(j.StartTime)
				if err != nil {
					return fmt.Errorf("invalid start time: %w", err)
				}
				jobs[i].StartTime = parsedStartTime.Add(7 * 24 * time.Hour).Format("Mon 3:04pm")
				break
			}
		}

		data, err := yaml.Marshal(&jobs)
		if err != nil {
			return fmt.Errorf("could not marshal jobs to YAML: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("could not write config file: %w", err)
		}
	}

	return nil
}
