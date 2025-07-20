package keycloak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type KeycloakClient struct {
	BaseURL        string
	Realm          string
	ClientID       string
	ClientSecret   string
	Username       string
	Password       string
	IdentityProvider string
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func NewKeycloakClient() *KeycloakClient {
	return &KeycloakClient{
		BaseURL:        os.Getenv("KEYCLOAK_URL"),
		Realm:          os.Getenv("REALM"),
		ClientID:       os.Getenv("CLIENT_ID"),
		ClientSecret:   os.Getenv("CLIENT_SECRET"),
		Username:       os.Getenv("USERNAME"),
		Password:       os.Getenv("PASSWORD"),
		IdentityProvider: os.Getenv("IDENTITY_PROVIDER"),
	}
}

func (kc *KeycloakClient) GetAccessToken() (string, error) {
	url := fmt.Sprintf("%s/auth/realms/%s/protocol/openid-connect/token", kc.BaseURL, kc.Realm)
	data := fmt.Sprintf("client_id=%s&client_secret=%s&username=%s&password=%s&grant_type=password",
		kc.ClientID, kc.ClientSecret, kc.Username, kc.Password)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get access token: %s", resp.Status)
	}

	var tokenResponse AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func (kc *KeycloakClient) MapperExists(mapperName, accessToken string) (bool, error) {
	url := fmt.Sprintf("%s/auth/admin/realms/%s/identity-provider/instances/%s/mappers", kc.BaseURL, kc.Realm, kc.IdentityProvider)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to check mapper existence: %s", resp.Status)
	}

	var mappers []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&mappers); err != nil {
		return false, err
	}

	for _, mapper := range mappers {
		if mapper["name"] == mapperName {
			return true, nil
		}
	}

	return false, nil
}

func (kc *KeycloakClient) CheckAndCreateGroup(groupName, accessToken string) (bool, error) {
    url := fmt.Sprintf("%s/auth/admin/realms/%s/groups", kc.BaseURL, kc.Realm)
    headers := map[string]string{
        "Authorization": "Bearer " + accessToken,
        "Content-Type":  "application/json",
    }

    // Проверяем существование группы
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return false, err
    }
    for k, v := range headers {
        req.Header.Set(k, v)
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return false, fmt.Errorf("ошибка получения групп: %d", resp.StatusCode)
    }

    var groups []map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
        return false, err
    }
    for _, group := range groups {
        if group["name"] == groupName {
            return true, nil // Группа уже есть
        }
    }

    // Если группы нет — создаём
    newGroup := map[string]string{"name": groupName}
    body, _ := json.Marshal(newGroup)
    req, err = http.NewRequest("POST", url, bytes.NewReader(body))
    if err != nil {
        return false, err
    }
    for k, v := range headers {
        req.Header.Set(k, v)
    }
    resp, err = client.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusCreated {
        return true, nil // Группа создана
    }
    return false, fmt.Errorf("ошибка создания группы: %d", resp.StatusCode)
}