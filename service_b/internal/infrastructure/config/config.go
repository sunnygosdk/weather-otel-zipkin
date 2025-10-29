package config

import "os"

type Conf struct {
	URLZipKin     string `mapstructure:"URL_ZIPKIN"`
	WeatherAPIKey string `mapstructure:"WEATHER_API_KEY"`
}

func LoadConfig() *Conf {
	return &Conf{
		URLZipKin:     os.Getenv("URL_ZIPKIN"),
		WeatherAPIKey: os.Getenv("WEATHER_API_KEY"),
	}
}
