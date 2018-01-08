package config

import (
	"io/ioutil"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

type Resources struct {
	Resource map[string][]string `yaml:"Resource"`
}

func GetResourceConfig(config string) Resources {
	yamlFile, err := ioutil.ReadFile(config)
	resources := Resources{}
	if err != nil {
		glog.Errorf("Can't open resource config file: %v", err)

	}
	err = yaml.Unmarshal(yamlFile, &resources)
	if err != nil {
		glog.Fatalf("error: %v", err)
	}
	return resources
}
