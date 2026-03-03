package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	OpinetKey          string
	NominatimUserAgent string
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "nearby-gas-prices", "config.toml"), nil
}

// Load reads a very small subset of TOML (key = "value") from the default config path.
//
// Supported keys:
// - opinet_key
// - nominatim_user_agent
func Load() (Config, string, error) {
	p, err := DefaultPath()
	if err != nil {
		return Config{}, "", err
	}

	f, err := os.Open(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, p, nil
		}
		return Config{}, p, err
	}
	defer f.Close()

	cfg := Config{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// strip inline comments: key = "value" # comment
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}

		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key := strings.TrimSpace(k)
		val := strings.TrimSpace(v)
		val = strings.Trim(val, "\"'")

		switch key {
		case "opinet_key":
			cfg.OpinetKey = val
		case "nominatim_user_agent":
			cfg.NominatimUserAgent = val
		}
	}
	if err := s.Err(); err != nil {
		return Config{}, p, fmt.Errorf("config read: %w", err)
	}
	return cfg, p, nil
}
