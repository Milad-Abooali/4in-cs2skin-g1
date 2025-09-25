package he

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/utils"
	"strconv"
)

func GetAvgHE(gameTable string, limit int) (float64, bool) {
	query := fmt.Sprintf(
		`SELECT AVG(cb.he) AS avg_he FROM %s cb JOIN ( SELECT id FROM %s WHERE is_live = 0 AND income > 0 ORDER BY created_at DESC LIMIT %d ) ids ON cb.id = ids.id`,
		gameTable,
		gameTable,
		limit,
	)
	res, err := grpcclient.SendQuery(query)

	if err != nil || res == nil || res.Status != "ok" {
		return 0, false
	}
	dataDB := res.Data.GetFields()
	exist := dataDB["count"].GetNumberValue()
	if exist == 0 {
		return 0, false
	}
	sHE := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()["avg_he"].GetStringValue()
	HE, _ := strconv.ParseFloat(sHE, 64)
	return utils.RoundToTwoDigits(HE), true
}
