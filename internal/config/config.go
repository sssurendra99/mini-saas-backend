package config

import(
	"os"
	"log"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	DBHost string
	DBPort string
	DBUser string
	DBPassword string
	DBName string
	DBurl string
}

// This is the constructor pattern.
func Load() *Config {

	err := godotenv.Load()
	if err != nil{
		log.Println(".env file not found!")
	}

	return &Config{
		Port: os.Getenv("PORT"),
		DBHost: os.Getenv("DB_HOST"),
		DBPort: os.Getenv("DB_PORT"),
		DBUser: os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName: os.Getenv("DB_NAME"),
		DBurl: os.Getenv("DB_URL"),
	}
}