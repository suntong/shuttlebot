package main

import (
	"io/ioutil"

	tb "gopkg.in/tucnak/telebot.v2"
	"gopkg.in/yaml.v2"
)

////////////////////////////////////////////////////////////////////////////
// Constant and data type/structure definitions

type envConfig struct {
	TelegramToken string `env:"SHUTTLEBOT_TOKEN,required"`
	ConfigFile    string `env:"SHUTTLEBOT_CFG"`
	LogLevel      string `env:"SHUTTLEBOT_LOG"`
}

// https://yaml2go.prasadg.dev/
type Config struct {
	Fetchable  bool
	Fetch      Fetch      `yaml:"fetch"`
	Forward    []ForwardT `yaml:"forward"`
	FromGroups []int
}

// Fetch
type Fetch struct {
	Users      []int    `yaml:"users"`
	Command    string   `yaml:"command"`
	Downloader string   `yaml:"downloader"`
	Vformat    []string `yaml:"vformat"`
	Folder     string   `yaml:"folder"`
}

type ForwardT struct {
	Name string  `yaml:"name"`
	From int     `yaml:"from"`
	To   []int64 `yaml:"to"`
	User []int   `yaml:"user"`
	Chat []*tb.Chat
}

////////////////////////////////////////////////////////////////////////////
// Global variables definitions

var (
	c   envConfig
	cfg *Config
)

////////////////////////////////////////////////////////////////////////////
// Function definitions

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
