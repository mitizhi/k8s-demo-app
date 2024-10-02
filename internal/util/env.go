package util

import (
	"os"
)

func GetEnvDefault(envvar string, defval string) string {
	val := os.Getenv(envvar)
	if val == "" {
		val = defval
	}
	return val
}
