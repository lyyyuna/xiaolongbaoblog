package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Title       string `yaml:"title"`
	SubTitle    string `yaml:"subtitle"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	Email       string `yaml:"email"`

	Url  string `yaml:"url"`
	Root string `yaml:"root"`

	SourceDir string `yaml:"source_dir"`
	PostDir   string `yaml:"post_dir"`
	AboutDir  string `yaml:"about_dir"`

	OutputDir string `yaml:"output_dir"`
	SeriesDir string `yaml:"series_dir"`

	Deploys   []Deploy `yaml:"deploy"`
	Analytics string   `yaml:"analytics"`
}

type Deploy struct {
	Type       string `yaml:"type"`
	Repository string `yaml:"repository"`
	Branch     string `yaml:"branch"`
}

func NewConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("fail to read config file: %v", err)
	}

	c := Config{}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		log.Fatalf("fail to parse config file: %v", err)
	}

	return &c
}
