package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/szykes/simple-backend/errors"
	"github.com/szykes/simple-backend/models"
)

type Config struct {
	PSQL models.PostgresCfg
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Host string
		Port string
	}
}

func LoadDotEnvConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, errors.Wrap(err, "load .env file")
	}

	var cfg Config
	if cfg.PSQL.Host, err = stringEnv("PSQL_HOST"); err != nil {
		return nil, errors.Wrap(err, "failed to load .env file")
	}
	if cfg.PSQL.Port, err = stringEnv("PSQL_PORT"); err != nil {
		return nil, errors.Wrap(err, "failed to load .env file")
	}
	if cfg.PSQL.User, err = stringEnv("PSQL_USER"); err != nil {
		return nil, errors.Wrap(err, "failed to load .env file")
	}
	if cfg.PSQL.Password, err = stringEnv("PSQL_PASSWORD"); err != nil {
		return nil, errors.Wrap(err, "failed to load .env file")
	}
	if cfg.PSQL.Database, err = stringEnv("PSQL_DATABASE"); err != nil {
		return nil, errors.Wrap(err, "failed to load .env file")
	}
	if cfg.PSQL.SSLMode, err = stringEnv("PSQL_SSL_MODE"); err != nil {
		return nil, errors.Wrap(err, "failed to load .env file")
	}

	if cfg.CSRF.Key, err = stringEnv("CSRF_KEY"); err != nil {
		return nil, errors.Wrap(err, "failed to load .env file")
	}
	if cfg.CSRF.Secure, err = boolEnv("CFRF_SECURE"); err != nil {
		return nil, errors.Wrap(err, "failed to load .env file")
	}

	if cfg.Server.Host, err = stringEnv("SERVER_HOST"); err != nil {
		return nil, errors.Wrap(err, "failed to load .env file")
	}
	if cfg.Server.Port, err = stringEnv("SERVER_PORT"); err != nil {
		return nil, errors.Wrap(err, "failed to load .env file")
	}
	return &cfg, nil
}

func stringEnv(key string) (string, error) {
	value := os.Getenv(key)
	if len(value) == 0 {
		return "", errors.New("string env empty", "key", key)
	}
	return value, nil
}

func boolEnv(key string) (bool, error) {
	env := os.Getenv(key)
	switch env {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errors.New("non Boolean value", "key", key)
	}
}
