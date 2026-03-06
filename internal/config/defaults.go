package config

import (
	_ "embed"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

//go:embed default_config.toml
var defaultConfigTOML []byte

var (
	configSections = []string{"logger", "http", "performance", "deception", "ssh"}
	defNoColor     = false
)

var defOpts = map[string]interface{}{
	"logger": map[string]interface{}{
		"debug":               true,
		"trace":               false,
		"nocolor":             defNoColor,
		"use_date_filename":   true,
		"docker_logging":      false,
		"console_time_format": time.Kitchen,
	},
	"http": map[string]interface{}{
		"use_unix_socket":         false,
		"unix_socket_path":        "/var/run/hellpot",
		"unix_socket_permissions": "0666",
		"bind_addr":               "127.0.0.1",
		"bind_port":               "8080",
		"real_ip_header":          "X-Real-IP",

		"router": map[string]interface{}{
			"catchall":   false,
			"makerobots": true,
			"paths": []string{
				"wp-login.php",
				"wp-login",
			},
		},
		"uagent_string_blacklist": []string{
			"Cloudflare-Traffic-Manager",
		},
	},
	"performance": map[string]interface{}{
		"restrict_concurrency": false,
		"max_workers":          256,
	},
	"deception": map[string]interface{}{
		"server_name": "nginx",
	},
}

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

func setDefaults() {
	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == "windows" {
		defNoColor = true
	}
	for _, def := range configSections {
		section, ok := defOpts[def]
		if !ok {
			continue
		}
		sectionMap, ok := section.(map[string]interface{})
		if !ok {
			continue
		}
		for key, val := range sectionMap {
			if _, ok := val.(map[string]interface{}); !ok {
				if err := snek.Set(def+"."+key, val); err != nil {
					println(err.Error())
					os.Exit(1)
				}
				continue
			}
			for k, v := range val.(map[string]interface{}) {
				if err := snek.Set(def+"."+key+"."+k, v); err != nil {
					println(err.Error())
					os.Exit(1)
				}
			}
			continue
		}
	}

	if GenConfig {
		gen("")
	}
}
