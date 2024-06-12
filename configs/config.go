package configs

import (
	"github.com/spf13/viper"
)

type conf struct {
	RateLimitWithIPPerSecond        string `mapstructure:"RATE_LIMIT_WITH_IP_PER_SECOND"`
	RateLimitWithTokenPerSecond     string `mapstructure:"RATE_LIMIT_WITH_TOKEN_PER_SECOND"`
	RateLimitBlockDurationInMinutes string `mapstructure:"RATE_LIMIT_BLOCK_DURATION_IN_MINUTES"`
	WebServerPort                   string `mapstructure:"WEB_SERVER_PORT"`
}

func LoadConfig(path string) (*conf, error) {
	var cfg *conf

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
