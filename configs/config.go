package configs

import (
	"github.com/spf13/viper"
)

var cfg *conf

type conf struct {
	RateLimitPerSecond           string `mapstructure:"RATE_LIMIT_PER_SECOND"`
	WebServerPort                string `mapstructure:"WEB_SERVER_PORT"`
	RateLimitSecret              string `mapstructure:"RATE_LIMIT_SECRET"`
	RateLimitWithSecretPerSecond string `mapstructure:"RATE_LIMIT_WITH_SECRET_PER_SECOND"`
}

func LoadConfig(path string) (*conf, error) {
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}

	return cfg, err
}
