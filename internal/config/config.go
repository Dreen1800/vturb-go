package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	Env      string
	Database DatabaseConfig
	Redis    RedisConfig
	Cloudflare CloudflareConfig
	Embed    EmbedConfig
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL  string
	Addr string
}

type CloudflareConfig struct {
	AccountID string
	APIToken  string
}

type EmbedConfig struct {
	APIURL string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := &Config{
		Port:     getEnv("PORT", "8080"),
		Env:      getEnv("ENV", "development"),
		Database: DatabaseConfig{URL: getEnv("DATABASE_URL", "postgres://vturb:vturb@localhost:5433/vturb")},
		Redis:    RedisConfig{URL: getEnv("REDIS_URL", "redis://localhost:6380/0"), Addr: getRedisAddr(getEnv("REDIS_URL", "redis://localhost:6380/0"))},
		Cloudflare: CloudflareConfig{
			AccountID: getEnv("CLOUDFLARE_STREAM_ACCOUNT_ID", ""),
			APIToken:  getEnv("CLOUDFLARE_STREAM_API_TOKEN", ""),
		},
		Embed: EmbedConfig{
			APIURL: getEnv("EMBED_API_URL", "http://localhost:8080"),
		},
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: %s is not a valid integer, using default %d", key, defaultValue)
		return defaultValue
	}
	return value
}

func getRedisAddr(redisURL string) string {
	// Remove redis:// prefix
	addr := redisURL
	if len(addr) > 8 && addr[:8] == "redis://" {
		addr = addr[8:]
	}
	// Remove database suffix (/0, /1, etc)
	if idx := len(addr) - 1; idx > 0 {
		for i := len(addr) - 1; i >= 0; i-- {
			if addr[i] == '/' {
				addr = addr[:i]
				break
			}
		}
	}
	return addr
}
