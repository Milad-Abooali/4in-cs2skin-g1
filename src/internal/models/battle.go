package models

import "time"

type Slot struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
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
	Summery    Summery                `json:"summery"`
	CreatedAt  time.Time              `json:"createdAt"`
	CreatedBy  int                    `json:"createdBy"`
	HE         HE                     `json:"he"`
	PFair      map[string]interface{} `json:"pfair"`
}
