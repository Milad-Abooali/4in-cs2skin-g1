package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func VerifyGoogleToken(token string) (googleID string, displayName string, err error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "", "", errors.New("invalid id_token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", errors.New("failed to decode payload")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return "", "", errors.New("failed to parse payload")
	}

	sub, ok1 := data["sub"].(string)
	name, ok2 := data["name"].(string)
	if !ok1 {
		return "", "", errors.New("sub field not found")
	}
	if !ok2 {
		name = "GoogleUser"
	}

	return sub, name, nil
}

func VerifyDiscordToken(token string) (discordID string, displayName string, err error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", errors.New("discord token invalid or expired")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var data struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", "", err
	}

	if data.ID == "" {
		return "", "", errors.New("no ID in discord response")
	}

	return data.ID, data.Username, nil
}
