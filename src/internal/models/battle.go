package models

import "time"

type Slot struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	ClientSeed  string `json:"client_seed"`
	Type        string `json:"type"` // player یا bot
}

type StepResult struct {
	Slot   string  `json:"slot"`   // s1, s2, ...
	ItemID int     `json:"itemId"` // ID
	Price  float64 `json:"price"`
}

type Summery struct {
	Steps   map[int][]StepResult `json:"steps"` // r1 → [StepResult, StepResult, ...]
	Winners []string             `json:"winners"`
	Prizes  map[string]float64   `json:"prizes"` // s1 → total prize
}

type HE struct {
	PayIn   float64 `json:"payIn"`
	PayOut  float64 `json:"payOut"`
	Rate    float64 `json:"rate"`
	Balance float64 `json:"balance"`
}

type Battle struct {
	ID         int                    `json:"id"`
	PlayerType string                 `json:"playerType"`
	Options    []string               `json:"options"`
	Cases      []int                  `json:"cases"`
	CaseCounts int                    `json:"caseCounts"`
	Cost       float64                `json:"cost"`
	Slots      map[string]Slot        `json:"slots"`
	Players    []int                  `json:"players"`
	Bots       []int                  `json:"bots"`
	Status     string                 `json:"status"`
	StatusCode int                    `json:"statusCode"`
	Summery    Summery                `json:"summery"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
	CreatedBy  int                    `json:"createdBy"`
	HE         HE                     `json:"he"`
	PFair      map[string]interface{} `json:"pFair"`
	Logs       []BattleLog            `json:"logs"`
}

type BattleLog struct {
	Time   string `json:"time"`
	Action string `json:"action"`
	UserID int64  `json:"user_id"`
}

type BattleResponse struct {
	ID         int                 `json:"id"`
	PlayerType string              `json:"playerType"`
	Options    []string            `json:"options"`
	CaseCounts int                 `json:"caseCounts"`
	Cost       float64             `json:"cost"`
	Slots      map[string]SlotResp `json:"slots"`
	Status     string              `json:"status"`
	Summery    Summery             `json:"summery"`
	CreatedAt  time.Time           `json:"createdAt"`
}

type SlotResp struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
}

type SummeryResponse struct {
	Steps   map[string][]int `json:"steps"`
	Winners []string         `json:"winners,omitempty"`
	Prizes  []float64        `json:"prizes,omitempty"`
}

type BattleClient struct {
	ID             int             `json:"id"`
	PlayerType     string          `json:"playerType"`
	Options        []string        `json:"options"`
	Cases          []int           `json:"cases"`
	CaseCounts     int             `json:"caseCounts"`
	Cost           float64         `json:"cost"`
	Slots          map[string]Slot `json:"slots"`
	Status         string          `json:"status"`
	StatusCode     int             `json:"statusCode"`
	Summery        Summery         `json:"summery"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	ServerSeedHash string          `json:"serverSeedHash"`
}
