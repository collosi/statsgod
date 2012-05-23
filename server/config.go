package server

import (
	"encoding/json"
	"os"
)

type StatConfig struct {
	Name      string `json:"name"`
	Key       string `json:"key"`
	IsCounter bool   `json:"isCounter"`
	Capacity  int    `json:"capacity,omitempty"`
}

type ConfigFile struct {
	Stats []StatConfig `json:"stats"`
}

func ReadConfig(path string) (*ConfigFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	var cf ConfigFile
	dec := json.NewDecoder(f)
	err = dec.Decode(&cf)
	if err != nil {
		return nil, err
	}
	return &cf, nil
}

func (c *ConfigFile) Write(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	return enc.Encode(c)

}
