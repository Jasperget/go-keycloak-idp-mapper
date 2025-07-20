package main

import (
	"bufio"
	"log"
	"os"

	"go-keycloak-mapper/internal/config"
	"go-keycloak-mapper/internal/keycloak"
	"go-keycloak-mapper/internal/mapper"
)

func main() {
	// Load configuration from environment variables or file
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	kc := keycloak.NewKeycloakClient()
	accessToken, err := kc.GetAccessToken()
	if err != nil {
		log.Fatalf("Ошибка получения токена: %v", err)
	}

	mapperService := mapper.NewMapper(kc)

	file, err := os.Open(cfg.MapperGroupFile)
	if err != nil {
		log.Fatalf("Не удалось открыть файл %s: %v", cfg.MapperGroupFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		groupName := scanner.Text()
		if groupName == "" {
			continue
		}

		// 1. Проверяем и создаём группу, если её нет
		exists, err := kc.CheckAndCreateGroup(groupName, accessToken)
		if err != nil {
			log.Printf("Ошибка при проверке/создании группы %s: %v", groupName, err)
			continue
		}
		if exists {
			log.Printf("Группа %s уже существует или успешно создана.", groupName)
		}

		// 2. Создаём маппер для этой группы
		err = mapperService.CreateMapper(groupName, accessToken)
		if err != nil {
			log.Printf("Ошибка при создании маппера %s: %v", groupName, err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Ошибка чтения файла: %v", err)
	}

	log.Println("Keycloak mapper application finished.")
}
