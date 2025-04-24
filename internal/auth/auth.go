package auth

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
)

var (
	ErrUnauthenticated = errors.New("client is unauthenticated")
)

type Client struct {
	Id      string   `json:"id"`
	APIKeys []string `json:"api_keys"`
}

type AuthService struct {
	clients []Client
}

func NewAuthService(clientsFilePath string) (AuthService, error) {
	file, fileErr := os.Open(clientsFilePath)
	if fileErr != nil {
		return AuthService{}, fileErr
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var clients []Client
	decodeErr := decoder.Decode(&clients)
	if decodeErr != nil {
		return AuthService{}, decodeErr
	}

	return AuthService{clients: clients}, nil
}

func (a *AuthService) CheckAPIKey(apiKey string) (Client, error) {
	for _, client := range a.clients {
		if slices.Contains(client.APIKeys, apiKey) {
			return client, nil
		}
	}

	return Client{}, ErrUnauthenticated
}
