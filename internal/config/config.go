package config

import (
    "encoding/json"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    KeycloakURL      string `json:"keycloak_url"`
    Realm            string `json:"realm"`
    ClientID         string `json:"client_id"`
    ClientSecret     string `json:"client_secret"`
    Username         string `json:"username"`
    Password         string `json:"password"`
    IdentityProvider string `json:"identity_provider"`
    MapperGroupFile  string `json:"mapper_group_file"`
}

func LoadConfig() (*Config, error) {
    // Load environment variables from .env file if it exists
    err := godotenv.Load()
    if err != nil {
        return nil, err
    }

    config := &Config{
        KeycloakURL:      os.Getenv("KEYCLOAK_URL"),
        Realm:            os.Getenv("REALM"),
        ClientID:         os.Getenv("CLIENT_ID"),
        ClientSecret:     os.Getenv("CLIENT_SECRET"),
        Username:         os.Getenv("USERNAME"),
        Password:         os.Getenv("PASSWORD"),
        IdentityProvider: os.Getenv("IDENTITY_PROVIDER"),
        MapperGroupFile:  os.Getenv("MAPPER_GROUP_FILE"),
    }

    return config, nil
}

func (c *Config) SaveToFile(filePath string) error {
    file, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    return encoder.Encode(c)
}