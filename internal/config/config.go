package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr                        string
	DBUrl                       string
	SkinportClientID            string
	SkinportClientSecret        string
	SkinportAddr                string
	CacheTTLSeconds             int
	CacheCleanUpIntervalSeconds int
}

func Load() *Config {
	_ = godotenv.Load() // Load from .env file

	conf := &Config{
		CacheTTLSeconds:             getInt("CACHE_TTL", 300),             // by default, cache ttl is 5 minutes.
		CacheCleanUpIntervalSeconds: getInt("CACHE_CLEANUP_INTERVAL", 60), // by default, cache clean up interval is a minute.
		Addr:                        mustGetEnv("ADDR"),
		DBUrl:                       mustGetEnv("DB_URL"),
		SkinportAddr:                mustGetEnv("SKINPORT_ADDR"),
		SkinportClientID:            os.Getenv("SKINPORT_CLIENT_ID"),
		SkinportClientSecret:        os.Getenv("SKINPORT_CLIENT_SECRET"),
	}

	return conf
}

// mustGetEnv will panic, if required env var was not specified.
func mustGetEnv(key string) string {
	value, found := os.LookupEnv(key)
	if !found {
		panic(fmt.Errorf("environment variable %s not set", key))
	}
	return value
}

func getInt(key string, defaultValue int) int {
	value, found := os.LookupEnv(key)
	if found {
		valueInt, err := strconv.Atoi(value)
		if err == nil {
			return valueInt
		}
	}
	return defaultValue
}
