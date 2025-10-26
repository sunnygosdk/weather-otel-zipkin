package config

import "os"

type Conf struct {
	URLZipKin string `mapstructure:"URL_ZIPKIN"`
}

func LoadConfig() *Conf {
	return &Conf{
		URLZipKin: os.Getenv("URL_ZIPKIN"),
	}
}
