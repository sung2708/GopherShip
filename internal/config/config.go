package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Ingester struct {
		BufferSize int    `yaml:"buffer_size"`
		Addr       string `yaml:"addr"`
		TLS        struct {
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
			CAFile   string `yaml:"ca_file"`
		} `yaml:"tls"`
	} `yaml:"ingester"`
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

	return cfg, nil
}

func Default() *Config {
	cfg := &Config{}
	cfg.Ingester.BufferSize = 8192
	cfg.Ingester.Addr = ":4317"
	return cfg
}
