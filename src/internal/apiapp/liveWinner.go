package apiapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func InsertWinner(gid int, gtime time.Time, displayName string, bet string, multiplier string, payout string) error {
	payload := Payload{
		GID:         gid,
		GTime:       gtime.Format("2006-01-02 15:04:05"),
		DisplayName: displayName,
		Bet:         bet,
		Multiplier:  multiplier,
		Payout:      payout,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL+"/recent_winners/insert", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	apiKey := os.Getenv("API_APP_KEY")
	if apiKey == "" {
		return fmt.Errorf("API_KEY env variable is not set")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}
