package utils

import (
	"encoding/json"
	"incognito-chat-api/types"
	"net/http"
)

func SendErrorMessage(w http.ResponseWriter, message any, status int) {
	response, _ := json.Marshal(message)
	w.WriteHeader(status)
	w.Write(response)
}

func SendSuccessMessage(w http.ResponseWriter, message any) {
	response, _ := json.Marshal(message)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func SocketResponse(connections []types.Connection, message types.WS_Signal) {
	response, _ := json.Marshal(message)
	for _, connection := range connections {
		if err := connection.Conn.WriteMessage(1, response); err != nil {
			continue
		}
	}
}