package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var Cfg = &config{}

type Decrypt struct {
	AlgoName string
	Key      string // public key,private key or key
	IV       string
}

type config struct {
	SAZ2go struct {
		Cookie struct {
			Remove []string
		}
		Request struct {
			Decrypt
		}
		Response struct {
			Decrypt
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
