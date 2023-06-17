package config

import (
	"encoding/json"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	"github.com/tkanos/gonfig"
)

var secretCache, _ = secretcache.New()

type Configuration struct {
	AppPort     string `env:"APP_PORT"`
	DbHost      string `env:"DB_HOST"`
	DbPort      string `env:"DB_PORT"`
	DbName      string `env:"DB_NAME"`
	DbUser      string `env:"DB_USER"`
	DbPassword  string `env:"DB_PASSWORD"`
	DbSecretArn string `env:"DB_SECRET_ARN"`
}

func GetConfig() (*Configuration, error) {
	configuration := Configuration{}

	var err error
	err = gonfig.GetConf("", &configuration)

	if err != nil {
		return nil, err
	}

	if configuration.DbPassword == "" {
		err := loadPasswordFromSecretManager(&configuration)

		if err != nil {
			return nil, err
		}

	}

	return &configuration, nil
}

func loadPasswordFromSecretManager(configuration *Configuration) error {
	var err error
	var secretString string

	secretString, err = secretCache.GetSecretString(configuration.DbSecretArn)

	if err != nil {
		return err
	}

	var secretData map[string]any
	err = json.Unmarshal([]byte(secretString), &secretData)

	if err != nil {
		return err
	}

	configuration.DbPassword = secretData["password"].(string)

	return nil
}
