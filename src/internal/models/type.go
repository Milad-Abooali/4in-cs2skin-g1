package models

import "encoding/json"

type Payload struct {
	Payload json.RawMessage `json:"payload"`
}
type ReqRes struct {
	Type   string      `json:"type"`
	ReqID  int64       `json:"reqId,omitempty"`
	Status int         `json:"status"`          // 1 = success, 0 = error
	Error  int         `json:"error,omitempty"` // present only on error
	Data   interface{} `json:"data,omitempty"`  // present only on success
}
type Request struct {
	Type  string      `json:"type"`
	ReqID int64       `json:"reqId,omitempty"`
	Token string      `json:"token,omitempty"` // For Admin Side
	Data  interface{} `json:"data,omitempty"`  // present only on success
}

type HandlerError struct {
	Type string `json:"type"`
	Code int    `json:"code"`
	Data any    `json:"data,omitempty"` // present only on success
}

type HandlerOK struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}
