package handlers

import (
	"encoding/json"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/configs"
	errorsreg "github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/errors"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/events"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
)

func SendWSResponse(conn *websocket.Conn, reqId int64, resType string, data interface{}) {
	resp := models.ReqRes{
		ReqID:  reqId,
		Type:   resType,
		Status: 1,
		Data:   data,
	}
	err := conn.WriteJSON(resp)
	if err != nil {
		return
	}
}
func SendWSError(conn *websocket.Conn, reqId int64, resType string, eCode int, eExtra ...any) {
	if configs.Debug {
		log.Printf("Error %d | %s", eCode, resType)
	}
	resp := models.ReqRes{
		ReqID:  reqId,
		Type:   resType,
		Status: 0,
		Error:  eCode,
		Data:   eExtra,
	}
	err := conn.WriteJSON(resp)
	if err != nil {
		return
	}
}

func SendWebResponse(w http.ResponseWriter, resType string, data interface{}) {
	resp := models.ReqRes{
		Type:   resType,
		Status: 1,
		Data:   data,
	}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}
func SendWebError(w http.ResponseWriter, resType string, eCode int, eExtra ...any) {
	if configs.Debug {
		log.Printf("Error %d | %s", eCode, resType)
	}
	resp := models.ReqRes{
		Type:   resType,
		Status: 0,
		Error:  eCode,
		Data:   eExtra,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	if e, ok := errorsreg.Lookup(eCode); ok && e.Key != nil {
		w.Header().Set("X-Error-Key", *e.Key)
	}
	w.Header().Set("X-Error-Code", strconv.Itoa(eCode))
	w.WriteHeader(errorsreg.HTTPStatus(eCode))

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

func EmitServer(resType string) {
	switch resType {
	case "test", "getBots", "getCases", "getBattleHistory":
		// no emit
	default:
		events.Bus <- events.Event{
			Target: "all",
			Type:   "heartbeat",
			Data:   ClientBattleIndex(BattleIndex),
		}
	}
}
