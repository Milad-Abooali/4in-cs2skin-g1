package errorsreg

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
)

//go:embed errors.json
var errorsJSON []byte

type Entry struct {
	Code   int     `json:"code"`
	HTTP   int     `json:"http"`
	Key    *string `json:"key"`    // nullable
	Detail any     `json:"detail"` // nullable
	Text   string  `json:"text"`
}

var byCode map[int]Entry

func init() {
	var list []Entry
	if err := json.Unmarshal(errorsJSON, &list); err != nil {
		log.Fatalf("load errors.json: %v", err)
	}
	byCode = make(map[int]Entry, len(list))
	for _, e := range list {
		byCode[e.Code] = e
	}
}

func HTTPStatus(code int) int {
	if code == 0 {
		return http.StatusOK
	}
	if e, ok := byCode[code]; ok && e.HTTP > 0 {
		return e.HTTP
	}
	return http.StatusInternalServerError
}

func Lookup(code int) (Entry, bool) {
	e, ok := byCode[code]
	return e, ok
}
