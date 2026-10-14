package env

import (
	"fmt"
	"os"
	"github.com/joho/godotenv"
)

func LoadEnvs() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("LoadEnvs(): Error loading .env file: %w", err)
	}
	return nil
}

func GetEnvByKey(envKey string) (string, error) {
	env, exists := os.LookupEnv(envKey)
	if !exists {
		return "", fmt.Errorf("GetEnvByKey(%s): Env not found", envKey)
	}
	return env, nil
}

func GetDbConnectionString() (string, error){
	host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")

	if host == "" && port == "" && user == "" && password == "" && dbname == "" {
		return "", fmt.Errorf("GetDbConnectionString(): Required empty envs for db connection string")
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname), nil
}
