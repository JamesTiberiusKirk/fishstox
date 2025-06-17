package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbUser string
	DbPass string
	DbHost string
	DbName string
}

func GetConfig() Config {
	_ = godotenv.Load()

	user := os.Getenv("DB_USER")
	if user == "" {
		panic("DB_USER not set")
	}

	pass := os.Getenv("DB_PASS")
	if pass == "" {
		panic("DB_PASS not set")
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		panic("DB_HOST not set")
	}

	name := os.Getenv("DB_NAME")
	if name == "" {
		panic("DB_NAME not set")
	}

	return Config{
		DbUser: user,
		DbPass: pass,
		DbHost: host,
		DbName: name,
	}
}
