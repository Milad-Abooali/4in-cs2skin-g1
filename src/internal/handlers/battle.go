package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/validate"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/utils"
	"time"
)

var (
	BattleIndex = make(map[int]*models.Battle)
)

func NewBattle(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Token
	userJWT, vErr, ok := validate.RequireString(data, "token", false)
	if !ok {
		return resR, vErr
	}
	resp, err := utils.VerifyJWT(userJWT)
	if err != nil {
		return resR, models.HandlerError{}
	}
	errCode, status, errType := utils.SafeExtractErrorStatus(resp)
	if status != 1 {
		errR.Type = errType
		errR.Code = errCode
		if resp["data"] != nil {
			errR.Data = resp["data"]
		}
		return resR, errR
	}
	userData := resp["data"].(map[string]interface{})
	profile := userData["profile"].(map[string]interface{})
	userID := int(profile["id"].(float64))
	displayName := profile["display_name"].(string)

	// Make Battle
	newBattle := &models.Battle{
		PlayerType: fmt.Sprintf("%v", data["playerType"]),
		Options:    castStringSlice(data["options"]),
		Cases:      castCases(data["cases"]),
		Players:    []int{},
		CreatedBy:  0,
		Status:     "pending",
		Slots:      make(map[string]models.Slot),
		PFair:      make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}

	// Count Cases
	if casesArr, ok := data["cases"].([]interface{}); ok {
		for _, c := range casesArr {
			if m, ok := c.(map[string]interface{}); ok {
				for _, v := range m {
					if count, ok := v.(float64); ok { // JSON numbers -> float64
						newBattle.CaseCounts += int(count)
						break // first
					}
				}
			}
		}
	}
	if newBattle.CaseCounts < 1 {
		errR.Type = "INVALID_TYPE_OR_FORMAT"
		errR.Code = 5003
		errR.Data = map[string]interface{}{
			"fieldName": "cases",
			"fieldType": "[{caseID:count}]",
		}
		return resR, errR
	}

	// Cal Price

	// Add Steps
	newBattle.Summery.Steps = make(map[string][]int)
	for i := 1; i <= newBattle.CaseCounts; i++ {
		key := fmt.Sprintf("r%d", i)
		newBattle.Summery.Steps[key] = []int{}
	}

	// Fit Slots
	var slots int
	switch data["playerType"] {
	case "1v1":
		slots = 2
	case "1v1v1":
		slots = 3
	case "1v1v1v1", "2v2":
		slots = 4
	case "1v6", "2v2v2", "3v3":
		slots = 6
	default:
		errR.Type = "INVALID_TYPE_OR_FORMAT"
		errR.Code = 5003
		errR.Data = map[string]interface{}{
			"fieldName": "playerType",
			"fieldType": "eNum 0v0",
		}
		return resR, errR
	}
	newBattle.Slots = make(map[string]models.Slot)
	for i := 1; i <= slots; i++ {
		key := fmt.Sprintf("s%d", i)
		newBattle.Slots[key] = models.Slot{
			Type: "Empty",
		}
	}

	// Join Battle
	newBattle.Players = append(newBattle.Players, userID)
	newBattle.CreatedBy = userID
	newBattle.Slots["1"] = models.Slot{
		ID:          userID,
		DisplayName: displayName,
		Type:        "Player",
	}

	// Success
	resR.Type = "newBattle"
	resR.Data = newBattle
	return resR, errR
}

func castStringSlice(val interface{}) []string {
	out := []string{}
	if arr, ok := val.([]interface{}); ok {
		for _, v := range arr {
			out = append(out, fmt.Sprintf("%v", v))
		}
	}
	return out
}

func castCases(val interface{}) []map[string]int {
	out := []map[string]int{}
	if arr, ok := val.([]interface{}); ok {
		for _, v := range arr {
			if m, ok := v.(map[interface{}]interface{}); ok { // بسته به decode WS
				newMap := map[string]int{}
				for key, value := range m {
					k := fmt.Sprintf("%v", key)
					if n, ok := value.(float64); ok {
						newMap[k] = int(n)
					}
				}
				out = append(out, newMap)
			} else if m, ok := v.(map[string]interface{}); ok {
				newMap := map[string]int{}
				for key, value := range m {
					if n, ok := value.(float64); ok {
						newMap[key] = int(n)
					}
				}
				out = append(out, newMap)
			}
		}
	}
	return out
}
