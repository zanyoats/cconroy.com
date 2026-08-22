package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	WriteDSN, ReadDSN string
}

const (
	AppName = "com.cconroy"
	DBName  = "personalSite.db"
)

func InitializeConfig() (*Config, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	result := new(Config)

	appDir := filepath.Join(dir, AppName)
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(appDir, DBName)
	sqliteWritePragmas := []string{
		"mode=rwc",
		"cache=private",
		"_foreign_keys=on",
		"_synchronous=NORMAL",
		"_busy_timeout=5000",
		"_cache_size=-32768", // '-' means use kilobytes, thus 32 MB page store
		"_txlock=immediate",
	}
	sqliteReadPragmas := []string{
		"mode=ro",
		"cache=private",
		"_foreign_keys=on",
		"_busy_timeout=5000",
		"_cache_size=-32768",
		"_query_only=on",
	}

	result.WriteDSN = fmt.Sprintf("%s?%s", dbPath, strings.Join(sqliteWritePragmas, "&"))
	result.ReadDSN = fmt.Sprintf("%s?%s", dbPath, strings.Join(sqliteReadPragmas, "&"))
	return result, nil
}
