package cmd

// Job defines the structure for a recording job in the config file.
type Job struct {
	StreamURL string `yaml:"stream_url"`
	StartTime string `yaml:"start_time"`
	Duration  string `yaml:"duration"`
	OutputDir string `yaml:"output_dir"`
	Recurring bool   `yaml:"recurring,omitempty"`
}