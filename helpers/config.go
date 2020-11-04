package helpers

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
)

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DBNum    int    `yaml:"dbNum"`
}

type LocalStorage struct {
	LocalPath string `yaml:"localpath"`
}

type Config struct {
	Redis        RedisConfig    `yaml:"redis"`
	StoragePrefix      LocalStorage `yaml:"storage_prefix"`
	UploadSlotTTL string `yaml:"upload_slot_ttl"`
}

func (c *Config) UploadSlotTTLDuration() (time.Duration, error) {
	return time.ParseDuration(c.UploadSlotTTL)
}

func ReadConfig(configFile string) (*Config, error) {
	configBytes, readErr := ioutil.ReadFile(configFile)
	if readErr != nil {
		log.Printf("Could not read config from '%s': %s\n", configFile, readErr)
		return nil, readErr
	}

	var conf Config

	err := yaml.Unmarshal(configBytes, &conf)
	if err != nil {
		log.Printf("Could not understand config from '%s': %s\n", configFile, err)
		return nil, err
	}
	return &conf, nil
}