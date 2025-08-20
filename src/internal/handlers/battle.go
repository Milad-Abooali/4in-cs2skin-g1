package handlers

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/configs"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/provablyfair"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/validate"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/utils"
	"google.golang.org/protobuf/types/known/structpb"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	BattleIndex    = make(map[int64]*models.Battle)
	BattleIndexOut = make(map[int64]*models.BattleClient)
	battleIndexMu  sync.RWMutex
)

func GetBattle(id int64) (*models.Battle, bool) {
	battleIndexMu.RLock()
	defer battleIndexMu.RUnlock()
	b, ok := BattleIndex[id]
	return b, ok
}

func SetBattle(id int64, b *models.Battle) {
	battleIndexMu.Lock()
	defer battleIndexMu.Unlock()
	BattleIndex[id] = b
}

func DeleteBattle(id int64) {
	AddLog(BattleIndex[id], "archive", int64(0))

	battleIndexMu.Lock()
	defer battleIndexMu.Unlock()
	delete(BattleIndex, id)
}

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
		Options:    ToLowerArray(options),
		Cases:      expandCases(castCases(data["cases"])),
		Players:    []int{},
		CreatedBy:  0,
		Status:     "initialized",
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
								newBattle.Cost += price

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
	clientSeed := MD5UserID(userID)
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

	// Winner Team
	switch newBattle.PlayerType {
	case "1v1", "1v1v1", "1v1v1v1", "1v6":
		var i int = 0
		for key, _ := range newBattle.Slots {
			i++
			newBattle.WinTeams = append(newBattle.WinTeams, models.WinTeam{
				Slots: []string{key},
			})
			SetSlotTeam(newBattle, key, i)
		}
	case "2v2":
		newBattle.WinTeams = append(newBattle.WinTeams, models.WinTeam{
			Slots: []string{"s1", "s2"},
		})
		SetSlotTeam(newBattle, "s1", 1)
		SetSlotTeam(newBattle, "s2", 1)

		newBattle.WinTeams = append(newBattle.WinTeams, models.WinTeam{
			Slots: []string{"s3", "s4"},
		})
		SetSlotTeam(newBattle, "s3", 2)
		SetSlotTeam(newBattle, "s4", 2)

	case "2v2v2":
		newBattle.WinTeams = append(newBattle.WinTeams, models.WinTeam{
			Slots: []string{"s1", "s2"},
		})
		SetSlotTeam(newBattle, "s1", 1)
		SetSlotTeam(newBattle, "s2", 1)

		newBattle.WinTeams = append(newBattle.WinTeams, models.WinTeam{
			Slots: []string{"s3", "s4"},
		})
		SetSlotTeam(newBattle, "s3", 2)
		SetSlotTeam(newBattle, "s4", 2)

		newBattle.WinTeams = append(newBattle.WinTeams, models.WinTeam{
			Slots: []string{"s5", "s6"},
		})
		SetSlotTeam(newBattle, "s5", 3)
		SetSlotTeam(newBattle, "s6", 3)

	case "3v3":
		newBattle.WinTeams = append(newBattle.WinTeams, models.WinTeam{
			Slots: []string{"s1", "s2", "s3"},
		})
		SetSlotTeam(newBattle, "s1", 1)
		SetSlotTeam(newBattle, "s2", 1)
		SetSlotTeam(newBattle, "s3", 1)

		newBattle.WinTeams = append(newBattle.WinTeams, models.WinTeam{
			Slots: []string{"s4", "s5", "s6"},
		})
		SetSlotTeam(newBattle, "s4", 2)
		SetSlotTeam(newBattle, "s5", 2)
		SetSlotTeam(newBattle, "s6", 2)
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
	if inArray(newBattle.Options, "private") {
		PrivateKey := GenerateShortBattleHash(strconv.Itoa(id))
		newBattle.PrivateKey = PrivateKey
	}

	AddLog(newBattle, "create", int64(userID))

	var update, errV = UpdateBattle(newBattle)
	if update != true {
		return resR, errV
	}

	// Success
	resR.Type = "newBattle"
	resR.Data = ToBattleResponse(BattleIndex[int64(id)])
	return resR, errR
}

func inArray[T comparable](arr []T, item T) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}
	return false
}

func SetSlotTeam(b *models.Battle, slotKey string, team int) {
	if slot, ok := b.Slots[slotKey]; ok {
		slot.Team = team
		b.Slots[slotKey] = slot
	}
}

