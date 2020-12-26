package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Forward []ForwardT `yaml:"forward"`
}

type ForwardT struct {
	Name string   `yaml:"name"`
	From int      `yaml:"from"`
	To   []string `yaml:"to"`
	User []int    `yaml:"user"`
}

func (c *Config) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}

func getConfig(filename string) (*Config, error) {
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = config.Parse(source)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
