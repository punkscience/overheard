package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display configuration information",
	Long:  `Displays the path to the configuration file and validates its contents.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("could not get user config directory: %w", err)
		}
		configPath := filepath.Join(configDir, "overheard", "config.yaml")

		fmt.Printf("Configuration file path: %s\n", configPath)

		v := viper.New()
		v.SetConfigFile(configPath)

		if err := v.ReadInConfig(); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Configuration file not found.")
				fmt.Printf("You can create a configuration file at '%s'\n", configPath)
				// check if the directory exists
				if _, err := os.Stat(filepath.Dir(configPath)); os.IsNotExist(err) {
					fmt.Printf("Creating directory '%s'\n", filepath.Dir(configPath))
					err = os.MkdirAll(filepath.Dir(configPath), 0755)
					if err != nil {
						return fmt.Errorf("could not create directory: %w", err)
					}
				}
				return nil
			} else {
				return fmt.Errorf("could not read configuration file: %w", err)
			}
		}

		fmt.Println("Configuration file found and is valid YAML.")
		return nil
	},
}