func ToLowerArray(arr []string) []string {
	lowerArr := make([]string, len(arr))
	for i, v := range arr {
		lowerArr[i] = strings.ToLower(v)
	}
	return lowerArr
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

func UpdateBattle(battle *models.Battle) (bool, models.HandlerError) {
	var (
		errR models.HandlerError
		bID  int = battle.ID
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

func ToBattleResponse(b *models.Battle) models.BattleCreated {
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
		Summery:    b.Summery,
		CreatedAt:  b.CreatedAt,
		PrivateKey: b.PrivateKey,
	}
}

func MD5UserID(userID int) string {
	data := []byte(fmt.Sprintf("%d", userID))
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

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
	if inArray(battle.Options, "private") {
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
		errR.Code = 1017
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
	clientSeed := MD5UserID(userID)
	battle.Slots[slotK] = models.Slot{
		ID:          userID,
		DisplayName: displayName,
		ClientSeed:  clientSeed,
		Type:        "Players",
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
		battle.Status = "Battle is running ..."
		Roll(int64(battle.ID), 0)
	} else {
		battle.Status = fmt.Sprintf(`Waiting for %d users`, emptyCount)
	}
	var update, errV = UpdateBattle(battle)
	if update != true {
		return resR, errV
	}

	// Success
	resR.Type = "join"
	resR.Data = map[string]interface{}{
		"emptySlots": emptyCount,
		"clientSeed": clientSeed,
	}
	return resR, errR
}

func IsPlayerInBattle(players []int, userID int) bool {
	for _, id := range players {
		if id == userID {
			return true
		}
	}
	return false
}

func AddClientSeed(battle map[string]interface{}, key string, value interface{}) {
	cs, ok := battle["clientSeed"].(map[string]interface{})
	if !ok {
		cs = make(map[string]interface{})
		battle["clientSeed"] = cs
	}
	cs[key] = value
}

func AddLog(b *models.Battle, action string, userID int64) {
	b.Logs = append(b.Logs, models.BattleLog{
		Time:   time.Now().UTC().Format(time.RFC3339),
		Action: action,
		UserID: userID,
	})
}

func RemoveClientSeed(battle map[string]interface{}, key string) {
	cs, ok := battle["clientSeed"].(map[string]interface{})
	if !ok {
		return
	}
	delete(cs, key)
}

func BuildBattleIndex(battles map[int64]*models.Battle) map[int64]models.BattleClient {
	out := make(map[int64]models.BattleClient)
	for _, b := range battles {
		dto := models.BattleClient{
			ID:             b.ID,
			PlayerType:     b.PlayerType,
			Options:        b.Options,
			Cases:          b.Cases,
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
		out[int64(b.ID)] = dto
	}
	return out
}

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

func Roll(battleID int64, roundKey int) {

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
		log.Printf("Error: Round %d has already been rolled", roundKey)
	} else {

		// Wait for animation
		time.Sleep(3 * time.Second)

		// Check max roll
		if roundKey < 0 || roundKey >= len(battle.Cases) {
			// Move to Option Level
			log.Printf("Error: Round %d has already been rolled", roundKey)
			if configs.Debug {
				log.Printf("Battle %d steps(%d) are done.", battleID, roundKey)
			}
			// Go to check Options
			optionActions(battleID)
			return
		}

		battle.Status = fmt.Sprintf("Roll %d", roundKey+1)
		battle.StatusCode = 1

		if battle.Summery.Steps == nil {
			battle.Summery.Steps = make(map[int][]models.StepResult)
		}
		if battle.Summery.Prizes == nil {
			battle.Summery.Prizes = make(map[string]float64)
		}

		nonce := ((roundKey + 7) * 2) + roundKey
		caseID := battle.Cases[roundKey]
		caseData := CasesImpacted[caseID]

		for slot, _ := range battle.Slots {
			clientSeed, ok := battle.PFair["clientSeed"].(map[string]interface{})[slot].(string)
			if !ok {
				log.Println("No clientSeed for slot:", slot)
				continue
			}
			nonce++

			item := provablyfair.PickItem(
				caseData,
				battle.PFair["serverSeed"].(string),
				clientSeed,
				nonce,
			)

			if configs.Debug {
				log.Println("Roll", slot, caseID, nonce, item["price"])
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
				Price:  price,
			}

			battle.Summery.Steps[roundKey] = append(battle.Summery.Steps[roundKey], step)
			battle.Summery.Prizes[slot] += step.Price

		}

		AddLog(battle, fmt.Sprintf("Roll %d", roundKey+1), 0)
		UpdateBattle(battle)
	}

	Roll(battleID, roundKey+1)
}

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

func optionActions(battleID int64) {
	battle, ok := GetBattle(battleID)
	if !ok {
		log.Println("Battle not found:", battleID)
		return
	}

	// Winner Team

	if len(battle.Options) == 0 {
		// No Options - Default

	} else {
		// Handel Options
		executedOptions := make(map[string]bool)
		for _, o := range battle.Options {
			opt := strings.ToLower(o)
			if executedOptions[opt] {
				continue
			}
			log.Println(opt)

			switch opt {
			case "fast spin":
				// No Action

			case "Private":
				// No Action

			case "Madness":
				// Winner Based on count of rolls.

			case "Jackpot":
				// Winner Based on count of win on rolls.

			case "Equality":
				// Div win prize to all slots

			}
			executedOptions[opt] = true
		}
	}

	archive()
	return
}

func archive() {

}

func Test(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	Roll(battleId, 0)

	// Success
	resR.Type = "test"
	resR.Data = battle
	return resR, errR
}
