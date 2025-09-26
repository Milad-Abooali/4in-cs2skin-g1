package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// UMRequestData defines the request data structure for user management operations.
type UMRequestData struct {
	XKey   string `json:"X_KEY"`
	Token  string `json:"token"`
	UserID int    `json:"userID"`
}

// UMRequest wraps the request type and data for UM API calls.
type UMRequest struct {
	Type string        `json:"type"`
	Data UMRequestData `json:"data"`
}

// VerifyJWT validates a JWT token with the UM API.
func VerifyJWT(userToken string) (map[string]interface{}, error) {
	env := os.Getenv("API_UM") // Example: "https://um.main.cs2skin.com/web, appToken, xKey"
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return result, nil
}

// GetUser fetches user details by userID from the UM API.
func GetUser(userID int) (map[string]interface{}, error) {
	env := os.Getenv("API_UM") // Example: "https://um.main.cs2skin.com/web, appToken, xKey"
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return result, nil
}

/* Transaction */

// UMTransactionData defines the structure of transaction data.
type UMTransactionData struct {
	XKey        string  `json:"X_KEY"`
	UserID      int     `json:"userID"`
	Type        string  `json:"type"`        // e.g. "req_withdrawal"
	ReferenceID string  `json:"referenceID"` // reference identifier
	Amount      float64 `json:"amount"`
	TxRef       string  `json:"txRef"`
	Description string  `json:"description"`
}

// UMTransactionRequest wraps the request type and data for transactions.
type UMTransactionRequest struct {
	Type string            `json:"type"` // "xAddTransaction"
	Data UMTransactionData `json:"data"`
}

// AddTransaction sends a transaction request to the UM API.
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return result, nil
}

/* XP */

// UMXpData defines the structure of XP data.
type UMXpData struct {
	XKey      string `json:"X_KEY"`
	UserID    int    `json:"userID"`
	Amount    int    `json:"amount"`
	Reason    string `json:"reason"`
	CreatedBy string `json:"createdBy"`
}

// UMXpRequest wraps the request type and data for XP.
type UMXpRequest struct {
	Type string   `json:"type"` // "xAddXp"
	Data UMXpData `json:"data"`
}

// AddXp sends an XP request to the UM API.
func AddXp(userID, amount int, reason, createdBy string) (map[string]interface{}, error) {
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

	reqBody := UMXpRequest{
		Type: "xAddXp",
		Data: UMXpData{
			XKey:      xKey,
			UserID:    userID,
			Amount:    amount,
			Reason:    reason,
			CreatedBy: createdBy,
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	return result, nil
}
