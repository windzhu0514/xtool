package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var Cfg = &config{}

type Decrypt struct {
	AlgoName string `yaml:"algoName"`
	Key      string `yaml:"key"` // public key,private key or key
	IV       string `yaml:"iv"`
}

type config struct {
	SAZ struct {
		Head struct {
			Del []string
		}
		Body struct {
			Request struct {
				Decrypt Decrypt
			}
			Response struct {
				Decrypt Decrypt
			}
		}
	}
}

func programPath() string {
	return filepath.Dir(os.Args[0])
}

func LoadConfig() error {
	fmt.Println(programPath())
	data, err := ioutil.ReadFile(programPath() + "/xtool.yaml")
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		return err
	}

	return nil
}
