package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type UMRequestData struct {
	XKey   string `json:"X_KEY"`
	Token  string `json:"token"`
	UserID int    `json:"userID"`
}

type UMRequest struct {
	Type string        `json:"type"`
	Data UMRequestData `json:"data"`
}

func VerifyJWT(userToken string) (map[string]interface{}, error) {
	env := os.Getenv("API_UM") // مثال: "https://um.main.cs2skin.com/web, 4fb1c6d6a5be06d65be004e2558bep2r, 1304025bdb3066dfb5c402c63ce1c02bbc6da41"
	parts := make([]string, 3)
	for i, p := range bytes.Split([]byte(env), []byte(",")) {
		if i < 3 {
			parts[i] = string(bytes.TrimSpace(p))
		}
	}
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid API_UM env format")
	}
	baseURL := parts[0]
	appToken := parts[1]
	xKey := parts[2]

	reqBody := UMRequest{
		Type: "xGetJWT",
		Data: UMRequestData{
			XKey:  xKey,
			Token: userToken,
		},
	}
	jsonBody, err := json.Marshal(reqBody)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+appToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return result, nil
}

func GetUser(userID int) (map[string]interface{}, error) {
	env := os.Getenv("API_UM") // مثال: "https://um.main.cs2skin.com/web, 4fb1c6d6a5be06d65be004e2558bep2r, 1304025bdb3066dfb5c402c63ce1c02bbc6da41"
	parts := make([]string, 3)
	for i, p := range bytes.Split([]byte(env), []byte(",")) {
		if i < 3 {
			parts[i] = string(bytes.TrimSpace(p))
		}
	}
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid API_UM env format")
	}
	baseURL := parts[0]
	appToken := parts[1]
	xKey := parts[2]

	reqBody := UMRequest{
		Type: "xGetUser",
		Data: UMRequestData{
			XKey:   xKey,
			UserID: userID,
		},
	}
	jsonBody, err := json.Marshal(reqBody)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+appToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return result, nil
}

type UMTransactionData struct {
	XKey        string  `json:"X_KEY"`
	UserID      int     `json:"userID"`
	Type        string  `json:"type"`        // مثلا "req_withdrawal"
	ReferenceID string  `json:"referenceID"` // شناسه مرجع
	Amount      float64 `json:"amount"`
	TxRef       string  `json:"txRef"`
	Description string  `json:"description"`
}

type UMTransactionRequest struct {
	Type string            `json:"type"` // "xAddTransaction"
	Data UMTransactionData `json:"data"`
}

func AddTransaction(userID int, txType, referenceID string, amount float64, txRef, description string) (map[string]interface{}, error) {
	env := os.Getenv("API_UM")
	parts := make([]string, 3)
	for i, p := range bytes.Split([]byte(env), []byte(",")) {
		if i < 3 {
			parts[i] = string(bytes.TrimSpace(p))
		}
	}
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid API_UM env format")
	}
	baseURL := parts[0]
	appToken := parts[1]
	xKey := parts[2]

	reqBody := UMTransactionRequest{
		Type: "xAddTransaction",
		Data: UMTransactionData{
			XKey:        xKey,
			UserID:      userID,
			Type:        txType,
			ReferenceID: referenceID,
			Amount:      amount,
			TxRef:       txRef,
			Description: description,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+appToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return result, nil
}
