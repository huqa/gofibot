package config

import (
    "encoding/json"
    "github.com/huqa/gofibot/internal/pkg/logger"
    "io/ioutil"
    "os"
)

type Configuration struct {
    Nick string `json:"nick"`
    Ident string `json:"ident"`
    Realname string `json:"realname"`
    Version string `json:"version"`
    QuitMessage string `json:"quitMessage"`
    Server string `json:"server"`
    Channels []string `json:"channels"`
}

func (c Configuration) ToString() string {
    return toJSON(c)
}

func toJSON(c interface{}) string {
    bytes, err := json.Marshal(c)
    if err != nil {
        logger.Fatal(err)
        os.Exit(1)
    }
    return string(bytes)
}

func LoadConfiguration(filePath string) (Configuration, error) {
    raw, err := ioutil.ReadFile(filePath)
    if err != nil {
        logger.Fatal(err)
        os.Exit(1)
    }
    var config Configuration
    err = json.Unmarshal(raw, &config)
    return config, err
}
