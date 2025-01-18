package client

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

var ErrAbortedByUser = errors.New("operation aborted by user")

//go:embed default_config.yaml
var defaultConfig []byte

type Config struct {
	// The root directory of the config file. This field is not in the config file, but is set by LoadConfig.
	HomeDir string       `yaml:"-"`
	Server  ServerConfig `yaml:"server"`
}

type ServerConfig struct {
	URL string `yaml:"url"`
}

func LoadConfig(ctx context.Context) (*Config, error) {
	args := getArgs()
	configPath := path.Join(args.homeDir, "config.yaml")

	_, err := os.Stat(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("os.Stat: %w", err)
		} else {
			err = writeDefaultConfigFile(args.homeDir)
			if err != nil {
				return nil, fmt.Errorf("writeDefaultConfigFile: %w", err)
			}
		}
	}

	cfg, err := readConfigFile(args.homeDir)
	if err != nil {
		return nil, fmt.Errorf("readConfigFile: %w", err)
	}
	return cfg, nil
}

func writeDefaultConfigFile(homeDir string) error {
	configPath := path.Join(homeDir, "config.yaml")
	fmt.Printf("Config file not found, will create a new one at %q\nConfirm? (y/n) ", configPath)
	var choice string
	fmt.Scanln(&choice)
	if choice != "y" {
		return ErrAbortedByUser
	}
	err := os.MkdirAll(homeDir, 0700)
	if err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}
	err = os.WriteFile(configPath, defaultConfig, 0600)
	if err != nil {
		return fmt.Errorf("os.WriteFile: %w", err)
	}
	fmt.Printf("Config file created at %q\n", configPath)
	return nil
}

func readConfigFile(homeDir string) (*Config, error) {
	configPath := path.Join(homeDir, "config.yaml")
	cfg := new(Config)
	cfg.HomeDir = homeDir
	// Read the config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %w", err)
	}
	// Unmarshal the config file
	err = yaml.Unmarshal(content, cfg)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}
	return cfg, nil
}

type args struct {
	homeDir string
}

// getArgs parses and returns the command line arguments and/or relevant env vars
func getArgs() args {
	var homeDir string
	flag.StringVar(&homeDir, "home", "", "path to the talk home directory (defaults to $TALK_HOME, then $HOME/.config/talk)")
	flag.Parse()
	if homeDir == "" {
		homeDir = os.Getenv("TALK_HOME")
	}
	if homeDir == "" {
		homeDir = path.Join(os.Getenv("HOME"), ".config", "talk")
	}
	// Expand any user-supplied env vars
	homeDir = os.ExpandEnv(homeDir)
	return args{
		homeDir: homeDir,
	}
}
