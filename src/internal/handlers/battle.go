package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/utils"
	"log"
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
	user, err := utils.VerifyJWT(data["token"].(string))
	if err != nil {
		return models.HandlerOK{}, models.HandlerError{}
	}

	userData := user["data"].(map[string]interface{})
	profile := userData["profile"].(map[string]interface{})

	id := int(profile["id"].(float64))
	displayName := profile["display_name"].(string)

	log.Println(id, displayName)

	// Make Battle
	var userID int = 1

	newBattle := &models.Battle{
		PlayerType: fmt.Sprintf("%v", data["playerType"]),
		Options:    castStringSlice(data["options"]),
		Cases:      castCases(data["cases"]),
		Players:    []int{userID},
		CreatedBy:  userID,
		Status:     "pending",
		Slots:      make(map[string]models.Slot),
		PFair:      make(map[string]interface{}),
		CreatedAt:  time.Now(),
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
