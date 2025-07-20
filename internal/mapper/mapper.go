package mapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"go-keycloak-mapper/internal/keycloak" // Update with your actual module path
)

type Mapper struct {
	KeycloakClient *keycloak.KeycloakClient
}

func NewMapper(client *keycloak.KeycloakClient) *Mapper {
	return &Mapper{
		KeycloakClient: client,
	}
}

func (m *Mapper) CreateMapper(mapperName, accessToken string) error {
	// Logic to create a mapper in Keycloak
	// Example: Check if the mapper exists, if not, create it
	exists, err := m.KeycloakClient.MapperExists(mapperName, accessToken)
	if err != nil {
		return fmt.Errorf("error checking if mapper exists: %w", err)
	}

	if exists {
		fmt.Printf("Mapper %s already exists. Skipping creation.\n", mapperName)
		return nil
	}

	groups, err := json.Marshal([]map[string]string{{"key": mapperName, "value": mapperName}})
	if err != nil {
		return fmt.Errorf("error marshalling groups: %w", err)
	}
	claims, err := json.Marshal([]map[string]string{{"key": "group", "value": fmt.Sprintf("/%s", mapperName)}})
	if err != nil {
		return fmt.Errorf("error marshalling claims: %w", err)
	}

	// Prepare the request to create the mapper
	mapperData := map[string]interface{}{
		"name":                   mapperName,
		"identityProviderAlias":  os.Getenv("IDENTITY_PROVIDER"),
		"identityProviderMapper": "oidc-advanced-group-idp-mapper",
		"config": map[string]interface{}{
			"syncMode": "FORCE",
			"groups":   string(groups),
			"claims":   string(claims),
			"group":    fmt.Sprintf("/%s", mapperName),
		},
	}

	url := fmt.Sprintf("%s/auth/admin/realms/%s/identity-provider/instances/%s/mappers",
		os.Getenv("KEYCLOAK_URL"), os.Getenv("REALM"), os.Getenv("IDENTITY_PROVIDER"))

	reqBody, err := json.Marshal(mapperData)
	if err != nil {
		return fmt.Errorf("error marshalling mapper data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, ioutil.NopCloser(bytes.NewReader(reqBody)))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("error creating mapper: %s", body)
	}

	fmt.Printf("Successfully created mapper: %s\n", mapperName)
	return nil
}
