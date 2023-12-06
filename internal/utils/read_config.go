package utils

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func MustReadConfig(filename string, v interface{}) {
	rawData, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("cannot read config from ", filename, ": ", err)
	}

	err = yaml.Unmarshal(rawData, v)
	if err != nil {
		log.Fatal("cannot unmarshal yaml: ", err)
	}
}
