package handlers

import (
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"time"
)

func Ping(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Success
	resR.Type = "ping"
	resR.Data = time.Now().UTC().Format(time.RFC3339)
	return resR, errR
}
