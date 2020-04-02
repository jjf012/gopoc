package lib

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Poc struct {
	Name   string            `yaml:"name"`
	Set    map[string]string `yaml:"set"`
	Rules  []Rules           `yaml:"rules"`
	Detail Detail            `yaml:"detail"`
}

type Rules struct {
	Method          string            `yaml:"method"`
	Path            string            `yaml:"path"`
	Headers         map[string]string `yaml:"headers"`
	Body            string            `yaml:"body"`
	Search          string            `yaml:"search"`
	FollowRedirects bool              `yaml:"follow_redirects"`
	Expression      string            `yaml:"expression"`
}

type Detail struct {
	Author string   `yaml:"author"`
	Links  []string `yaml:"links"`
}

func LoadPocFromFile(filepath string) (*Poc, error) {
	p := &Poc{}
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, p)
	if err != nil {
		return nil, err
	}
	return p, err
}
