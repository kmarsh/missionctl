// Package config finds the API key.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ErrNoKey means neither the environment nor a config file supplied a key.
var ErrNoKey = errors.New("no API key")

// Files are checked in order and the first that exists is used, as the Mission
// Control Mac app does; ~/.houston.toml is its legacy name.
var Files = []string{".mission_control.toml", ".houston.toml"}

// APIKey comes from $MISSIONCTL_API_KEY, or else api_key in the first config file found.
func APIKey() (string, error) {
	if key := os.Getenv("MISSIONCTL_API_KEY"); key != "" {
		return key, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNoKey, err)
	}
	for _, name := range Files {
		path := filepath.Join(home, name)
		var config struct {
			APIKey string `toml:"api_key"`
		}
		_, err := toml.DecodeFile(path, &config)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("%s: %w", path, err)
		}
		if config.APIKey == "" {
			return "", fmt.Errorf("%w: %s has no api_key", ErrNoKey, path)
		}
		return config.APIKey, nil
	}
	return "", ErrNoKey
}
