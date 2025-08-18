package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/validate"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/utils"
	"google.golang.org/protobuf/types/known/structpb"
	"log"
	"math/rand"
	"time"
)

var (
	DbBots *structpb.ListValue
)

func GetBots(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	if DbBots == nil || len(DbBots.Values) == 0 {
		FillBots()
	}

	// Success
	resR.Type = "getBots"
	resR.Data = DbBots
	return resR, errR
}

func AddBot(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	if DbBots == nil || len(DbBots.Values) == 0 {
		FillBots()
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
	log.Println(userID)

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

	// Select a bot
	bot := randomBot(DbBots)
	botId := int(bot.GetStructValue().Fields["id"].GetNumberValue())
	botName := bot.GetStructValue().Fields["name"].GetStringValue()
	clientSeed := MD5UserID(botId)

	// Join Battle
	battle.Slots[slotK] = models.Slot{
		ID:          botId,
		DisplayName: botName,
		ClientSeed:  clientSeed,
		Type:        "Bot",
	}
	battle.Bots = append(battle.Bots, botId)

	// update battle
	var update, errV = UpdateBattle(battle)
	if update != true {
		return resR, errV
	}

	// Success
	resR.Type = "addBot"
	resR.Data = bot
	return resR, errR
}

func FillBots() (*structpb.ListValue, models.HandlerError) {
	log.Println("Fill DbBots...")
	var (
		errR models.HandlerError
	)
	// Sanitize and build query
	query := fmt.Sprintf(`SELECT * FROM bots`)
	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PROFILE_GRPC_ERROR"
		errR.Code = 1033
		if res != nil {
			errR.Data = res.Error
		}
		return DbBots, errR
	}
	// Extract gRPC struct
	dataDB := res.Data.GetFields()
	// DB result rows count
	exist := dataDB["count"].GetNumberValue()
	if exist == 0 {
		errR.Type = "DB_DATA"
		errR.Code = 1070
		return DbBots, errR
	}
	// DB result rows get fields
	DbBots = dataDB["rows"].GetListValue()

	return DbBots, errR
}

func randomBot(DbBots *structpb.ListValue) *structpb.Value {
	if DbBots == nil || len(DbBots.Values) == 0 {
		return nil
	}

	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(len(DbBots.Values))
	return DbBots.Values[idx]
}
