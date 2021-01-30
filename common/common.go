package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/0x0bsod/goLogz"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type CustomContext struct {
	BackgroundCtx context.Context
	Logger        goLogz.GoLogz
	StackIn       *Stack
	StackOut      *Stack
	Config        *Config
}

type Config struct {
	TopicI  string   `yaml:"topicI"`
	TopicO  string   `yaml:"topicO"`
	Brokers []string `yaml:"brokers"`
	PrometheusPort int `yaml:"prometheusPort"`
}

func (c *Config) ParseConfig(path string) error {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading YAML file: %s\n", err))
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing YAML file: %s\n", err))
	}

	return nil
}
