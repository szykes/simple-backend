package models

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresCfg struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (c *PostgresCfg) String() string {
	return fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=%v", c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

func DefaultPostgresCfg() PostgresCfg {
	return PostgresCfg{
		Host:     "localhost",
		Port:     "5432",
		User:     "user",
		Password: "some-pass",
		Database: "szykes",
		SSLMode:  "disable",
	}
}

func Open(cfg PostgresCfg) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.String())
	if err != nil {
		return nil, fmt.Errorf("open %w", err)
	}
	return db, nil
}
