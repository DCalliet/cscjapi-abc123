package devtools

import (
	"os"

	"github.com/spf13/viper"
)

func LoadConfig() (v *viper.Viper, err error) {
	v = viper.New()

	v.SetEnvPrefix("csc")
	v.BindEnv("ENV")

	v.BindEnv("APP_CONFIG_PATH")
	v.BindEnv("APP_CONFIG_FILENAME")

	v.BindEnv("HTTP_PORT")

	v.BindEnv("DB_USERNAME")
	v.BindEnv("DB_PASSWORD")
	v.BindEnv("DB_NAME")
	v.BindEnv("DB_PORT")
	v.BindEnv("DB_HOSTNAME")

	v.BindEnv("REDIS_HOSTNAME")
	v.BindEnv("REDIS_PASSWORD")
	v.BindEnv("REDIS_PORT")

	v.BindEnv("SWAGGER_PATH")
	v.BindEnv("SWAGGER_FILENAME")

	// Set Default Values where appropriate
	v.SetDefault("HTTP_PORT", 3001)
	v.SetDefault("ENV", "local")
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	v.SetDefault("APP_CONFIG_FILENAME", v.GetString("ENV"))
	v.SetDefault("APP_CONFIG_PATH", dir)

	// Loads configuration file from ./appconfig.{yaml|json}
	v.SetConfigName(v.GetString("APP_CONFIG_FILENAME"))
	v.AddConfigPath(v.GetString("APP_CONFIG_PATH"))
	_ = v.ReadInConfig()
	return
}

type ApplicationConfig struct {
	EnvName          string `json:"ENV" mapstructure:"ENV"`
	EnvHTTPPort      int    `json:"HTTP_PORT" mapstructure:"HTTP_PORT"`
	EnvDBUsername    string `json:"DB_USERNAME" mapstructure:"DB_USERNAME"`
	EnvDBPassword    string `json:"DB_PASSWORD" mapstructure:"DB_PASSWORD"`
	EnvDBName        string `json:"DB_NAME" mapstructure:"DB_NAME"`
	EnvDBHostname    string `json:"DB_HOSTNAME" mapstructure:"DB_HOSTNAME"`
	EnvDBPort        string `json:"DB_PORT" mapstructure:"DB_PORT"`
	EnvRedisHostname string `json:"REDIS_HOSTNAME" mapstructure:"REDIS_HOSTNAME"`
	EnvRedisPassword string `json:"REDIS_PASSWORD" mapstructure:"REDIS_PASSWORD"`
	EnvRedisPort     string `json:"REDIS_PORT" mapstructure:"REDIS_PORT"`
}

type SwaggerConfig struct {
	EnvName            string `json:"ENV" mapstructure:"ENV"`
	EnvHTTPPort        int    `json:"HTTP_PORT" mapstructure:"HTTP_PORT"`
	EnvSwaggerPath     string `json:"SWAGGER_PATH" mapstructure:"SWAGGER_PATH"`
	EnvSwaggerFilename string `json:"SWAGGER_FILENAME" mapstructure:"SWAGGER_FILENAME"`
}
