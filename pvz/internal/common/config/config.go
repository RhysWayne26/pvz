package config

import "flag"

// Config contains application configuration parameters
type Config struct {
	Path string
}

// Load parses command-line flags and returns application configuration
func Load() *Config {
	Path := flag.String("storage", "./storage.json", "Path to JSON storage file")
	flag.Parse()
	return &Config{
		Path: *Path,
	}
}
