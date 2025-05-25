package config

import "flag"

type Config struct {
	Path string
}

func Load() *Config {
	Path := flag.String("storage", "./storage.json", "Path to JSON storage file")
	flag.Parse()
	return &Config{
		Path: *Path,
	}
}
