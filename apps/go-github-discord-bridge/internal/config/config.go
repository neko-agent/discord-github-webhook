package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	Env                 string
	DiscordBotToken     string
	DiscordForumChID    string
	GitHubWebhookSecret string
	RedisURL            string
	GitHubDiscordUserMap map[string]string // GitHub username → Discord user ID
}

var AppConfig *Config

func Load() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	AppConfig = &Config{
		Port:                 getEnv("PORT", "3000"),
		Env:                  getEnv("ENV", "development"),
		DiscordBotToken:      requireEnv("DISCORD_BOT_TOKEN"),
		DiscordForumChID:     requireEnv("DISCORD_FORUM_CHANNEL_ID"),
		GitHubWebhookSecret:  getEnv("GITHUB_WEBHOOK_SECRET", ""),
		RedisURL:             requireEnv("REDIS_URL"),
		GitHubDiscordUserMap: parseUserMap(getEnv("GITHUB_DISCORD_USER_MAP", "{}")),
	}

	if AppConfig.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Env variable %s is required but not set", key)
	}
	return value
}

func parseUserMap(raw string) map[string]string {
	m := make(map[string]string)
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		log.Printf("Warning: failed to parse GITHUB_DISCORD_USER_MAP: %v", err)
	}
	return m
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
