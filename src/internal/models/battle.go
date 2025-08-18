package models

import "time"

type Slot struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	ClientSeed  string `json:"client_seed"`
	Type        string `json:"type"` // player یا bot
}

type Summery struct {
	Steps   map[string][]int `json:"steps"` // r1, r2, ...
	Winners []string         `json:"winners"`
	Prizes  []float64        `json:"prizes"`
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
	Cases      []map[string]int       `json:"cases"` // {"7":2}, {"2":1}
	CaseCounts int                    `json:"caseCounts"`
	Cost       float64                `json:"cost"`
	Slots      map[string]Slot        `json:"slots"` // s1, s2, ...
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
	Summery    SummeryResponse     `json:"summery"`
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
	ID         int              `json:"id"`
	PlayerType string           `json:"playerType"`
	Options    []string         `json:"options"`
	Cases      []map[string]int `json:"cases"`
	CaseCounts int              `json:"caseCounts"`
	Cost       float64          `json:"cost"`
	Slots      map[string]Slot  `json:"slots"`
	Status     string           `json:"status"`
	StatusCode int              `json:"statusCode"`
	Summery    Summery          `json:"summery"`
	CreatedAt  time.Time        `json:"createdAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`
	ServerSeed string           `json:"serverSeed"`
}
