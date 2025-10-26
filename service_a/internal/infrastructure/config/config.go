package config

import "os"

type Config struct {
	URLServiceB string `mapstructure:"URL_SERVICE_B"`
	URLZipKin   string `mapstructure:"URL_ZIPKIN"`
}

func SetConfig() *Config {
	return &Config{
		URLServiceB: os.Getenv("URL_SERVICE_B"),
		URLZipKin:   os.Getenv("URL_ZIPKIN"),
	}
}
