package config

import (
	"fmt"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Debug         bool
	RTXMode       bool
	TGToken       string
	DBDriver      string
	DBString      string
	QueuePool     int
	QueueParallel int
}

func GetConfig() Config {
	config := Config{
		Debug:    getenv("DEBUG", false) == "1",
		RTXMode:  getenv("RTX_MODE", false) == "on",
		TGToken:  getenv("TG_TOKEN", true),
		DBDriver: getenv("DB_DRIVER", true),
		DBString: getenv("DB_STRING", true),
	}

	var err error

	config.QueuePool, err = strconv.Atoi(getenv("QUEUE_POOL", true))
	if err != nil {
		panic(fmt.Sprintf("config: invalid QUEUE_POOL value %s", err.Error()))
	}

	config.QueueParallel, err = strconv.Atoi(getenv("QUEUE_PARALLEL", true))
	if err != nil {
		panic(fmt.Sprintf("config: invalid QUEUE_PARALLEL value %s", err.Error()))
	}

	return config
}

func getenv(key string, required bool) string {
	value := os.Getenv(key)
	if value == "" && required {
		panic(fmt.Sprintf("config: %s variable is required but not set", key))
	}
	return value
}
