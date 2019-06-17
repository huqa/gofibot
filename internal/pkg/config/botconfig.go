package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/huqa/gofibot/internal/pkg/logger"
)

type BotConfiguration struct {
	Nick        string   `json:"nick"`
	Ident       string   `json:"ident"`
	Realname    string   `json:"realname"`
	Version     string   `json:"version"`
	QuitMessage string   `json:"quitMessage"`
	Server      string   `json:"server"`
	Channels    []string `json:"channels"`
}

func (c BotConfiguration) String() string {
	return toJSON(c)
}

func toJSON(c interface{}) string {
	bytes, err := json.Marshal(c)
	if err != nil {
		logger.Fatal(err)
		return ""
	}
	return string(bytes)
}

func LoadBotConfiguration(filePath string) (BotConfiguration, error) {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return BotConfiguration{}, err
	}
	var config BotConfiguration
	err = json.Unmarshal(raw, &config)
	return config, err
}
