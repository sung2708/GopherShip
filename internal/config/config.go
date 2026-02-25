package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Ingester struct {
		BufferSize int    `yaml:"buffer_size,omitempty"`
		Addr       string `yaml:"addr,omitempty"`
		TLS        struct {
			CertFile string `yaml:"cert_file,omitempty"`
			KeyFile  string `yaml:"key_file,omitempty"`
			CAFile   string `yaml:"ca_file,omitempty"`
		} `yaml:"tls,omitempty"`
	} `yaml:"ingester,omitempty"`
	Monitoring struct {
		MaxRAM          uint64  `yaml:"max_ram,omitempty"`
		YellowThreshold float64 `yaml:"yellow_threshold,omitempty"`
		RedThreshold    float64 `yaml:"red_threshold,omitempty"`
		IngesterBudget  uint64  `yaml:"ingester_budget,omitempty"`
		VaultBudget     uint64  `yaml:"vault_budget,omitempty"`
	} `yaml:"monitoring,omitempty"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = os.Getenv("GS_CONFIG")
		if path == "" {
			path = "config.yaml"
		}
	}

	cfg := Default()
	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
			return nil, err
		}
	} else if !(os.IsNotExist(err) && path == "config.yaml") {
		return nil, err
	}

	// Environment variable overrides (AC3 / Pass 3 Review)
	if env := os.Getenv("GS_INGEST_CERT"); env != "" {
		cfg.Ingester.TLS.CertFile = env
	}
	if env := os.Getenv("GS_INGEST_KEY"); env != "" {
		cfg.Ingester.TLS.KeyFile = env
	}
	if env := os.Getenv("GS_INGEST_CA"); env != "" {
		cfg.Ingester.TLS.CAFile = env
	}
	if env := os.Getenv("GS_INGEST_ADDR"); env != "" {
		cfg.Ingester.Addr = env
	}

	// Monitoring environment overrides (Pass 6 Review)
	if env := os.Getenv("GS_MONITOR_MAX_RAM"); env != "" {
		if v, err := parseUint64(env); err == nil {
			cfg.Monitoring.MaxRAM = v
		}
	}
	if env := os.Getenv("GS_MONITOR_INGEST_BUDGET"); env != "" {
		if v, err := parseUint64(env); err == nil {
			cfg.Monitoring.IngesterBudget = v
		}
	}
	if env := os.Getenv("GS_MONITOR_VAULT_BUDGET"); env != "" {
		if v, err := parseUint64(env); err == nil {
			cfg.Monitoring.VaultBudget = v
		}
	}

	return cfg, nil
}

func parseUint64(s string) (uint64, error) {
	var val uint64
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}

func Default() *Config {
	cfg := &Config{}
	cfg.Ingester.BufferSize = 8192
	cfg.Ingester.Addr = ":4317"
	cfg.Monitoring.MaxRAM = 1024 * 1024 * 1024 // 1GB
	cfg.Monitoring.YellowThreshold = 0.80
	cfg.Monitoring.RedThreshold = 0.95
	cfg.Monitoring.IngesterBudget = 256 * 1024 * 1024 // 256MB
	cfg.Monitoring.VaultBudget = 512 * 1024 * 1024    // 512MB
	return cfg
}
