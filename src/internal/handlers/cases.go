package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"google.golang.org/protobuf/types/known/structpb"
	"log"
)

var (
	DbCases       *structpb.ListValue
	DbCaseItems   *structpb.ListValue
	CasesImpacted map[int]grpcclient.CaseWithItems
)

// GetCases - Handler
func GetCases(_ map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	if len(CasesImpacted) == 0 {
		FillCaseImpact()
	}

	// Success
	resR.Type = "getCases"
	resR.Data = CasesImpacted
	return resR, errR
}

// UpdateCases - Handler
func UpdateCases(_ map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	FillCaseImpact()

	// Success
	resR.Type = "updateCases"
	resR.Data = CasesImpacted
	return resR, errR
}

// FillCaseImpact - Helper
func FillCaseImpact() (map[int]grpcclient.CaseWithItems, models.HandlerError) {
	log.Println("Fill CasesImpacted...")
	var (
		errR models.HandlerError
	)
	// Sanitize and build query
	query := fmt.Sprintf(`SELECT id,name,color,price,distribution,rarity,weight FROM cases WHERE publish_status=1`)
	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "GRPC_ERROR"
		errR.Code = 1033
		if res != nil {
			errR.Data = res.Error
		}
		return CasesImpacted, errR
	}
	// Extract gRPC struct
	dataDB := res.Data.GetFields()
	// DB result rows count
	exist := dataDB["count"].GetNumberValue()
	if exist == 0 {
		errR.Type = "DB_DATA"
		errR.Code = 1070
		return CasesImpacted, errR
	}
	// DB result rows get fields
	DbCases = dataDB["rows"].GetListValue()

	// Sanitize and build query
	query = fmt.Sprintf(`SELECT 
    	ci.id,
    	ci.case_id,
    	ci.item_id,
   		ci.min_rand,
   		ci.max_rand,
   		ci.price,
   		ci.rarity,
   		ci.color,
   		ir.market_hash_name,
        "Emerlad" as item_cat FROM case_items ci LEFT JOIN item_repo ir ON ci.item_id = ir.id`)
	// gRPC Call
	res, err = grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "GRPC_ERROR"
		errR.Code = 1033
		if res != nil {
			errR.Data = res.Error
		}
		return CasesImpacted, errR
	}
	// Extract gRPC struct
	dataDB = res.Data.GetFields()
	// DB result rows count
	exist = dataDB["count"].GetNumberValue()
	if exist == 0 {
		errR.Type = "DB_DATA"
		errR.Code = 1070
		return CasesImpacted, errR
	}
	// DB result rows get fields
	DbCaseItems = dataDB["rows"].GetListValue()

	// Merge Data
	CasesImpacted = grpcclient.MergeCasesAndItems(grpcclient.ListValueToStructs(DbCases), grpcclient.ListValueToStructs(DbCaseItems))

	return CasesImpacted, errR
}
