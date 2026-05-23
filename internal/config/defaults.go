package config

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed default_config.toml
var defaultConfigTOML []byte

func gen(path string) {
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			println("ERROR: Cannot determine current directory: " + err.Error())
			os.Exit(1)
		}
		path = filepath.Join(cwd, "config.toml")
	}

	if err := os.WriteFile(path, defaultConfigTOML, 0o600); err != nil {
		println("ERROR: Cannot write config file: " + err.Error())
		println("")
		println("If installing system-wide, see README.md for proper setup.")
		os.Exit(1)
	}

	pathAbs, absErr := filepath.Abs(path)
	if absErr == nil && pathAbs != "" {
		path = pathAbs
	}

	println("Default config written to " + path)
	os.Exit(0)
}
