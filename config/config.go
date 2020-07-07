package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var Cfg = &config{}

type config struct {
	SAZ2go struct {
		Cookie struct {
			Remove []string
		}
		Request struct {
		}
	}
	SAZParse struct{}
}

func programPath() string {
	return filepath.Dir(os.Args[0])
}

func LoadConfig() error {
	data, err := ioutil.ReadFile(programPath() + "/xtool.yaml")
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		return err
	}

	return nil
}
