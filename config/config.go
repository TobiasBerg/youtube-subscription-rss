package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	Port                string `split_words:"true"`
	YoutubeClientID     string `required:"true" split_words:"true"`
	YoutubeClientSecret string `required:"true" split_words:"true"`
	YoutubeRefreshToken string `required:"true" split_words:"true"`
}

func CreateConfig() (AppConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return AppConfig{}, err
	}

	var config AppConfig
	err = envconfig.Process("yt_rss", &config)
	if err != nil {
		return AppConfig{}, err
	}

	return config, nil
}
