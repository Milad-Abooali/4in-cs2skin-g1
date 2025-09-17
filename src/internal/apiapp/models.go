package apiapp

var apiURL = "https://api.main.cs2skin.com"

type Payload struct {
	GID         int    `json:"g_id"`
	GTime       string `json:"g_time"`
	DisplayName string `json:"displayName"`
	Bet         string `json:"bet"`
	Multiplier  string `json:"multiplier"`
	Payout      string `json:"payout"`
}
