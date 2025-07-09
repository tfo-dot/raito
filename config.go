package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	DiscordToken  string `json:"DISCORD_BOT_TOKEN"`
	TatsuAPI      string `json:"TATSU_API_TOKEN"`
	TatsuMaxScore int    `json:"TATSU_MAX_SCORE"`
	TatsuMinScore int    `json:"TATSU_MIN_SCORE"`
	DiscordGid    string `json:"DISCORD_GID"`
	DiscordLogCid string `json:"DISCORD_LOG_CHANNEL"`
}

func ReadConfigFile() (*Config, error) {
	var config Config

	configFile := os.Args[1]

	data, err := os.ReadFile(configFile)

	if err != nil {
		slog.Error("error while reading config file", slog.Any("err", err))
		return nil, err
	}

	if err = json.Unmarshal(data, &config); err != nil {
		slog.Error("error while unmarshalling config file", slog.Any("err", err))
		return nil, err
	}

	return &config, nil
}

func ReadConfig() (*Config, error) {
	if len(os.Args) > 1 {
		return ReadConfigFile()
	}

	conf := Config{
		DiscordToken:  os.Getenv("DISCORD_BOT_TOKEN"),
		TatsuAPI:      os.Getenv("TATSU_API_TOKEN"),
		TatsuMaxScore: mustParseInt(os.Getenv("TATSU_MAX_SCORE")),
		TatsuMinScore: mustParseInt(os.Getenv("TATSU_MIN_SCORE")),
		DiscordGid:    os.Getenv("DISCORD_GID"),
		DiscordLogCid: os.Getenv("DISCORD_LOG_CHANNEL"),
	}

	return &conf, nil
}

func mustParseInt(str string) int {
	val, err := strconv.Atoi(str)

	if err != nil {
		panic(err)
	}

	return val
}
