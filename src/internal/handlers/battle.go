package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/configs"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/apiapp"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/events"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/provablyfair"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/validate"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/utils"
	"google.golang.org/protobuf/types/known/structpb"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	BattleIndex   = make(map[int64]*models.Battle)
	battleIndexMu sync.RWMutex
)

// NewBattle - Handler
func NewBattle(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	if len(CasesImpacted) == 0 {
		FillCaseImpact()
	}

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

	balanceStr := fmt.Sprintf("%v", profile["balance"])
	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		balance = 0
	}

	options := castStringSlice(data["options"])

	// Make Battle
	newBattle := &models.Battle{
		PlayerType: fmt.Sprintf("%v", data["playerType"]),
		Options:    utils.ToLowerArray(options),
		Cases:      expandCases(castCases(data["cases"])),
		CasesUi:    castCases(data["cases"]),
		Players:    []int{},
		CreatedBy:  0,
		Status:     "Waiting Room",
		StatusCode: 0,
		Slots:      make(map[string]models.Slot),
		PFair:      make(map[string]interface{}),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Cases
	if casesArr, ok := data["cases"].([]interface{}); ok {
		for _, c := range casesArr {
			if m, ok := c.(map[string]interface{}); ok {

				// Price
				for caseNumber, v := range m {
					if countFloat, ok := v.(float64); ok { // JSON numbers -> float64
						count := int(countFloat)
						for i := 0; i < count; i++ {
							// Steps - on Case
							caseInt, err := strconv.Atoi(caseNumber)
							if err != nil {
								errR.Type = "INVALID_CASE_ID"
								errR.Code = 1027
								errR.Data = map[string]interface{}{
									"fieldName": "cases",
									"fieldType": "[{caseID:count}]",
								}
								return resR, errR
							}

							if caseData, ok := CasesImpacted[caseInt]; ok {

								// Cal Price
								var price float64
								switch v := caseData["price"].(type) {
								case float64:
									price = v
								case string:
									p, err := strconv.ParseFloat(v, 64)
									if err != nil {
										log.Println("Invalid price value:", v)
										continue
									}
									price = p
								default:
									log.Println("Unknown price type:", v)
									continue
								}
								newBattle.Cost += utils.RoundToTwoDigits(price)

							} else {
								errR.Type = "INVALID_CASE_ID"
								errR.Code = 1027
								errR.Data = map[string]interface{}{
									"fieldName": "cases",
									"fieldType": "[{caseID:count}]",
								}
								return resR, errR
							}

						}
					}
				}

				// Count
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

	// Check Balance
	if balance < newBattle.Cost {
		errR.Type = "INSUFFICIENT_BALANCE"
		errR.Code = 7001
		errR.Data = map[string]interface{}{
			"cost":    newBattle.Cost,
			"balance": balance,
		}
		return resR, errR
	}
	// Add Transaction
	Transaction, err := utils.AddTransaction(
		userID,
		"game_loss",
		"1",
		newBattle.Cost,
		"",
		"Case Battle",
	)
	if err != nil {
		return resR, models.HandlerError{}
	}
	errCode, status, errType = utils.SafeExtractErrorStatus(Transaction)
	if status != 1 {
		errR.Type = errType
		errR.Code = errCode
		if resp["data"] != nil {
			errR.Data = resp["data"]
		}
		return resR, errR
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
	clientSeed := utils.MD5UserID(userID)
	newBattle.Slots["s1"] = models.Slot{
		ID:          userID,
		DisplayName: displayName,
		ClientSeed:  clientSeed,
		Type:        "Player",
	}

	// Provably Fair
	serverSeed, serverSeedHash := provablyfair.GenerateServerSeed()
	newBattle.PFair = map[string]interface{}{
		"serverSeed":     serverSeed,
		"serverSeedHash": serverSeedHash,
		"clientSeed": map[string]interface{}{
			"s1": clientSeed,
		},
	}

	// Save to DB
	battleJSON, err := json.Marshal(newBattle)
	if err != nil {
		log.Println("failed to marshal battle:", err)
		return resR, errR
	}

	// Sanitize and build query
	query := fmt.Sprintf(
		`INSERT INTO g1_battles (server_seed,server_seed_hash, battle) 
				VALUES ('%s', '%s', '%s')`,
		serverSeed,
		serverSeedHash,
		string(battleJSON),
	)

	// gRPC Call Insert User
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "DB_DATA"
		errR.Code = 1070
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract inserted_id from nested struct
	newBattle.Status = fmt.Sprintf(`Waiting for %d users`, rune(slots-1))
	newBattle.StatusCode = 0

	dataDB := res.Data.GetFields()
	id := int(dataDB["inserted_id"].GetNumberValue())
	if id < 1 {
		errR.Type = "DB_DATA"
		errR.Code = 1070
		return resR, errR
	}
	newBattle.ID = id

	// Options : Private
	if utils.InArray(newBattle.Options, "private") {
		PrivateKey := GenerateShortBattleHash(strconv.Itoa(id))
		newBattle.PrivateKey = PrivateKey
	}

	// Winner Team
	switch newBattle.PlayerType {
	case "1v1", "1v1v1", "1v1v1v1", "1v6":
		var i = 0
		for key := range newBattle.Slots {
			newBattle.Teams = append(newBattle.Teams, models.Team{
				Slots: []string{key},
			})
			SetSlotTeam(newBattle, key, i)
			i++
		}
	case "2v2":
		newBattle.Teams = append(newBattle.Teams, models.Team{
			Slots: []string{"s1", "s2"},
		})
		SetSlotTeam(newBattle, "s1", 0)
		SetSlotTeam(newBattle, "s2", 0)

		newBattle.Teams = append(newBattle.Teams, models.Team{
			Slots: []string{"s3", "s4"},
		})
		SetSlotTeam(newBattle, "s3", 1)
		SetSlotTeam(newBattle, "s4", 1)

	case "2v2v2":
		newBattle.Teams = append(newBattle.Teams, models.Team{
			Slots: []string{"s1", "s2"},
		})
		SetSlotTeam(newBattle, "s1", 0)
		SetSlotTeam(newBattle, "s2", 0)

		newBattle.Teams = append(newBattle.Teams, models.Team{
			Slots: []string{"s3", "s4"},
		})
		SetSlotTeam(newBattle, "s3", 1)
		SetSlotTeam(newBattle, "s4", 1)

		newBattle.Teams = append(newBattle.Teams, models.Team{
			Slots: []string{"s5", "s6"},
		})
		SetSlotTeam(newBattle, "s5", 2)
		SetSlotTeam(newBattle, "s6", 2)

	case "3v3":
		newBattle.Teams = append(newBattle.Teams, models.Team{
			Slots: []string{"s1", "s2", "s3"},
		})
		SetSlotTeam(newBattle, "s1", 0)
		SetSlotTeam(newBattle, "s2", 0)
		SetSlotTeam(newBattle, "s3", 0)

		newBattle.Teams = append(newBattle.Teams, models.Team{
			Slots: []string{"s4", "s5", "s6"},
		})
		SetSlotTeam(newBattle, "s4", 1)
		SetSlotTeam(newBattle, "s5", 1)
		SetSlotTeam(newBattle, "s6", 1)
	}

	AddLog(newBattle, "create", int64(userID))

	var update, errV = UpdateBattle(newBattle)
	if update != true {
		return resR, errV
	}

	// Success
	resR.Type = "newBattle"
	resR.Data = newBattleResponse(BattleIndex[int64(id)])
	return resR, errR
}

// CancelBattle - Handler
func CancelBattle(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	// Get Battle
	battleId, vErr, ok := validate.RequireInt(data, "battleId")
	if !ok {
		return resR, vErr
	}
	battle, ok := GetBattle(battleId)
	if !ok {
		errR.Type = "NOT_FOUND"
		errR.Code = 5003
		return resR, errR
	}

	// Is Owner
	if userID != battle.CreatedBy {
		errR.Type = "INVALID_CREDENTIALS"
		errR.Code = 208
		return resR, errR
	}

	// Check Status
	if battle.StatusCode > 0 {
		errR.Type = "GAME_IS_LOCKED"
		errR.Code = 5007
		return resR, errR
	}

	// Check Player Count
	if len(battle.Players) > 1 {
		errR.Type = "GAME_IS_LOCKED"
		errR.Code = 5007
		return resR, errR
	}

	// update battle
	AddLog(battle, "cancelBattle", int64(userID))

	// Refound Process

	// Add Transaction
	Transaction, err := utils.AddTransaction(
		userID,
		"game_win",
		"1",
		battle.Cost,
		"",
		"Refound",
	)
	if err != nil {
		return resR, models.HandlerError{}
	}
	errCode, status, errType = utils.SafeExtractErrorStatus(Transaction)
	if status != 1 {
		errR.Type = errType
		errR.Code = errCode
		if resp["data"] != nil {
			errR.Data = resp["data"]
		}
		return resR, errR
	}

	battle.Status = fmt.Sprintf(`Canceled by user`)
	battle.StatusCode = -2
	var update, errV = UpdateBattle(battle)
	if update != true {
		return resR, errV
	}

	go dropBattle(battle.ID, 0)

	// Success
	resR.Type = "cancelBattle"
	resR.Data = map[string]interface{}{}
	return resR, errR
}

// Join - Handler
func Join(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	if DbBots == nil || len(DbBots.Values) == 0 {
		FillBots()
	}
	if len(CasesImpacted) == 0 {
		FillCaseImpact()
	}

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

	balanceStr := fmt.Sprintf("%v", profile["balance"])
	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		balance = 0
	}

	// Get Battle
	battleId, vErr, ok := validate.RequireInt(data, "battleId")
	if !ok {
		return resR, vErr
	}
	battle, ok := GetBattle(battleId)
	if !ok {
		errR.Type = "NOT_FOUND"
		errR.Code = 5003
		return resR, errR
	}

	// Options : Private
	if utils.InArray(battle.Options, "private") {
		privateKey, vErr, ok := validate.RequireString(data, "privateKey", false)
		if !ok {
			return resR, vErr
		}
		if privateKey != battle.PrivateKey {
			errR.Type = "GAME_IS_PRIVATE"
			errR.Code = 5008
			return resR, errR
		}
	}

	if battle.StatusCode > 0 {
		errR.Type = "GAME_IS_LOCKED"
		errR.Code = 5007
		return resR, errR
	}

	// Is Joined
	if IsPlayerInBattle(battle.Players, userID) {
		errR.Type = "ALREADY_JOINED"
		errR.Code = 5009
		return resR, errR
	}

	// Check Slot
	slotId, vErr, ok := validate.RequireInt(data, "slotId")
	if !ok {
		return resR, vErr
	}
	slotK := fmt.Sprintf("s%d", slotId)
	if battle.Slots[slotK].Type != "Empty" {
		errR.Type = "SLOT_IS_NOT_EMPTY"
		errR.Code = 1027
		return resR, errR
	}

	// Check Balance
	if balance < battle.Cost {
		errR.Type = "INSUFFICIENT_BALANCE"
		errR.Code = 7001
		errR.Data = map[string]interface{}{
			"cost":    battle.Cost,
			"balance": balance,
		}
		return resR, errR
	}
	// Add Transaction
	Transaction, err := utils.AddTransaction(
		userID,
		"game_loss",
		"1",
		battle.Cost,
		"",
		"Case Battle",
	)
	if err != nil {
		return resR, models.HandlerError{}
	}
	errCode, status, errType = utils.SafeExtractErrorStatus(Transaction)
	if status != 1 {
		errR.Type = errType
		errR.Code = errCode
		if resp["data"] != nil {
			errR.Data = resp["data"]
		}
		return resR, errR
	}

	// Join Battle
	clientSeed := utils.MD5UserID(userID)
	team := battle.Slots[slotK].Team
	battle.Slots[slotK] = models.Slot{
		ID:          userID,
		DisplayName: displayName,
		ClientSeed:  clientSeed,
		Type:        "Players",
		Team:        team,
	}
	battle.Players = append(battle.Players, userID)
	AddClientSeed(battle.PFair, slotK, clientSeed)

	// update battle
	AddLog(battle, "join", int64(userID))

	emptyCount := 0
	for _, slot := range battle.Slots {
		if slot.Type == "Empty" {
			emptyCount++
		}
	}
	if emptyCount == 0 {
		// Force To Roll
		battle.Status = "Start Rolling"
		battle.StatusCode = 0
		var update, errV = UpdateBattle(battle)
		if update != true {
			return resR, errV
		}
		go func(bid int) {
			Roll(int64(battle.ID), 0)
		}(battle.ID)
	} else {
		battle.Status = fmt.Sprintf(`Waiting for %d users`, emptyCount)
		battle.StatusCode = 0
		var update, errV = UpdateBattle(battle)
		if update != true {
			return resR, errV
		}
	}

	// Success
	resR.Type = "join"
	resR.Data = map[string]interface{}{
		"emptySlots": emptyCount,
		"clientSeed": clientSeed,
	}
	return resR, errR
}

// ChangeSeat - Handler
func ChangeSeat(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	if DbBots == nil || len(DbBots.Values) == 0 {
		FillBots()
	}
	if len(CasesImpacted) == 0 {
		FillCaseImpact()
	}

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

	// Get Battle
	battleId, vErr, ok := validate.RequireInt(data, "battleId")
	if !ok {
		return resR, vErr
	}
	battle, ok := GetBattle(battleId)
	if !ok {
		errR.Type = "NOT_FOUND"
		errR.Code = 5003
		return resR, errR
	}

	if battle.StatusCode > 0 {
		errR.Type = "GAME_IS_LOCKED"
		errR.Code = 5007
		return resR, errR
	}

	// Get Current Slot
	var oldSlot string
	for key, slot := range battle.Slots {
		if slot.ID == userID {
			oldSlot = key
			break
		}
	}
	log.Printf("Old slot key for userID %d : %s:", userID, oldSlot)

	// Check Slot
	slotId, vErr, ok := validate.RequireInt(data, "slotId")
	if !ok {
		return resR, vErr
	}
	slotK := fmt.Sprintf("s%d", slotId)
	if battle.Slots[slotK].Type != "Empty" {
		errR.Type = "SLOT_IS_NOT_EMPTY"
		errR.Code = 1027
		return resR, errR
	}
	log.Printf("Move to %s:", slotK)

	// Join New Slot
	clientSeed := utils.MD5UserID(userID)
	team := battle.Slots[slotK].Team
	battle.Slots[slotK] = models.Slot{
		ID:          userID,
		DisplayName: displayName,
		ClientSeed:  clientSeed,
		Type:        "Players",
		Team:        team,
	}
	AddClientSeed(battle.PFair, slotK, clientSeed)

	// Clear Old Slot
	battle.Slots[oldSlot] = models.Slot{
		ID:          0,
		DisplayName: "",
		ClientSeed:  "",
		Type:        "Empty",
	}
	RemoveClientSeed(battle.PFair, oldSlot)

	// update battle
	AddLog(battle, "changeSeat", int64(userID))

	emptyCount := 0
	for _, slot := range battle.Slots {
		if slot.Type == "Empty" {
			emptyCount++
		}
	}
	if emptyCount == 0 {
		// Force To Roll
		battle.Status = "Start Rolling"
		battle.StatusCode = 0
		var update, errV = UpdateBattle(battle)
		if update != true {
			return resR, errV
		}
		go func(bid int) {
			Roll(int64(battle.ID), 0)
		}(battle.ID)
	} else {
		battle.Status = fmt.Sprintf(`Waiting for %d users`, emptyCount)
		battle.StatusCode = 0
		var update, errV = UpdateBattle(battle)
		if update != true {
			return resR, errV
		}
	}

	// Success
	resR.Type = "changeSeat"
	resR.Data = map[string]interface{}{
		"emptySlots": emptyCount,
		"clientSeed": clientSeed,
	}
	return resR, errR
}

// GetBattleHistory - Handler
func GetBattleHistory(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	battleID, vErr, ok := validate.RequireInt(data, "battleId")
	if !ok {
		return resR, vErr
	}

	// Sanitize and build query
	query := fmt.Sprintf(
		`SELECT battle FROM g1_battles WHERE id = %d`,
		battleID,
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PROFILE_GRPC_ERROR"
		errR.Code = 1033
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()

	// DB result rows count
	exist := dataDB["count"].GetNumberValue()
	if exist == 0 {
		errR.Type = "Battle_NOT_FOUND"
		errR.Code = 1035
		return resR, errR
	}

	// Get rows
	rows := dataDB["rows"].GetListValue().Values
	if len(rows) == 0 {
		errR.Type = "Battle_NOT_FOUND"
		errR.Code = 1035
		return resR, errR
	}

	row := rows[0].GetStructValue()
	if row == nil {
		errR.Type = "BATTLE_ROW_EMPTY"
		errR.Code = 1038
		return resR, errR
	}

	fields := row.GetFields()

	battleVal := fields["battle"]
	battleStr := battleVal.GetStringValue()

	var battleMap models.Battle

	if strings.HasPrefix(battleStr, "{") {
		if err := json.Unmarshal([]byte(battleStr), &battleMap); err != nil {
			errR.Type = "BATTLE_JSON_ERROR"
			errR.Code = 1036
			return resR, errR
		}
	} else {
		unquoted, err := strconv.Unquote(battleStr)
		if err != nil {
			errR.Type = "BATTLE_JSON_DECODE_ERROR"
			errR.Code = 1037
			return resR, errR
		}

		if err := json.Unmarshal([]byte(unquoted), &battleMap); err != nil {
			errR.Type = "BATTLE_JSON_ERROR"
			errR.Code = 1036
			return resR, errR
		}
	}

	// @todo - remove some items

	// Success
	resR.Type = "getBattleHistory"
	resR.Data = ClientBattle(&battleMap)
	return resR, errR
}

// GetLiveBattles - Handler
func GetLiveBattles(_ map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Success
	resR.Type = "getLiveBattles"
	resR.Data = ClientBattleIndex(BattleIndex)
	return resR, errR
}

// GetBattleAdmin - Handler
func GetBattleAdmin(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := utils.ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	battleID, vErr, ok := validate.RequireInt(data, "battleId")
	if !ok {
		return resR, vErr
	}

	// Sanitize and build query
	query := fmt.Sprintf(
		`SELECT battle FROM g1_battles WHERE id = %d`,
		battleID,
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PROFILE_GRPC_ERROR"
		errR.Code = 1033
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()

	// DB result rows count
	exist := dataDB["count"].GetNumberValue()
	if exist == 0 {
		errR.Type = "Battle_NOT_FOUND"
		errR.Code = 1035
		return resR, errR
	}

	// Get rows
	rows := dataDB["rows"].GetListValue().Values
	if len(rows) == 0 {
		errR.Type = "Battle_NOT_FOUND"
		errR.Code = 1035
		return resR, errR
	}

	row := rows[0].GetStructValue()
	if row == nil {
		errR.Type = "BATTLE_ROW_EMPTY"
		errR.Code = 1038
		return resR, errR
	}

	fields := row.GetFields()

	battleVal := fields["battle"]
	battleStr := battleVal.GetStringValue()

	var battleMap map[string]interface{}

	if strings.HasPrefix(battleStr, "{") {
		if err := json.Unmarshal([]byte(battleStr), &battleMap); err != nil {
			errR.Type = "BATTLE_JSON_ERROR"
			errR.Code = 1036
			return resR, errR
		}
	} else {
		unquoted, err := strconv.Unquote(battleStr)
		if err != nil {
			errR.Type = "BATTLE_JSON_DECODE_ERROR"
			errR.Code = 1037
			return resR, errR
		}

		if err := json.Unmarshal([]byte(unquoted), &battleMap); err != nil {
			errR.Type = "BATTLE_JSON_ERROR"
			errR.Code = 1036
			return resR, errR
		}
	}

	// @todo - remove some items

	// Success
	resR.Type = "getBattleAdmin"
	resR.Data = battleMap
	return resR, errR
}

// GetLiveBattlesAdmin - Handler
func GetLiveBattlesAdmin(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := utils.ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Success
	resR.Type = "getLiveBattles"
	resR.Data = BattleIndex
	return resR, errR
}

// GetBattle - Safe Battle Actions
func GetBattle(id int64) (*models.Battle, bool) {
	battleIndexMu.RLock()
	defer battleIndexMu.RUnlock()
	b, ok := BattleIndex[id]
	return b, ok
}

// SetBattle - Safe Battle Actions
func SetBattle(id int64, b *models.Battle) {
	battleIndexMu.Lock()
	defer battleIndexMu.Unlock()
	BattleIndex[id] = b
}

// DeleteBattle - Safe Battle Actions
func DeleteBattle(id int64) {
	AddLog(BattleIndex[id], "archive", int64(0))

	battleIndexMu.Lock()
	defer battleIndexMu.Unlock()
	delete(BattleIndex, id)
}

// SetSlotTeam - Battle Helper
func SetSlotTeam(b *models.Battle, slotKey string, team int) {
	if slot, ok := b.Slots[slotKey]; ok {
		slot.Team = team
		b.Slots[slotKey] = slot
	}
}

// AddTeamPrizes - Battle Helper
func AddTeamPrizes(b *models.Battle, slotKey string, Prizes float64) {
	team := b.Slots[slotKey].Team
	b.Teams[team].TotalPrizes += Prizes
}

// AddTeamRollWin - Battle Helper
func AddTeamRollWin(b *models.Battle, slotKey string) {
	team := b.Slots[slotKey].Team
	b.Teams[team].RolWin++
}

// castStringSlice - Battle Helper
func castStringSlice(val interface{}) []string {
	var out []string
	if arr, ok := val.([]interface{}); ok {
		for _, v := range arr {
			out = append(out, fmt.Sprintf("%v", v))
		}
	}
	return out
}

// castCases - Battle Helper
func castCases(val interface{}) []map[string]int {
	var out []map[string]int
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

// UpdateBattle - Battle Helper
func UpdateBattle(battle *models.Battle) (bool, models.HandlerError) {
	var (
		errR models.HandlerError
		bID  = battle.ID
	)
	battle.UpdatedAt = time.Now()
	battleJSON, err := json.Marshal(battle)
	if err != nil {
		errR.Type = "json.Marshal(battle)"
		errR.Code = 1027
		return false, errR
	}
	// Sanitize and build query
	query := fmt.Sprintf(
		`Update g1_battles SET battle = '%s' WHERE id = %d`,
		string(battleJSON),
		bID,
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PROFILE_GRPC_ERROR"
		errR.Code = 1033
		if res != nil {
			errR.Data = res.Error
		}
		return false, errR
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()

	// DB result rows count
	exist := dataDB["rows_affected"].GetNumberValue()
	if exist == 0 {
		errR.Type = "USER_NOT_FOUND"
		errR.Code = 1035
		return false, errR
	}

	// Add To Battle Index
	SetBattle(int64(battle.ID), battle)

	return true, errR
}

// newBattleResponse - Battle Helper
func newBattleResponse(b *models.Battle) models.BattleCreated {
	slots := make(map[string]models.SlotResp)
	for k, v := range b.Slots {
		slots[k] = models.SlotResp{
			ID:          v.ID,
			DisplayName: v.DisplayName,
			Type:        v.Type,
		}
	}
	return models.BattleCreated{
		ID:         b.ID,
		PlayerType: b.PlayerType,
		Options:    b.Options,
		CaseCounts: b.CaseCounts,
		Cost:       b.Cost,
		Slots:      slots,
		Status:     b.Status,
		StatusCode: b.StatusCode,
		Summery:    b.Summery,
		CreatedAt:  b.CreatedAt,
		PrivateKey: b.PrivateKey,
	}
}

// FillBattleIndex - Battle Helper
func FillBattleIndex() (bool, models.HandlerError) {
	var (
		errR      models.HandlerError
		dbBattles *structpb.ListValue
	)

	log.Println("Fill BattleIndex..")

	// Sanitize and build query
	query := `SELECT battle FROM g1_battles WHERE is_live=1`

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PROFILE_GRPC_ERROR"
		errR.Code = 1033
		if res != nil {
			errR.Data = res.Error
		}
		return false, errR
	}
	// Extract gRPC struct
	dataDB := res.Data.GetFields()
	// DB result rows count
	exist := dataDB["count"].GetNumberValue()
	if exist == 0 {
		errR.Type = "DB_DATA"
		errR.Code = 1070
		return false, errR
	}
	// DB result rows get fields
	dbBattles = dataDB["rows"].GetListValue()

	for idx, row := range dbBattles.Values {
		structRow := row.GetStructValue()
		battleJSON := structRow.Fields["battle"].GetStringValue() // JSON string

		var b models.Battle
		err := json.Unmarshal([]byte(battleJSON), &b)
		if err != nil {
			log.Println("Failed to unmarshal battle:", err)
			continue
		}

		key := int64(b.ID)
		if key == 0 {
			key = int64(idx + 1)
		}

		BattleIndex[key] = &b
	}

	return true, errR
}

// IsPlayerInBattle - Battle Helper
func IsPlayerInBattle(players []int, userID int) bool {
	for _, id := range players {
		if id == userID {
			return true
		}
	}
	return false
}

// AddClientSeed - Battle Helper
func AddClientSeed(battle map[string]interface{}, key string, value interface{}) {
	cs, ok := battle["clientSeed"].(map[string]interface{})
	if !ok {
		cs = make(map[string]interface{})
		battle["clientSeed"] = cs
	}
	cs[key] = value
}

// AddLog - Battle Helper
func AddLog(b *models.Battle, action string, userID int64) {
	b.Logs = append(b.Logs, models.BattleLog{
		Time:   time.Now().UTC().Format(time.RFC3339),
		Action: action,
		UserID: userID,
	})
}

// RemoveClientSeed - Battle Helper
func RemoveClientSeed(battle map[string]interface{}, key string) {
	cs, ok := battle["clientSeed"].(map[string]interface{})
	if !ok {
		return
	}
	delete(cs, key)
}

// ClientBattleIndex - Battle Helper
func ClientBattleIndex(battles map[int64]*models.Battle) map[int64]models.BattleClient {
	out := make(map[int64]models.BattleClient)
	for _, b := range battles {
		dto := models.BattleClient{
			ID:             b.ID,
			PlayerType:     b.PlayerType,
			Options:        b.Options,
			Cases:          b.Cases,
			CasesUi:        b.CasesUi,
			CaseCounts:     b.CaseCounts,
			Cost:           b.Cost,
			Slots:          b.Slots,
			Status:         b.Status,
			StatusCode:     b.StatusCode,
			Summery:        b.Summery,
			CreatedAt:      b.CreatedAt,
			CreatedBy:      b.CreatedBy,
			UpdatedAt:      b.UpdatedAt,
			ServerSeedHash: b.PFair["serverSeedHash"].(string),
		}
		out[int64(b.ID)] = dto
	}
	return out
}

// ClientBattle - Battle Helper
func ClientBattle(b *models.Battle) models.BattleClient {
	return models.BattleClient{
		ID:             b.ID,
		PlayerType:     b.PlayerType,
		Options:        b.Options,
		Cases:          b.Cases,
		CasesUi:        b.CasesUi,
		CaseCounts:     b.CaseCounts,
		Cost:           b.Cost,
		Slots:          b.Slots,
		Status:         b.Status,
		StatusCode:     b.StatusCode,
		Summery:        b.Summery,
		CreatedAt:      b.CreatedAt,
		UpdatedAt:      b.UpdatedAt,
		ServerSeedHash: b.PFair["serverSeedHash"].(string),
	}
}

// expandCases - Battle Helper
func expandCases(input []map[string]int) []int {
	var out []int
	for _, m := range input {
		for k, count := range m {
			caseID, _ := strconv.Atoi(k)
			for i := 0; i < count; i++ {
				out = append(out, caseID)
			}
		}
	}
	return out
}

// GenerateShortBattleHash - Battle Helper
func GenerateShortBattleHash(battleID string) string {
	secretKey := []byte(os.Getenv("HMAC_SECRET"))
	timestamp := time.Now().Unix()
	message := fmt.Sprintf("%s:%d", battleID, timestamp)
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(message))
	fullHash := h.Sum(nil)
	shortHash := hex.EncodeToString(fullHash[:8])
	return shortHash
}

// Roll - Battle Helper
func Roll(battleID int64, roundKey int) {
	time.Sleep(3 * time.Second)

	if DbBots == nil || len(DbBots.Values) == 0 {
		FillBots()
	}
	if len(CasesImpacted) == 0 {
		FillCaseImpact()
	}

	battle, ok := GetBattle(battleID)
	if !ok {
		log.Println("Battle not found:", battleID)
		return
	}

	// Check if roll has already done
	if steps, exists := battle.Summery.Steps[roundKey]; exists && len(steps) > 0 {
		if configs.Debug {
			log.Printf("Info: Round %d has already been rolled", roundKey)
		}
	} else {

		// Count Last Roll Percentages
		if roundKey > 0 {

			var (
				parts []float64
				total float64
			)

			lastStep := battle.Summery.Steps[roundKey-1]
			if lastStep != nil {

				for _, slot := range lastStep {
					total += slot.Price
					parts = append(parts, slot.Price)
				}
				result := utils.CalculatePercentages(parts, total)

				for j := range lastStep {
					battle.Summery.Steps[roundKey-1][j].Percentage = result[j]
				}

			}

		}

		// Last Roll
		if roundKey >= len(battle.Cases) {
			battle.Status = fmt.Sprintf("Rolled")
			battle.StatusCode = 1

			if utils.InArray(battle.Options, "jackpot") {
				// Ensure Jackpot map is initialized
				if battle.Summery.Jackpot == nil {
					battle.Summery.Jackpot = make(map[string]float64)
				}

				// Calculate total
				var total float64
				for _, value := range battle.Summery.Prizes {
					total += value
				}

				// Avoid division by zero
				if total == 0 {
					return
				}

				// Fill jackpot percentages
				for key, value := range battle.Summery.Prizes {
					battle.Summery.Jackpot[key] = utils.RoundToTwoDigits((value / total) * 100)
				}
			}

			// Move to Option Level
			if configs.Debug {
				log.Printf("Battle %d steps(%d) are done.", battleID, roundKey)
			}
			UpdateBattle(battle)
			events.Emit("all", "heartbeat", ClientBattleIndex(BattleIndex))
			// Go to check Options
			optionActions(battleID)
			return
		}

		// Run Roll
		battle.Status = fmt.Sprintf("Roll %d", roundKey+1)
		battle.StatusCode = 0
		if battle.Summery.Steps == nil {
			battle.Summery.Steps = make(map[int][]models.StepResult)
		}
		if battle.Summery.Prizes == nil {
			battle.Summery.Prizes = make(map[string]float64)
		}
		nonce := ((roundKey + 7) * 2) + roundKey
		caseID := battle.Cases[roundKey]
		caseData := CasesImpacted[caseID]
		var (
			rollWinner string
			lastPrize  float64
		)
		lastPrize = 0
		for slot := range battle.Slots {
			clientSeed, ok := battle.PFair["clientSeed"].(map[string]interface{})[slot].(string)
			if !ok {
				log.Println("No clientSeed for slot:", slot)
				continue
			}

			nonce += 97
			item := provablyfair.PickItem(
				caseData,
				battle.PFair["serverSeed"].(string),
				clientSeed,
				nonce,
			)
			if configs.Debug {
				log.Println("Roll "+strconv.Itoa(roundKey), slot, caseID, nonce, item["price"])
			}

			if item == nil {
				log.Println("No item picked for slot:", slot)
				continue
			}

			priceStr, _ := item["price"].(string)
			price, _ := strconv.ParseFloat(priceStr, 64)

			step := models.StepResult{
				Slot:   slot,
				ItemID: int(item["id"].(float64)),
				Price:  utils.RoundToTwoDigits(price),
			}

			// reRun
			if roundKey == len(battle.Cases)-1 {

				for item["price"] == strconv.FormatFloat(lastPrize, 'f', -1, 64) {
					nonce += 97
					item = provablyfair.PickItem(
						caseData,
						battle.PFair["serverSeed"].(string),
						clientSeed,
						nonce,
					)
					if configs.Debug {
						log.Println("Roll "+strconv.Itoa(roundKey), slot, caseID, nonce, item["price"])
					}

					if item == nil {
						log.Println("No item picked for slot:", slot)
						continue
					}

					priceStr, _ = item["price"].(string)
					price, _ = strconv.ParseFloat(priceStr, 64)

					step = models.StepResult{
						Slot:   slot,
						ItemID: int(item["id"].(float64)),
						Price:  utils.RoundToTwoDigits(price),
					}
				}

			}

			battle.Summery.Steps[roundKey] = append(battle.Summery.Steps[roundKey], step)
			battle.Summery.Prizes[slot] += step.Price
			AddTeamPrizes(battle, slot, step.Price)

			if lastPrize < step.Price {
				rollWinner = slot
			}
			lastPrize = step.Price
		}
		AddTeamRollWin(battle, rollWinner)
		AddLog(battle, fmt.Sprintf("Roll %d", roundKey+1), 0)
	}
	Roll(battleID, roundKey+1)
}

// optionActions - Battle Helper
func optionActions(battleID int64) {
	battle, ok := GetBattle(battleID)
	if !ok {
		log.Println("Battle not found:", battleID)
		return
	}

	// Wait for animations
	time.Sleep(time.Duration(6*battle.CaseCounts) * time.Second)

	// Winner Team
	winner := battle.Teams[0]
	if len(battle.Options) == 0 {
		// No Options
		for _, t := range battle.Teams {
			if t.TotalPrizes > winner.TotalPrizes {
				winner = t
			}
		}
	} else {
		// Handel Options
		if utils.InArray(battle.Options, "equality") {
			for _, t := range battle.Teams {
				if t.TotalPrizes > winner.TotalPrizes {
					winner = t
				}
			}
			keys := make([]string, 0, len(battle.Summery.Prizes))
			for k := range battle.Summery.Prizes {
				keys = append(keys, k)
			}
			winner.Slots = keys
		} else {
			if utils.InArray(battle.Options, "madness") {
				for _, t := range battle.Teams {
					if t.TotalPrizes < winner.TotalPrizes {
						winner = t
					}
				}
			}
			if utils.InArray(battle.Options, "jackpot") && !utils.InArray(battle.Options, "madness") {
				winnerSlot := weightedRandom(battle.Summery.Jackpot, false)
				battle.Summery.JackpotWinner = winnerSlot
				teamID := battle.Slots[winnerSlot].Team
				winner = battle.Teams[teamID]
			}
			if utils.InArray(battle.Options, "jackpot") && utils.InArray(battle.Options, "madness") {
				winnerSlot := weightedRandom(battle.Summery.Jackpot, true)
				battle.Summery.JackpotWinner = winnerSlot
				teamID := battle.Slots[winnerSlot].Team
				winner = battle.Teams[teamID]
			}
		}
	}

	// Get Total Prize
	var total float64
	for _, v := range battle.Summery.Prizes {
		total += v
	}
	battle.Summery.Winners = winner
	battle.Summery.Winners.TotalPrizes = utils.RoundToTwoDigits(total)
	battle.Summery.Winners.SlotPrizes = utils.RoundToTwoDigits(total / float64(len(battle.Summery.Winners.Slots)))

	battle.Status = "Resolving"
	battle.StatusCode = 2

	AddLog(battle, "Handel Options", 0)
	UpdateBattle(battle)

	// Emit | heartbeat
	events.Emit("all", "heartbeat", ClientBattleIndex(BattleIndex))

	// Archive battle
	archive(battle.ID)
	return
}

// archive - Battle Helper
func archive(battleID int) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	battle, ok := GetBattle(int64(battleID))
	if !ok {
		log.Println("Battle not found:", battleID)
		return resR, models.HandlerError{}
	}

	for _, v := range battle.Summery.Winners.Slots {
		userID := battle.Slots[v].ID

		// Skip Empty / Bot
		if battle.Slots[v].Type != "Player" {
			continue
		}

		// Get Users
		resp, err := utils.GetUser(userID)
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

		// Add Transaction
		Transaction, err := utils.AddTransaction(
			userID,
			"game_win",
			strconv.Itoa(battleID),
			battle.Summery.Winners.SlotPrizes,
			"",
			"Case Battle",
		)
		if err != nil {
			return resR, models.HandlerError{}
		}
		errCode, status, errType = utils.SafeExtractErrorStatus(Transaction)
		if status != 1 {
			errR.Type = errType
			errR.Code = errCode
			if resp["data"] != nil {
				errR.Data = resp["data"]
			}
			return resR, errR
		}

		// Send Live Winner
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered in sendLiveWinner: %v\n", r)
				}
			}()
			ok := sendLiveWinner(
				battle.Slots[v].DisplayName,
				fmt.Sprintf("%.2f", battle.Cost),
				"",
				fmt.Sprintf("%.2f", battle.Summery.Winners.SlotPrizes),
			)
			if ok == false {
				log.Printf("sendLiveWinner error")
			}
		}()

		UpdateBattle(battle)
	}

	battle.Status = "Rewarding"
	battle.StatusCode = 3
	UpdateBattle(battle)

	// Sanitize and build query
	query := fmt.Sprintf(
		`Update g1_battles SET is_live = 0 WHERE id = %d`,
		battle.ID,
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PROFILE_GRPC_ERROR"
		errR.Code = 1033
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()

	// DB result rows count
	exist := dataDB["rows_affected"].GetNumberValue()
	if exist == 0 {
		errR.Type = "USER_NOT_FOUND"
		errR.Code = 1035
		return resR, errR
	}

	// Emit | heartbeat
	events.Emit("all", "heartbeat", ClientBattleIndex(BattleIndex))

	// Wait for animation
	time.Sleep(600 * time.Second)

	battle.Status = "Archived"
	battle.StatusCode = -1
	UpdateBattle(battle)

	// Drop Battle from Index
	go dropBattle(battle.ID, 0)

	// Emit | heartbeat
	events.Emit("all", "heartbeat", ClientBattleIndex(BattleIndex))

	return resR, models.HandlerError{}
}

// dropBattle - Battle Helper
func dropBattle(battleId int, after int) {
	time.Sleep(time.Duration(after) * time.Second)
	battle, ok := GetBattle(int64(battleId))
	if ok {
		DeleteBattle(int64(battle.ID))
	}
}

// weightedRandom - Jackpot Helper
// selects a key from the given prize map based on weighted probability.
// If inverse is true, the weights are inverted (1/weight).
func weightedRandom(prizes map[string]float64, inverse bool) string {
	var total float64
	weights := make(map[string]float64)

	// Calculate weights and total
	if inverse {
		for key, weight := range prizes {
			if weight == 0 {
				weights[key] = 0
			} else {
				weights[key] = 1 / weight
			}
			total += weights[key]
		}
	} else {
		for key, weight := range prizes {
			weights[key] = weight
			total += weight
		}
	}

	// Generate a random value within the total range
	random := rand.Float64() * total

	// Iterate over weights and find the selected key
	var cumulative float64
	for key, weight := range weights {
		cumulative += weight
		if random <= cumulative {
			return key
		}
	}

	// Fallback: return the first available key
	for key := range prizes {
		return key
	}

	return ""
}

func sendLiveWinner(displayName string, bet string, multiplier string, payout string) bool {
	apiAppErr := apiapp.InsertWinner(
		1,
		time.Now(),
		displayName,
		bet,
		multiplier,
		payout,
	)
	if apiAppErr != nil {
		log.Println("Error:", apiAppErr)
		return false
	}
	return true
}
