package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"google.golang.org/protobuf/types/known/structpb"
	"log"
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
		log.Println("Fill DbBots...")

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
			return resR, errR
		}
		// Extract gRPC struct
		dataDB := res.Data.GetFields()
		// DB result rows count
		exist := dataDB["count"].GetNumberValue()
		if exist == 0 {
			errR.Type = "DB_DATA"
			errR.Code = 1070
			return resR, errR
		}
		// DB result rows get fields
		DbBots = dataDB["rows"].GetListValue()
	}

	// Success
	resR.Type = "getBots"
	resR.Data = DbBots
	return resR, errR
}
