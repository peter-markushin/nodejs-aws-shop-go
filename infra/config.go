package main

import "github.com/tkanos/gonfig"

type Configuration struct {
	DbName string `env:"DB_NAME"`
	DbUser string `env:"DB_USER"`
}

func GetConfig() Configuration {
	configuration := Configuration{}

	gonfig.GetConf("", &configuration)

	return configuration
}
