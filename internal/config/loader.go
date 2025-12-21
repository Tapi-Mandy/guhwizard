// FILE: internal/config/loader.go
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Item struct {
	Name        string `yaml:"name"`
	Description string `yaml:"desc"`
	Selected    bool   `yaml:"-"`
}

type Step struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
	Type  string `yaml:"type"`
	Items []Item `yaml:"items"`
}

type Script struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

type DotfileItem struct {
	Src  string `yaml:"src"`
	Dest string `yaml:"dest"`
}

type DotfilesConfig struct {
	Repo      string        `yaml:"repo"`
	TargetDir string        `yaml:"target_dir"`
	Items     []DotfileItem `yaml:"items"`
}

type Config struct {
	Settings struct {
		AURHelper       string         `yaml:"aur_helper"`
		BasePackages    []string       `yaml:"base_packages"`
		ExternalScripts []Script       `yaml:"external_scripts"`
		Dotfiles        DotfilesConfig `yaml:"dotfiles"`
	} `yaml:"settings"`
	Steps []Step `yaml:"steps"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
