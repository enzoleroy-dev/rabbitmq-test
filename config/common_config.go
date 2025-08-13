package config

import (
	"fmt"
	"sync"
)

type Log struct {
	Level string `env:"LOG_LEVEL"  envDefault:"debug"`
}

var once sync.Once

func prefix(e string) string {
	if e == "" {
		return ""
	}

	return fmt.Sprintf("%s_", e)
}
