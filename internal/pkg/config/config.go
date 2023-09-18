package config

import (
	"github.com/Vaayne/aienvoy/pkg/config"
)

var globalConfig = &Config{}

func init() {
	config.Load(globalConfig)
}

func GetConfig() *Config {
	return globalConfig
}
