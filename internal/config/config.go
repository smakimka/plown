package config

import "github.com/caarlos0/env"

type ServerConfig struct {
	Addr      string `env:"SERVER_ADDR" envDefault:":8080"`
	StaticDir string `env:"STATIC_DIR" envDefault:"web/build"`
	PGAddr    string `env:"PG_ADDR" envDefault:"localhost:5432"`
	PGDB      string `env:"PG_DB" envDefault:"parser_db"`
	PGUser    string `env:"PG_USER" envDefault:"parser"`
	PGPass    string `env:"PG_PASS" envDefault:"parser_password"`
}

func Parse[T any]() (T, error) {
	var cfg T

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
