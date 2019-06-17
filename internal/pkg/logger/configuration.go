package logger


type Configuration struct {
    zap.Config
}

func (c *Configuration) UnmarshalJSON(data []byte) error {
    // First try unmarshaling a string:
    var preset string
    err := json.Unmarshal(data, &preset)
    if err == nil {
        switch preset {
        case "development":
            *c = Configuration{zap.NewDevelopmentConfig()}
        case "production":
            *c = Configuration{zap.NewProductionConfig()}
        case "production-debug":
            cfg := zap.NewProductionConfig()
            cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
            *c = Configuration{cfg}
        }

        return nil
    }

    return json.Unmarshal(data, &c.Config)
}

func DefaultConfiguration() Configuration {
    return Configuration{zap.NewDevelopmentConfig()}
}
